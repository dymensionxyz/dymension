package keeper

import (
	"time"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// iteratorAfterTime returns an iterator over all streams in the {prefix} space of state, that begin distributing rewards after a specific time.
func (k Keeper) iteratorAfterTime(ctx sdk.Context, prefix []byte, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(storetypes.InclusiveEndBytes(key), storetypes.PrefixEndBytes(prefix))
}

// iteratorBeforeTime returns an iterator over all streams in the {prefix} space of state, that begin distributing rewards before a specific time.
func (k Keeper) iteratorBeforeTime(ctx sdk.Context, prefix []byte, time time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	timeKey := getTimeKey(time)
	key := combineKeys(prefix, timeKey)
	return store.Iterator(prefix, storetypes.InclusiveEndBytes(key))
}

// iterator returns an iterator over all streams in the {prefix} space of state.
func (k Keeper) iterator(ctx sdk.Context, prefix []byte) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, prefix)
}

// UpcomingStreamsIteratorAfterTime returns the iterator to get all upcoming streams that start distribution after a specific time.
func (k Keeper) UpcomingStreamsIteratorAfterTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorAfterTime(ctx, types.KeyPrefixUpcomingStreams, time)
}

// UpcomingStreamsIteratorBeforeTime returns the iterator to get all upcoming streams that have already started distribution before a specific time.
func (k Keeper) UpcomingStreamsIteratorBeforeTime(ctx sdk.Context, time time.Time) sdk.Iterator {
	return k.iteratorBeforeTime(ctx, types.KeyPrefixUpcomingStreams, time)
}

// StreamsIterator returns the iterator for all streams.
func (k Keeper) StreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixStreams)
}

// UpcomingStreamsIterator returns the iterator for all upcoming streams.
func (k Keeper) UpcomingStreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixUpcomingStreams)
}

// ActiveStreamsIterator returns the iterator for all active streams.
func (k Keeper) ActiveStreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixActiveStreams)
}

// FinishedStreamsIterator returns the iterator for all finished streams.
func (k Keeper) FinishedStreamsIterator(ctx sdk.Context) sdk.Iterator {
	return k.iterator(ctx, types.KeyPrefixFinishedStreams)
}

// FilterLocksByMinDuration returns locks whose lock duration is greater than the provided minimum duration.
func FilterLocksByMinDuration(locks []lockuptypes.PeriodLock, minDuration time.Duration) []lockuptypes.PeriodLock {
	filteredLocks := make([]lockuptypes.PeriodLock, 0, len(locks)/2)
	for _, lock := range locks {
		if lock.Duration >= minDuration {
			filteredLocks = append(filteredLocks, lock)
		}
	}
	return filteredLocks
}
