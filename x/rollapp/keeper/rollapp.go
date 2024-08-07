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

func (k Keeper) CanUpdateRollapp(ctx sdk.Context, update *types.MsgUpdateRollappInformation) (types.Rollapp, error) {
	current, found := k.GetRollapp(ctx, update.RollappId)
	if !found {
		return current, errRollappNotFound
	}

	if update.Creator != current.Creator {
		return current, sdkerrors.ErrUnauthorized
	}

	if current.Frozen {
		return current, types.ErrRollappFrozen
	}

	// immutable values cannot be updated when the rollapp is sealed
	if update.UpdatingImmutableValues() && current.Sealed {
		return current, types.ErrImmutableFieldUpdateAfterSealed
	}

	if update.InitialSequencer != "" {
		current.InitialSequencer = update.InitialSequencer
	}

	if update.GenesisChecksum != "" {
		current.GenesisChecksum = update.GenesisChecksum
	}

	if update.Metadata != nil && !update.Metadata.IsEmpty() {
		current.Metadata = update.Metadata
	}

	if err := current.ValidateBasic(); err != nil {
		return current, fmt.Errorf("validate rollapp: %w", err)
	}

	return current, nil
}

// CheckIfRollappExists checks if a rollapp with the same ID, EIP155ID (if supported) already exists in the store.
// An exception is made for EIP155ID when the rollapp is frozen, in which case it is allowed to replace the existing rollapp.
func (k Keeper) CheckIfRollappExists(ctx sdk.Context, rollappId types.ChainID) error {
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return types.ErrRollappIDExists
	}

	if !rollappId.IsEIP155() {
		return nil
	}
	// check to see if the RollappId has been registered before with same key
	existingRollapp, isFound := k.GetRollappByEIP155(ctx, rollappId.GetEIP155ID())
	// allow replacing EIP155 only when forking (previous rollapp is frozen)
	if !isFound {
		return nil
	}
	if !existingRollapp.Frozen {
		return types.ErrRollappIDExists
	}
	existingRollappChainId, _ := types.NewChainID(existingRollapp.RollappId)

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

	// check if chain-id is EVM compatible. no err check as rollapp is already validated
	rollappID, _ := types.NewChainID(rollapp.RollappId)
	if !rollappID.IsEIP155() {
		return
	}

	// In case the chain id is EVM compatible, we store it by EIP155 id, to be retrievable by EIP155 id key
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

	if rollapp.GenesisChecksum == "" || rollapp.InitialSequencer == "" {
		return types.ErrSealWithImmutableFieldsNotSet
	}

	rollapp.Sealed = true
	k.SetRollapp(ctx, rollapp)

	return nil
}

// GetRollappByEIP155 returns a rollapp from its EIP155 id (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)  for EVM compatible rollapps
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

func (k Keeper) IsRollappSealed(ctx sdk.Context, rollappId string) bool {
	rollapp, found := k.GetRollapp(ctx, rollappId)
	return found && rollapp.Sealed
}
