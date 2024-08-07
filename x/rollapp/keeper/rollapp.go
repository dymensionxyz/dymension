package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) RegisterRollapp(ctx sdk.Context, rollapp types.Rollapp) error {
	// Already validated chain id in ValidateBasic, so we assume it's valid
	rollappId := types.MustNewChainID(rollapp.RollappId)
	if err := k.checkIfRollappExists(ctx, rollappId, rollapp.Alias); err != nil {
		return err
	}

	// Already validated chain id in ValidateBasic, so we assume it's valid
	creator := sdk.MustAccAddressFromBech32(rollapp.Owner)
	registrationFee := sdk.NewCoins(k.RegistrationFee(ctx))

	if !registrationFee.IsZero() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, registrationFee); err != nil {
			return errors.Join(types.ErrFeePayment, err)
		}

		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, registrationFee); err != nil {
			return fmt.Errorf("burn coins: %w", err)
		}
	}

	k.SetRollapp(ctx, rollapp)

	if err := k.hooks.RollappCreated(ctx, rollapp.RollappId); err != nil {
		return fmt.Errorf("rollapp created hook: %w", err)
	}

	return nil
}

func (k Keeper) UpdateRollapp(ctx sdk.Context, msg *types.MsgUpdateRollappInformation) error {
	if err := msg.ValidateBasic(); err != nil {
		return fmt.Errorf("validate update: %w", err)
	}

	updated, err := k.canUpdateRollapp(ctx, msg)
	if err != nil {
		return err
	}

	k.SetRollapp(ctx, updated)

	return nil
}

func (k Keeper) canUpdateRollapp(ctx sdk.Context, update *types.MsgUpdateRollappInformation) (types.Rollapp, error) {
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

	// immutable values cannot be updated when the rollapp is sealed
	if update.UpdatingImmutableValues() && current.Sealed {
		return current, types.ErrImmutableFieldUpdateAfterSealed
	}

	var err error
	current.Alias, err = k.canUpdateAlias(ctx, current.Alias, update.Alias)
	if err != nil {
		return current, err
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

	if err = current.ValidateBasic(); err != nil {
		return current, fmt.Errorf("validate rollapp: %w", err)
	}

	return current, nil
}

func (k Keeper) canUpdateAlias(
	ctx sdk.Context,
	currentAlias, updateAlias string,
) (string, error) {
	if updateAlias == "" || currentAlias == updateAlias {
		return currentAlias, nil
	}

	if _, isFound := k.GetRollappByAlias(ctx, updateAlias); isFound {
		return "", gerrc.ErrAlreadyExists
	}
	return updateAlias, nil
}

// checkIfRollappExists checks if a rollapp with the same ID or alias already exists in the store.
// An exception is made for when the rollapp is frozen, in which case it is allowed to replace the existing rollapp (forking).
func (k Keeper) checkIfRollappExists(ctx sdk.Context, rollappId types.ChainID, alias string) error {
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return types.ErrRollappExists
	}

	if _, isFound := k.GetRollappByAlias(ctx, alias); isFound {
		return types.ErrRollappAliasExists
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

	// save mapping for rollapp-by-alias
	store = prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByAliasPrefix))
	store.Set(types.RollappByAliasKey(
		rollapp.GetAlias(),
	), []byte(rollapp.RollappId))

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

	if rollapp.GenesisChecksum == "" || rollapp.Alias == "" || rollapp.InitialSequencer == "" {
		return types.ErrSealWithImmutableFieldsNotSet
	}

	rollapp.Sealed = true
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

func (k Keeper) GetRollappByAlias(ctx sdk.Context, alias string) (val types.Rollapp, ok bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappByAliasPrefix))
	id := store.Get(types.RollappByAliasKey(
		alias,
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

func (k Keeper) IsRollappSealed(ctx sdk.Context, rollappId string) bool {
	rollapp, found := k.GetRollapp(ctx, rollappId)
	return found && rollapp.Sealed
}
