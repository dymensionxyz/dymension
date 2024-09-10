package keeper

import (
	"fmt"
	"time"

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

	if update.GenesisInfo.NativeDenom.Base != "" {
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

func (k Keeper) SetRollappAsLaunched(ctx sdk.Context, rollapp *types.Rollapp) error {
	if !rollapp.AllImmutableFieldsAreSet() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "immutable fields not set")
	}

	rollapp.GenesisInfo.Sealed = true
	rollapp.Launched = true
	k.SetRollapp(ctx, *rollapp)

	return nil
}

// SealGenesisInfoWithLaunchTime seals the rollapp genesis info and sets it pre launch time according to the iro plan end time
// - GenesisInfo fields must be set
// - Rollapp must not be Launched
func (k Keeper) SealGenesisInfoWithLaunchTime(ctx sdk.Context, rollapp *types.Rollapp, preLaunchTime time.Time) error {
	if rollapp.Launched {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp already launched")
	}
	if !rollapp.GenesisInfoFieldsAreSet() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis info not set")
	}
	rollapp.GenesisInfo.Sealed = true
	rollapp.PreLaunchTime = preLaunchTime
	k.SetRollapp(ctx, *rollapp)
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
func (k Keeper) GetAllRollapps(ctx sdk.Context) (list []types.Rollapp) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// IsRollappStarted returns true if the rollapp is started
func (k Keeper) IsRollappStarted(ctx sdk.Context, rollappId string) bool {
	_, found := k.GetLatestStateInfoIndex(ctx, rollappId)
	return found
}
