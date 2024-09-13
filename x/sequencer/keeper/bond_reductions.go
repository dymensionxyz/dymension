package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

func (k Keeper) HandleBondReduction(ctx sdk.Context, currTime time.Time) {
	bondReductionIDs := k.GetMatureDecreasingBondIDs(ctx, currTime)
	for _, bondReductionID := range bondReductionIDs {
		wrapFn := func(ctx sdk.Context) error {
			return k.completeBondReduction(ctx, bondReductionID)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("reducing sequencer bond", "error", err, "bond reduction ID", bondReductionID)
			continue
		}
	}
}

func (k Keeper) completeBondReduction(ctx sdk.Context, bondReductionID uint64) error {
	reduction, found := k.GetBondReduction(ctx, bondReductionID)
	if !found {
		return errorsmod.Wrapf(
			types.ErrUnknownBondReduction,
			"bond reduction ID %d not found",
			bondReductionID,
		)
	}
	seq := k.MustGetSequencer(ctx, reduction.SequencerAddress)

	if !seq.Tokens.IsAllGTE(sdk.NewCoins(reduction.DecreaseBondAmount)) {
		return errorsmod.Wrapf(
			types.ErrInsufficientBond,
			"sequencer does not have enough bond to reduce insufficient bond: got %s, reducing by %s",
			seq.Tokens.String(),
			reduction.DecreaseBondAmount.String(),
		)
	}
	newBalance := seq.Tokens.Sub(reduction.DecreaseBondAmount)
	// in case between unbonding queue and now, the minbond value is increased,
	// handle it by only returning upto minBond amount and not all
	minBond := k.GetParams(ctx).MinBond
	if !newBalance.IsAllGTE(sdk.NewCoins(minBond)) {
		diff := minBond.SubAmount(newBalance.AmountOf(minBond.Denom))
		reduction.DecreaseBondAmount = reduction.DecreaseBondAmount.Sub(diff)
	}
	seqAddr := sdk.MustAccAddressFromBech32(reduction.SequencerAddress)
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAddr, sdk.NewCoins(reduction.DecreaseBondAmount))
	if err != nil {
		return err
	}

	seq.Tokens = seq.Tokens.Sub(reduction.DecreaseBondAmount)
	k.SetSequencer(ctx, seq)
	k.removeBondReduction(ctx, bondReductionID, reduction)

	return nil
}

// GetMatureDecreasingBondIDs returns all decreasing bond IDs for the given time
func (k Keeper) GetMatureDecreasingBondIDs(ctx sdk.Context, endTime time.Time) (bondReductionIDs []uint64) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.DecreasingBondQueueKey, sdk.PrefixEndBytes(types.DecreasingBondQueueByTimeKey(endTime)))
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		bondReductionID := sdk.BigEndianToUint64(iterator.Value())
		bondReductionIDs = append(bondReductionIDs, bondReductionID)
	}
	return
}

// SetDecreasingBondQueue sets the bond reduction item in the decreasing bond queue
func (k Keeper) SetDecreasingBondQueue(ctx sdk.Context, bondReduction types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	bondReductionID := k.increamentDecreasingBondID(ctx)
	b := k.cdc.MustMarshal(&bondReduction)
	store.Set(types.GetDecreasingBondQueueKey(bondReduction.SequencerAddress, bondReduction.DecreaseBondTime), sdk.Uint64ToBigEndian(bondReductionID))
	store.Set(append(types.GetDecreasingBondSequencerKey(bondReduction.SequencerAddress), sdk.Uint64ToBigEndian(bondReductionID)...), sdk.Uint64ToBigEndian(bondReductionID))
	store.Set(types.GetDecreasingBondIndexKey(bondReductionID), b)
}

// GetBondReduction returns the bond reduction item given bond reduction ID
func (k Keeper) GetBondReduction(ctx sdk.Context, bondReductionID uint64) (types.BondReduction, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDecreasingBondIndexKey(bondReductionID))
	if bz == nil {
		return types.BondReduction{}, false
	}
	var bd types.BondReduction
	k.cdc.MustUnmarshal(bz, &bd)
	return bd, true
}

func (k Keeper) GetAllBondReductions(ctx sdk.Context) (bds []types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DecreasingBondIndexKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		var bd types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &bd)
		bds = append(bds, bd)
	}
	return
}

// removeBondReduction removes the bond reduction item from the decreasing bond queue
func (k Keeper) removeBondReduction(ctx sdk.Context, bondReductionID uint64, bondReduction types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDecreasingBondQueueKey(bondReduction.SequencerAddress, bondReduction.DecreaseBondTime))
	store.Delete(append(types.GetDecreasingBondSequencerKey(bondReduction.SequencerAddress), sdk.Uint64ToBigEndian(bondReductionID)...))
	store.Delete(types.GetDecreasingBondIndexKey(bondReductionID))
}

// GetBondReductionsBySequencer returns the bond reduction item given sequencer address
func (k Keeper) GetBondReductionsBySequencer(ctx sdk.Context, sequencerAddr string) (bondReductions []types.BondReduction) {
	bondReductionIDs := k.getBondReductionIDsBySequencer(ctx, sequencerAddr)
	for _, bondReductionID := range bondReductionIDs {
		bd, found := k.GetBondReduction(ctx, bondReductionID)
		if found {
			bondReductions = append(bondReductions, bd)
		}
	}
	return
}

// getBondReductionIDsBySequencer returns the bond reduction item given sequencer address
func (k Keeper) getBondReductionIDsBySequencer(ctx sdk.Context, sequencerAddr string) (bondReductionIDs []uint64) {
	prefixKey := types.GetDecreasingBondSequencerKey(sequencerAddr)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		bondReductionID := sdk.BigEndianToUint64(iterator.Value())
		bondReductionIDs = append(bondReductionIDs, bondReductionID)
	}
	return
}

// increamentDecreasingBondID increments the decreasing bond ID anad returns the new ID
func (k Keeper) increamentDecreasingBondID(ctx sdk.Context) (decreasingBondID uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDecreasingBondIDKey())
	if bz != nil {
		decreasingBondID = sdk.BigEndianToUint64(bz)
	}
	decreasingBondID++

	bz = sdk.Uint64ToBigEndian(decreasingBondID)
	store.Set(types.GetDecreasingBondIDKey(), bz)
	return
}
