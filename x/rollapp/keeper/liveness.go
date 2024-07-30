package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func NextSlashOrJailHeight(
	hubBlockInterval time.Duration, // average time between hub blocks
	slashTimeNoUpdate time.Duration, // time until first slash if not updating
	slashInterval time.Duration, // gap between slash if still not updating
	jailTime time.Duration, // time until jail if not updating
	heightHub int64, // current hub height
	heightLastRollappUpdate int64, // when was the rollapp last updated
) (
	heightEvent int64, // hub height to schedule event
	isJail bool, // is it a jail event? (false -> slash)
) {
	// how long has the rollapp been down already?
	downTime := time.Duration(heightHub-heightLastRollappUpdate) * hubBlockInterval
	// when should we schedule the next slash/jail, in terms of down time duration?
	targetDuration := slashTimeNoUpdate + ((max(0, downTime-slashTimeNoUpdate)+slashInterval-1)/slashInterval)*slashInterval
	heightEvent = heightLastRollappUpdate + int64((targetDuration+hubBlockInterval-1)/hubBlockInterval)
	isJail = jailTime <= targetDuration
	return
}

// CheckLiveness will slash or jail any sequencers for whom their rollapp has been down
// and a slash or jail event is due
func (k Keeper) CheckLiveness(ctx sdk.Context) {
	h := ctx.BlockHeight()
	events := k.GetLivenessEvents(ctx, &h)
	for _, e := range events {
		if e.IsJail {
			err := k.sequencerKeeper.JailLiveness(ctx, e.RollappId)
			_ = err // TODO:
		} else {
			err := k.sequencerKeeper.SlashLiveness(ctx, e.RollappId)
			_ = err // TODO:
		}

		ra := k.MustGetRollapp(ctx, e.RollappId)
		k.DelLivenessEvents(ctx, ra.LivenessEventHeight, ra.RollappId)
		/*
			TODO: need to decide approach when rollapp does not have a sequencer, we can either
				a) not schedule the event
				b) schedule it but do the check when it occurs instead
				Leaning towards (b)
		*/
		k.ScheduleLivenessEvent(ctx, &ra)
	}
}

// IndicateLiveness will reschedule pending liveness events to a later date.
// Modifies the passed in rollapp object.
func (k Keeper) IndicateLiveness(ctx sdk.Context, ra *types.Rollapp) {
	ra.LastStateUpdateHeight = ctx.BlockHeight()
	k.DelLivenessEvents(ctx, ra.LivenessEventHeight, ra.RollappId)
	k.ScheduleLivenessEvent(ctx, ra)
}

// ScheduleLivenessEvent schedules a new liveness event. Assumes an event does not
// already exist for the rollapp. Assumes the rollapp has had at least one state update already.
// Modifies the passed in rollapp object.
func (k Keeper) ScheduleLivenessEvent(ctx sdk.Context, ra *types.Rollapp) {
	params := k.GetParams(ctx)
	nextH, isJail := NextSlashOrJailHeight(
		params.HubExpectedBlockTime,
		params.LivenessSlashTime,
		params.LivenessSlashInterval,
		params.LivenessJailTime,
		ctx.BlockHeight(),
		ra.LastStateUpdateHeight,
	)
	ra.LivenessEventHeight = nextH
	k.PutLivenessEvent(ctx, types.LivenessEvent{
		RollappId: ra.RollappId,
		HubHeight: nextH,
		IsJail:    isJail,
	})
}

// GetLivenessEvents returns events. If a height is specified, only for that height.
func (k Keeper) GetLivenessEvents(ctx sdk.Context, height *int64) []types.LivenessEvent {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.LivenessEventQueueIterKey(height))
	defer iterator.Close() // nolint: errcheck

	ret := []types.LivenessEvent{}
	for ; iterator.Valid(); iterator.Next() {
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
	for _, jail := range []bool{true, false} {
		k.DelLivenessEvent(ctx, types.LivenessEvent{
			RollappId: rollappID,
			HubHeight: height,
			IsJail:    jail,
		})
	}
}

// DelLivenessEvent deletes all liveness events for the rollapp from the queue
func (k Keeper) DelLivenessEvent(ctx sdk.Context, e types.LivenessEvent) {
	store := ctx.KVStore(k.storeKey)
	key := types.LivenessEventQueueKey(e)
	store.Delete(key)
}
