package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
)

var _ epochstypes.EpochHooks = epochHooks{}

type epochHooks struct {
	Keeper
}

func (k Keeper) GetEpochHooks() epochstypes.EpochHooks {
	return epochHooks{
		Keeper: k,
	}
}

// AfterEpochEnd is the epoch end hook.
// We want to clean up all the state info records that are older than the sequencer unbonding time.
func (e epochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) error {
	if epochIdentifier != e.StateInfoDeletionEpochIdentifier(ctx) {
		return nil
	}

	currentTimestamp := ctx.BlockTime()
	// for the time being, we can assume that the sequencer unbonding time will not change, therefore
	// we can assume that the number of resulting deletable state updates will remain constant
	seqUnbondingTime := e.sequencerKeeper.UnbondingTime(ctx)
	endTimestamp := currentTimestamp.Add(-seqUnbondingTime)

	e.DeleteStateInfoUntilTimestamp(ctx, endTimestamp)
	return nil
}

// BeforeEpochStart is the epoch start hook.
func (e epochHooks) BeforeEpochStart(sdk.Context, string, int64) error { return nil }
