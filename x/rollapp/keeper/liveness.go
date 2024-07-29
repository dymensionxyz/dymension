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
	evts := k.GetLivenessEvents(ctx)
	for _, evt := range evts {
		if !evt.IsJail {
			// TODO: slash
		} else {
			// TODO: jail
		}
		// TODO: check if jailed, if he is not then schedule another event
	}
}

// GetLivenessEvents returns all scheduled events for the current block height
func (k Keeper) GetLivenessEvents(ctx sdk.Context) []types.LivenessEvent {
	h := ctx.BlockHeight()
	return k.getLivenessEvents(ctx, &h)
}

// GetAllLivenessEvents returns all scheduled events (for genesis export)
func (k Keeper) GetAllLivenessEvents(ctx sdk.Context) []types.LivenessEvent {
	return k.getLivenessEvents(ctx, nil)
}

// getLivenessEvents returns events. If a height is specified, only for that height.
func (k Keeper) getLivenessEvents(ctx sdk.Context, height *int64) []types.LivenessEvent {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.LivenessEventQueueIterKey(height))
	defer iterator.Close() // nolint: errcheck

	ret := []types.LivenessEvent{}
	for ; iterator.Valid(); iterator.Next() {
		e := types.LivenessEventQueueItemToEvent(iterator.Key(), iterator.Value())
		if height != nil && *height < e.HubHeight {
			break
		}
		ret = append(ret, e)
	}
	return ret
}

// SetLivenessEvent returns all scheduled events (for genesis export)
func (k Keeper) SetLivenessEvent(ctx sdk.Context, e types.LivenessEvent) {
	store := ctx.KVStore(k.storeKey)
	key := types.LivenessEventQueueKey(e.HubHeight, e.RollappId)
	val := types.LivenessEventQueueSlash
	if e.IsJail {
		val = types.LivenessEventQueueJail
	}
	store.Set(key, val)
}
