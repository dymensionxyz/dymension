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
	unbondings := k.GetMatureDecreasingBondSequencers(ctx, currTime)
	for _, unbonding := range unbondings {
		wrapFn := func(ctx sdk.Context) error {
			return k.completeBondReduction(ctx, unbonding)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("reducing sequencer bond", "error", err, "sequencer", unbonding.SequencerAddress)
			continue
		}
	}
}

func (k Keeper) completeBondReduction(ctx sdk.Context, reduction types.BondReduction) error {
	seq, found := k.GetSequencer(ctx, reduction.SequencerAddress)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Tokens.AmountOf(reduction.DecreaseBondAmount.Denom).LT(reduction.DecreaseBondAmount.Amount) {
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
	if newBalance.AmountOf(minBond.Denom).LT(minBond.Amount) {
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
	k.removeDecreasingBondQueue(ctx, reduction)

	return nil
}

// GetMatureDecreasingBondSequencers returns all decreasing bond items for the given time
func (k Keeper) GetMatureDecreasingBondSequencers(ctx sdk.Context, endTime time.Time) (unbondings []types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.DecreasingBondQueueKey, sdk.PrefixEndBytes(types.DecreasingBondQueueByTimeKey(endTime)))
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		var b types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &b)
		unbondings = append(unbondings, b)
	}
	return
}

// SetDecreasingBondQueue sets the bond reduction item in the decreasing bond queue
func (k Keeper) SetDecreasingBondQueue(ctx sdk.Context, bondReduction types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&bondReduction)

	unbondingQueueKey := types.GetDecreasingBondQueueKey(bondReduction.SequencerAddress, bondReduction.DecreaseBondTime)
	store.Set(unbondingQueueKey, b)
}

// removeDecreasingBondQueue removes the bond reduction item from the decreasing bond queue
func (k Keeper) removeDecreasingBondQueue(ctx sdk.Context, bondReduction types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	unbondingQueueKey := types.GetDecreasingBondQueueKey(bondReduction.SequencerAddress, bondReduction.DecreaseBondTime)
	store.Delete(unbondingQueueKey)
}

// getSequencerDecreasingBonds returns the bond reduction item given sequencer address
func (k Keeper) getSequencerDecreasingBonds(ctx sdk.Context, sequencerAddr string) (bds []types.BondReduction) {
	prefixKey := types.DecreasingBondQueueKey
	store := prefix.NewStore(ctx.KVStore(k.storeKey), prefixKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var bd types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &bd)
		if bd.SequencerAddress == sequencerAddr {
			bds = append(bds, bd)
		}
	}

	return
}

func (k Keeper) GetAllBondReductions(ctx sdk.Context) (bds []types.BondReduction) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DecreasingBondQueueKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		var bd types.BondReduction
		k.cdc.MustUnmarshal(iterator.Value(), &bd)
		bds = append(bds, bd)
	}
	return
}
