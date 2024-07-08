package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) RegisterRollapp(ctx sdk.Context, rollapp types.Rollapp) error {
	if !k.RollappsEnabled(ctx) {
		return types.ErrRollappsDisabled
	}

	if err := rollapp.ValidateBasic(); err != nil {
		return fmt.Errorf("validate rollapp: %w", err)
	}

	err := k.checkIfRollappExists(ctx, rollapp.RollappId)
	if err != nil {
		return err
	}

	err = k.checkIfInitialSequencerAddressTaken(ctx, rollapp.InitialSequencerAddress)
	if err != nil {
		return fmt.Errorf("check if initial sequencer address taken: %w", err)
	}

	creator, _ := sdk.AccAddressFromBech32(rollapp.Creator)
	registrationFee := sdk.NewCoins(k.RegistrationFee(ctx))

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, registrationFee); err != nil {
		return errorsmod.Wrap(types.ErrFeePayment, err.Error())
	}

	if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, registrationFee); err != nil {
		return fmt.Errorf("burn coins: %w", err)
	}

	k.SetRollapp(ctx, rollapp)

	return nil
}

func (k Keeper) checkIfRollappExists(ctx sdk.Context, id string) error {
	rollappId, err := types.NewChainID(id)
	if err != nil {
		return err
	}
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return types.ErrRollappExists
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
		return types.ErrRollappExists
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

func (k Keeper) checkIfInitialSequencerAddressTaken(ctx sdk.Context, address string) error {
	if _, isFound := k.GetRollappByInitialSequencerAddress(ctx, address); isFound {
		return types.ErrInitialSequencerAddressTaken
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
	), b)
}

// GetRollappByEIP155 returns a rollapp from its EIP155 id (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)  for EVM compatible rollapps
func (k Keeper) GetRollappByEIP155(
	ctx sdk.Context,
	eip155 uint64,
) (val types.Rollapp, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByEIP155KeyPrefix))

	b := store.Get(types.RollappByEIP155Key(
		eip155,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) GetRollappByInitialSequencerAddress(ctx sdk.Context, address string) (types.Rollapp, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Rollapp
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.InitialSequencerAddress == address {
			return val, true
		}
	}
	return types.Rollapp{}, false
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
