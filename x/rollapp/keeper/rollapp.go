package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) CheckAndUpdateRollappFields(ctx sdk.Context, update *types.MsgUpdateRollappInformation) (types.Rollapp, error) {
	current, found := k.GetRollapp(ctx, update.RollappId)
	if !found {
		return current, errRollappNotFound
	}

	if update.Owner != current.Owner {
		return current, sdkerrors.ErrUnauthorized
	}

	if current.Frozen {
		return current, types.ErrRollappFrozen
	}

	// immutable values cannot be updated when the rollapp is launched
	if update.UpdatingImmutableValues() && current.Launched {
		return current, types.ErrImmutableFieldUpdateAfterLaunched
	}

	if update.UpdatingGenesisInfo() && current.GenesisInfo.Sealed {
		return current, types.ErrGenesisInfoSealed
	}

	if update.InitialSequencer != "" {
		current.InitialSequencer = update.InitialSequencer
	}

	if update.GenesisInfo.GenesisChecksum != "" {
		current.GenesisInfo.GenesisChecksum = update.GenesisInfo.GenesisChecksum
	}

	if update.GenesisInfo.Bech32Prefix != "" {
		current.GenesisInfo.Bech32Prefix = update.GenesisInfo.Bech32Prefix
	}

	if update.GenesisInfo.NativeDenom != nil {
		current.GenesisInfo.NativeDenom = update.GenesisInfo.NativeDenom
	}

	if !update.GenesisInfo.InitialSupply.IsNil() {
		current.GenesisInfo.InitialSupply = update.GenesisInfo.InitialSupply
	}

	if update.Metadata != nil && !update.Metadata.IsEmpty() {
		current.Metadata = update.Metadata
	}

	if err := current.ValidateBasic(); err != nil {
		return current, fmt.Errorf("validate rollapp: %w", err)
	}

	return current, nil
}

// CheckIfRollappExists checks if a rollapp with the same ID or alias already exists in the store.
// An exception is made for when the rollapp is frozen, in which case it is allowed to replace the existing rollapp (forking).
func (k Keeper) CheckIfRollappExists(ctx sdk.Context, rollappId types.ChainID) error {
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return types.ErrRollappExists
	}

	existingRollapp, isFound := k.GetRollappByEIP155(ctx, rollappId.GetEIP155ID())
	// allow replacing EIP155 only when forking (previous rollapp is frozen)
	if !isFound {
		// if not forking, check to see if the Rollapp has been registered before with same name
		if _, isFound = k.GetRollappByName(ctx, rollappId.GetName()); isFound {
			return types.ErrRollappExists
		}
		return nil
	}
	if !existingRollapp.Frozen {
		return types.ErrRollappExists
	}
	existingRollappChainId := types.MustNewChainID(existingRollapp.RollappId)

	if rollappId.GetName() != existingRollappChainId.GetName() {
		return errorsmod.Wrapf(types.ErrInvalidRollappID, "rollapp name should be %s", existingRollappChainId.GetName())
	}

	nextRevisionNumber := existingRollappChainId.GetRevisionNumber() + 1
	if rollappId.GetRevisionNumber() != nextRevisionNumber {
		return errorsmod.Wrapf(types.ErrInvalidRollappID, "revision number should be %d", nextRevisionNumber)
	}
	return nil
}

// SetRollapp set a specific rollapp in the store from its index
func (k Keeper) SetRollapp(ctx sdk.Context, rollapp types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	b := k.cdc.MustMarshal(&rollapp)
	store.Set(types.RollappKey(
		rollapp.RollappId,
	), b)

	// no err check as rollapp is already validated
	rollappID := types.MustNewChainID(rollapp.RollappId)

	store = prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))
	store.Set(types.RollappByEIP155Key(
		rollappID.GetEIP155ID(),
	), []byte(rollapp.RollappId))
}

func (k Keeper) SealRollapp(ctx sdk.Context, rollappId string) error {
	rollapp, found := k.GetRollapp(ctx, rollappId)
	if !found {
		return gerrc.ErrNotFound
	}

	if rollapp.Launched {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp already launched")
	}

	if !rollapp.AllImmutableFieldsAreSet() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "seal with immutable fields not set")
	}

	rollapp.Launched = true
	k.SetRollapp(ctx, rollapp)

	return nil
}

func (k Keeper) SealRollappGenesisInfo(ctx sdk.Context, rollappId string) error {
	rollapp, found := k.GetRollapp(ctx, rollappId)
	if !found {
		return gerrc.ErrNotFound
	}

	if rollapp.GenesisInfo.Sealed {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis info already sealed")
	}

	if !rollapp.GenesisInfoFieldsAreSet() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis info fields not set")
	}

	rollapp.GenesisInfo.Sealed = true
	k.SetRollapp(ctx, rollapp)

	return nil
}

// GetRollappByEIP155 returns a rollapp from its EIP155 id (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)
func (k Keeper) GetRollappByEIP155(ctx sdk.Context, eip155 uint64) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))
	id := store.Get(types.RollappByEIP155Key(
		eip155,
	))
	if id == nil {
		return val, false
	}

	return k.GetRollapp(ctx, string(id))
}

// GetRollapp returns a rollapp from its chain name
func (k Keeper) GetRollapp(
	ctx sdk.Context,
	rollappId string,
) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))

	b := store.Get(types.RollappKey(
		rollappId,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetRollappByName(
	ctx sdk.Context,
	name string,
) (val types.Rollapp, found bool) {
	name = name + "_"
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte(name))

	defer iterator.Close() // nolint: errcheck

	if !iterator.Valid() {
		return val, false
	}

	k.cdc.MustUnmarshal(iterator.Value(), &val)
	return val, true
}

func (k Keeper) MustGetRollapp(ctx sdk.Context, rollappId string) types.Rollapp {
	ret, found := k.GetRollapp(ctx, rollappId)
	if !found {
		panic(fmt.Sprintf("rollapp not found: id: %s", rollappId))
	}
	return ret
}

// RemoveRollapp removes a rollapp from the store using rollapp name
func (k Keeper) RemoveRollapp(
	ctx sdk.Context,
	rollappId string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	store.Delete(types.RollappKey(
		rollappId,
	))
}

// GetAllRollapps returns all rollapp
func (k Keeper) GetAllRollapps(ctx sdk.Context) []types.Rollapp {
	return k.FilterRollapps(ctx, func(rollapp types.Rollapp) bool { return true })
}

// IsRollappStarted returns true if the rollapp is started
func (k Keeper) IsRollappStarted(ctx sdk.Context, rollappId string) bool {
	_, found := k.GetLatestStateInfoIndex(ctx, rollappId)
	return found
}

func (k Keeper) MarkRollappAsVulnerable(ctx sdk.Context, rollappId string) error {
	return k.FreezeRollapp(ctx, rollappId)
}

// FreezeRollapp marks the rollapp as frozen and reverts all pending states.
// NB! This method is going to be changed as soon as the "Freezing" ADR is ready.
func (k Keeper) FreezeRollapp(ctx sdk.Context, rollappID string) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound
	}

	rollapp.Frozen = true

	k.RevertPendingStates(ctx, rollappID)

	if rollapp.ChannelId != "" {
		clientID, _, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", rollapp.ChannelId)
		if err != nil {
			return fmt.Errorf("get channel client state: %w", err)
		}

		err = k.freezeClientState(ctx, clientID)
		if err != nil {
			return fmt.Errorf("freeze client state: %w", err)
		}
	}

	k.SetRollapp(ctx, rollapp)
	return nil
}

// verifyClientID verifies that the provided clientID is the same clientID used by the provided rollapp.
// Possible scenarios:
//  1. both channelID and clientID are empty -> okay
//  2. channelID is empty while clientID is not -> error: rollapp does not have a channel
//  3. clientID is empty while channelID is not -> error: rollapp does have a channel, but the provided clientID is empty
//  4. both channelID and clientID are not empty -> okay: compare the provided channelID against the one from IBC
func (k Keeper) verifyClientID(ctx sdk.Context, rollappID, clientID string) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound
	}

	var (
		emptyRollappChannelID = rollapp.ChannelId == ""
		emptyClientID         = clientID == ""
	)

	switch {
	// both channelID and clientID are empty
	case emptyRollappChannelID && emptyClientID:
		return nil // everything is fine, expected scenario

	// channelID is empty while clientID is not
	case emptyRollappChannelID:
		return fmt.Errorf("rollapp does not have a channel: rollapp '%s'", rollappID)

	// clientID is empty while channelID is not
	case emptyClientID:
		return fmt.Errorf("empty clientID while the rollapp channelID is not empty")

	// both channelID and clientID are not empty
	default:
		// extract rollapp channelID
		extractedClientId, _, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", rollapp.ChannelId)
		if err != nil {
			return fmt.Errorf("get channel client state: %w", err)
		}
		// compare it with the passed clientID
		if extractedClientId != clientID {
			return fmt.Errorf("clientID does not match the one in the rollapp: clientID %s, rollapp clientID %s", clientID, extractedClientId)
		}
		return nil
	}
}

func (k Keeper) FilterRollapps(ctx sdk.Context, f func(types.Rollapp) bool) []types.Rollapp {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	var result []types.Rollapp
	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if f(val) {
			result = append(result, val)
		}
	}
	return result
}

func FilterNonVulnerable(b types.Rollapp) bool {
	return !b.IsVulnerable()
}

func (k Keeper) IsDRSVersionVulnerable(ctx sdk.Context, version string) (bool, error) {
	return k.vulnerableDRSVersions.Has(ctx, version)
}

func (k Keeper) SetVulnerableDRSVersion(ctx sdk.Context, version string) error {
	return k.vulnerableDRSVersions.Set(ctx, version)
}

func (k Keeper) GetAllVulnerableDRSVersions(ctx sdk.Context) ([]string, error) {
	iter, err := k.vulnerableDRSVersions.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	return iter.Keys()
}
