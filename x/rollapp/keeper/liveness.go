package keeper

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

/*
Slash rollapps based on liveness requirements (time since last update (actually number of hub blocks)).
See https://www.notion.so/dymension/sequencer-jailing-slashing-3455fe70923143cbbfd8f96d71deb583?pvs=4#11fa4a51f86a80ac9c2cda8ab7f8d39a
*/

// NextSlashHeight returns the next height on the HUB to slash or jail the rollapp
// It will respect all parameters passed in.
// Assumes that if current hub height is already a slash height, then to schedule for the next one.
// CONTRACT: heightHub >= heightLastRollappUpdate
func NextSlashHeight(
	blocksSlashNoUpdate uint64, // time until first slash if not updating
	blocksSlashInterval uint64, // gap between slash if still not updating
	heightHub int64, // current height on the hub
	heightLastRollappUpdate int64, // when was the rollapp last updated
) (
	heightEvent int64, // hub height to schedule event
) {
	// how long has the rollapp been down ?
	down := uint64(heightHub - heightLastRollappUpdate) //nolint:gosec
	// when should we schedule the next slash, in terms of downtime duration?
	interval := blocksSlashNoUpdate
	if blocksSlashNoUpdate <= down {
		// round up to next slash interval
		interval += ((down-blocksSlashNoUpdate)/blocksSlashInterval + 1) * blocksSlashInterval
	}
	heightEvent = heightLastRollappUpdate + int64(interval) //nolint:gosec
	return
}

// CheckLiveness will slash or jail any sequencers for whom their rollapp has been down
// and a slash or jail event is due. Run in end block.
func (k Keeper) CheckLiveness(ctx sdk.Context) {
	h := ctx.BlockHeight()
	events := k.GetLivenessEvents(ctx, &h)
	for _, e := range events {
		err := osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
			return k.HandleLivenessEvent(ctx, e)
		})
		if err != nil {
			// We intentionally do not reschedule the event. It's not MVP. Also, if it failed once, why would it succeed second time?
			k.Logger(ctx).Error(
				"Check liveness event",
				"event", e,
				"err", err,
			)
		}
	}
}

// HandleLivenessEvent will slash or jail and then schedule a new event in the future.
func (k Keeper) HandleLivenessEvent(ctx sdk.Context, e types.LivenessEvent) error {
	err := k.SequencerK.SlashLiveness(ctx, e.RollappId)
	if err != nil {
		return errorsmod.Wrap(err, "slash liveness")
	}

	ra := k.MustGetRollapp(ctx, e.RollappId)
	k.DelLivenessEvents(ctx, e.HubHeight, e.RollappId)
	k.ScheduleLivenessEvent(ctx, &ra)
	k.SetRollapp(ctx, ra)
	return nil
}

func (k Keeper) IndicateLiveness(ctx sdk.Context, ra *types.Rollapp) {
	k.ResetLivenessClock(ctx, ra)
	k.ScheduleLivenessEvent(ctx, ra)
}

// ResetLivenessClock will reschedule pending liveness events to a later block height.
// Modifies the passed-in rollapp object.
func (k Keeper) ResetLivenessClock(ctx sdk.Context, ra *types.Rollapp) {
	k.DelLivenessEvents(ctx, ra.LivenessEventHeight, ra.RollappId)
	ra.LivenessEventHeight = 0
	ra.LivenessCountdownStartHeight = ctx.BlockHeight()
}

// ScheduleLivenessEvent schedules a new liveness event. Assumes an event does not
// already exist for the rollapp. Modifies the passed-in rollapp object.
func (k Keeper) ScheduleLivenessEvent(ctx sdk.Context, ra *types.Rollapp) {
	nextH := NextSlashHeight(
		k.LivenessSlashBlocks(ctx),
		k.LivenessSlashInterval(ctx),
		ctx.BlockHeight(),
		ra.LivenessCountdownStartHeight,
	)
	ra.LivenessEventHeight = nextH
	k.PutLivenessEvent(ctx, types.LivenessEvent{
		RollappId: ra.RollappId,
		HubHeight: nextH,
	})
}

// GetLivenessEvents returns events. If a height is specified, only for that height.
func (k Keeper) GetLivenessEvents(ctx sdk.Context, height *int64) []types.LivenessEvent {
	store := ctx.KVStore(k.storeKey)
	key := types.LivenessEventQueueKeyPrefix
	if height != nil {
		key = types.LivenessEventQueueIterHeightKey(*height)
	}
	iterator := storetypes.KVStorePrefixIterator(store, key)
	defer iterator.Close() // nolint: errcheck

	ret := []types.LivenessEvent{}
	for ; iterator.Valid(); iterator.Next() {
		// events are stored in height non-decreasing order
		e := types.LivenessEventQueueKeyToEvent(iterator.Key())
		if height != nil && *height < e.HubHeight {
			break
		}
		ret = append(ret, e)
	}
	return ret
}

// PutLivenessEvent puts a new event in the queue
func (k Keeper) PutLivenessEvent(ctx sdk.Context, e types.LivenessEvent) {
	store := ctx.KVStore(k.storeKey)
	key := types.LivenessEventQueueKey(e)
	store.Set(key, []byte{})
}

// DelLivenessEvents deletes all liveness events for the rollapp from the queue
func (k Keeper) DelLivenessEvents(ctx sdk.Context, height int64, rollappID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.LivenessEventQueueKey(types.LivenessEvent{
		RollappId: rollappID,
		HubHeight: height,
	})
	store.Delete(key)
}
