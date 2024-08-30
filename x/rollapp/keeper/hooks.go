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
// We want to clean up the demand orders that are with underlying packet status which are finalized.
func (e epochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, _ int64) error {
	if epochIdentifier != e.EpochIdentifier(ctx) {
		return nil
	}

	currentTimestamp := ctx.BlockTime()
	seqUnbondingTime := e.sequencerKeeper.UnbondingTime(ctx)
	endTimestamp := currentTimestamp.Add(-seqUnbondingTime)

	e.DeleteStateInfoUntilTimestamp(ctx, endTimestamp)
	return nil
}

// BeforeEpochStart is the epoch start hook.
func (e epochHooks) BeforeEpochStart(sdk.Context, string, int64) error {
	return nil
}
