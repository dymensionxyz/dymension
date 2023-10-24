package keeper

import (
	"fmt"
	"time"

	db "github.com/tendermint/tm-db"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getDistributedCoinsFromStreams returns coins that have been distributed already from the provided streams
func (k Keeper) getDistributedCoinsFromStreams(streams []types.Stream) sdk.Coins {
	coins := sdk.Coins{}
	for _, stream := range streams {
		coins = coins.Add(stream.DistributedCoins...)
	}
	return coins
}

// getToDistributeCoinsFromStreams returns coins that have not been distributed yet from the provided streams
func (k Keeper) getToDistributeCoinsFromStreams(streams []types.Stream) sdk.Coins {
	coins := sdk.Coins{}
	distributed := sdk.Coins{}

	for _, stream := range streams {
		coins = coins.Add(stream.Coins...)
		distributed = distributed.Add(stream.DistributedCoins...)
	}
	return coins.Sub(distributed...)
}

// getToDistributeCoinsFromIterator utilizes iterator to return a list of streams.
// From these streams, coins that have not yet been distributed are returned
func (k Keeper) getToDistributeCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getToDistributeCoinsFromStreams(k.getStreamsFromIterator(ctx, iterator))
}

// getDistributedCoinsFromIterator utilizes iterator to return a list of streams.
// From these streams, coins that have already been distributed are returned
func (k Keeper) getDistributedCoinsFromIterator(ctx sdk.Context, iterator db.Iterator) sdk.Coins {
	return k.getDistributedCoinsFromStreams(k.getStreamsFromIterator(ctx, iterator))
}

// moveUpcomingStreamToActiveStream moves a stream that has reached it's start time from an upcoming to an active status.
func (k Keeper) moveUpcomingStreamToActiveStream(ctx sdk.Context, stream types.Stream) error {
	// validation for current time and distribution start time
	if ctx.BlockTime().Before(stream.StartTime) {
		return fmt.Errorf("stream is not able to start distribution yet: %s >= %s", ctx.BlockTime().String(), stream.StartTime.String())
	}

	timeKey := getTimeKey(stream.StartTime)
	if err := k.deleteStreamRefByKey(ctx, combineKeys(types.KeyPrefixUpcomingStreams, timeKey), stream.Id); err != nil {
		return err
	}
	if err := k.addStreamRefByKey(ctx, combineKeys(types.KeyPrefixActiveStreams, timeKey), stream.Id); err != nil {
		return err
	}
	return nil
}

// moveActiveStreamToFinishedStream moves a stream that has completed its distribution from an active to a finished status.
func (k Keeper) moveActiveStreamToFinishedStream(ctx sdk.Context, stream types.Stream) error {
	timeKey := getTimeKey(stream.StartTime)
	if err := k.deleteStreamRefByKey(ctx, combineKeys(types.KeyPrefixActiveStreams, timeKey), stream.Id); err != nil {
		return err
	}
	if err := k.addStreamRefByKey(ctx, combineKeys(types.KeyPrefixFinishedStreams, timeKey), stream.Id); err != nil {
		return err
	}
	if err := k.deleteStreamIDForDenom(ctx, stream.Id, stream.DistributeTo.Denom); err != nil {
		return err
	}
	k.hooks.AfterFinishDistribution(ctx, stream.Id)
	return nil
}

// getLocksToDistributionWithMaxDuration returns locks that match the provided lockuptypes QueryCondition,
// are greater than the provided minDuration, AND have yet to be distributed to.
func (k Keeper) getLocksToDistributionWithMaxDuration(ctx sdk.Context, distrTo lockuptypes.QueryCondition, minDuration time.Duration) []lockuptypes.PeriodLock {
	switch distrTo.LockQueryType {
	case lockuptypes.ByDuration:
		denom := lockuptypes.NativeDenom(distrTo.Denom)
		if distrTo.Duration > minDuration {
			return k.lk.GetLocksLongerThanDurationDenom(ctx, denom, minDuration)
		}
		return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, distrTo.Duration)
	case lockuptypes.ByTime:
		panic("Stream by time is present, however is no longer supported. This should have been blocked in ValidateBasic")
	default:
	}
	return []lockuptypes.PeriodLock{}
}

// FilteredLocksDistributionEst estimates distribution amount of coins from stream.
// It also applies an update for the stream, handling the sending of the rewards.
// (Note this update is in-memory, it does not change state.)
func (k Keeper) FilteredLocksDistributionEst(ctx sdk.Context, stream types.Stream, filteredLocks []lockuptypes.PeriodLock) (types.Stream, sdk.Coins, bool, error) {
	TotalAmtLocked := k.lk.GetPeriodLocksAccumulation(ctx, stream.DistributeTo)
	if TotalAmtLocked.IsZero() {
		return types.Stream{}, nil, false, nil
	}
	if TotalAmtLocked.IsNegative() {
		return types.Stream{}, nil, true, nil
	}

	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	// remainEpochs is the number of remaining epochs that the stream will pay out its rewards.
	// for a perpetual stream, it will pay out everything in the next epoch, and we don't make
	// an assumption of the rate at which it will get refilled at.
	remainEpochs := uint64(1)
	if !stream.IsPerpetual {
		remainEpochs = stream.NumEpochsPaidOver - stream.FilledEpochs
	}
	if remainEpochs == 0 {
		return stream, sdk.Coins{}, false, nil
	}

	remainCoinsPerEpoch := sdk.Coins{}
	for _, coin := range remainCoins {
		// distribution amount per epoch = stream_size / (remain_epochs)
		amt := coin.Amount.QuoRaw(int64(remainEpochs))
		remainCoinsPerEpoch = remainCoinsPerEpoch.Add(sdk.NewCoin(coin.Denom, amt))
	}

	// now we compute the filtered coins
	filteredDistrCoins := sdk.Coins{}
	if len(filteredLocks) == 0 {
		// if were doing no filtering, we want to calculate the total amount to distributed in
		// the next epoch.
		// distribution in next epoch = stream_size  / (remain_epochs)
		filteredDistrCoins = remainCoinsPerEpoch
	}
	for _, lock := range filteredLocks {
		denomLockAmt := lock.Coins.AmountOf(stream.DistributeTo.Denom)

		for _, coin := range remainCoinsPerEpoch {
			// distribution amount = stream_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			// distribution amount = stream_size_per_epoch * denom_lock_amount / total_denom_lock_amount
			amt := coin.Amount.Mul(denomLockAmt).Quo(TotalAmtLocked)
			filteredDistrCoins = filteredDistrCoins.Add(sdk.NewCoin(coin.Denom, amt))
		}
	}

	// increase filled epochs after distribution
	stream.FilledEpochs += 1
	stream.DistributedCoins = stream.DistributedCoins.Add(remainCoinsPerEpoch...)

	return stream, filteredDistrCoins, false, nil
}

// distributionInfo stores all of the information for pent up sends for rewards distributions.
// This enables us to lower the number of events and calls to back.
type distributionInfo struct {
	nextID            int
	lockOwnerAddrToID map[string]int
	idToBech32Addr    []string
	idToDecodedAddr   []sdk.AccAddress
	idToDistrCoins    []sdk.Coins
}

// newDistributionInfo creates a new distributionInfo struct
func newDistributionInfo() distributionInfo {
	return distributionInfo{
		nextID:            0,
		lockOwnerAddrToID: make(map[string]int),
		idToBech32Addr:    []string{},
		idToDecodedAddr:   []sdk.AccAddress{},
		idToDistrCoins:    []sdk.Coins{},
	}
}

// addLockRewards adds the provided rewards to the lockID mapped to the provided owner address.
func (d *distributionInfo) addLockRewards(owner string, rewards sdk.Coins) error {
	if id, ok := d.lockOwnerAddrToID[owner]; ok {
		oldDistrCoins := d.idToDistrCoins[id]
		d.idToDistrCoins[id] = rewards.Add(oldDistrCoins...)
	} else {
		id := d.nextID
		d.nextID += 1
		d.lockOwnerAddrToID[owner] = id
		decodedOwnerAddr, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return err
		}
		d.idToBech32Addr = append(d.idToBech32Addr, owner)
		d.idToDecodedAddr = append(d.idToDecodedAddr, decodedOwnerAddr)
		d.idToDistrCoins = append(d.idToDistrCoins, rewards)
	}
	return nil
}

// doDistributionSends utilizes provided distributionInfo to send coins from the module account to various recipients.
func (k Keeper) doDistributionSends(ctx sdk.Context, distrs *distributionInfo) error {
	numIDs := len(distrs.idToDecodedAddr)
	if len(distrs.idToDistrCoins) != numIDs {
		return fmt.Errorf("number of addresses and coins to distribute to must be equal")
	}
	ctx.Logger().Debug(fmt.Sprintf("Beginning distribution to %d users", numIDs))

	for id := 0; id < numIDs; id++ {
		err := k.bk.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			distrs.idToDecodedAddr[id],
			distrs.idToDistrCoins[id])

		if err != nil {
			return err
		}
	}
	ctx.Logger().Debug("Finished sending, now creating liquidity add events")
	for id := 0; id < numIDs; id++ {
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.TypeEvtDistribution,
				sdk.NewAttribute(types.AttributeReceiver, distrs.idToBech32Addr[id]),
				sdk.NewAttribute(types.AttributeAmount, distrs.idToDistrCoins[id].String()),
			),
		})
	}
	ctx.Logger().Debug(fmt.Sprintf("Finished Distributing to %d users", numIDs))
	return nil
}

// distributeSyntheticInternal runs the distribution logic for a synthetic rewards distribution stream, and adds the sends to
// the distrInfo struct. It also updates the stream for the distribution.
// locks is expected to be the correct set of lock recipients for this stream.
func (k Keeper) distributeSyntheticInternal(
	ctx sdk.Context, stream types.Stream, locks []lockuptypes.PeriodLock, distrInfo *distributionInfo,
) (sdk.Coins, error) {
	qualifiedLocks := k.lk.GetLocksLongerThanDurationDenom(ctx, stream.DistributeTo.Denom, stream.DistributeTo.Duration)

	// map from lockID to present index in resultant list
	// to be state compatible with what we had before, we iterate over locks, to get qualified locks
	// to be in the same order as what is present in locks.
	// in a future release, we can just use qualified locks directly.
	type lockIndexPair struct {
		lock  lockuptypes.PeriodLock
		index int
	}
	qualifiedLocksMap := make(map[uint64]lockIndexPair, len(qualifiedLocks))
	for _, lock := range qualifiedLocks {
		qualifiedLocksMap[lock.ID] = lockIndexPair{lock, -1}
	}
	curIndex := 0
	for _, lock := range locks {
		if v, ok := qualifiedLocksMap[lock.ID]; ok {
			qualifiedLocksMap[lock.ID] = lockIndexPair{v.lock, curIndex}
			curIndex += 1
		}
	}

	sortedAndTrimmedQualifiedLocks := make([]lockuptypes.PeriodLock, curIndex)
	for _, v := range qualifiedLocksMap {
		if v.index < 0 {
			continue
		}
		sortedAndTrimmedQualifiedLocks[v.index] = v.lock
	}

	return k.distributeInternal(ctx, stream, sortedAndTrimmedQualifiedLocks, distrInfo)
}

// distributeInternal runs the distribution logic for a stream, and adds the sends to
// the distrInfo struct. It also updates the stream for the distribution.
// Locks is expected to be the correct set of lock recipients for this stream.
func (k Keeper) distributeInternal(
	ctx sdk.Context, stream types.Stream, locks []lockuptypes.PeriodLock, distrInfo *distributionInfo,
) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
	denom := lockuptypes.NativeDenom(stream.DistributeTo.Denom)
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	if lockSum.IsZero() {
		return nil, nil
	}

	remainCoins := stream.Coins.Sub(stream.DistributedCoins...)
	// if its a perpetual stream, we set remaining epochs to 1.
	// otherwise is is a non perpetual stream and we determine how many epoch payouts are left
	remainEpochs := uint64(1)
	if !stream.IsPerpetual {
		remainEpochs = stream.NumEpochsPaidOver - stream.FilledEpochs
	}

	for _, lock := range locks {
		distrCoins := sdk.Coins{}
		for _, coin := range remainCoins {
			// distribution amount = stream_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			denomLockAmt := lock.Coins.AmountOfNoDenomValidation(denom)
			amt := coin.Amount.Mul(denomLockAmt).Quo(lockSum.Mul(sdk.NewInt(int64(remainEpochs))))
			if amt.IsPositive() {
				newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: amt}
				distrCoins = distrCoins.Add(newlyDistributedCoin)
			}
		}
		distrCoins = distrCoins.Sort()
		if distrCoins.Empty() {
			continue
		}
		// update the amount for that address
		err := distrInfo.addLockRewards(lock.Owner, distrCoins)
		if err != nil {
			return nil, err
		}

		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	err := k.updateStreamPostDistribute(ctx, stream, totalDistrCoins)
	return totalDistrCoins, err
}

// updateStreamPostDistribute increments the stream's filled epochs field.
// Also adds the coins that were just distributed to the stream's distributed coins field.
func (k Keeper) updateStreamPostDistribute(ctx sdk.Context, stream types.Stream, newlyDistributedCoins sdk.Coins) error {
	stream.FilledEpochs += 1
	stream.DistributedCoins = stream.DistributedCoins.Add(newlyDistributedCoins...)
	if err := k.setStream(ctx, &stream); err != nil {
		return err
	}
	return nil
}

// getDistributeToBaseLocks takes a stream along with cached period locks by denom and returns locks that must be distributed to
func (k Keeper) getDistributeToBaseLocks(ctx sdk.Context, stream types.Stream, cache map[string][]lockuptypes.PeriodLock) []lockuptypes.PeriodLock {
	// if stream is empty, don't get the locks
	if stream.Coins.Empty() {
		return []lockuptypes.PeriodLock{}
	}
	// Confusingly, there is no way to get all synthetic lockups. Thus we use a separate method `distributeSyntheticInternal` to separately get lockSum for synthetic lockups.
	// All streams have a precondition of being ByDuration.
	distributeBaseDenom := lockuptypes.NativeDenom(stream.DistributeTo.Denom)
	if _, ok := cache[distributeBaseDenom]; !ok {
		cache[distributeBaseDenom] = k.getLocksToDistributionWithMaxDuration(
			ctx, stream.DistributeTo, time.Millisecond)
	}
	// get this from memory instead of hitting iterators / underlying stores.
	// due to many details of cacheKVStore, iteration will still cause expensive IAVL reads.
	allLocks := cache[distributeBaseDenom]
	return FilterLocksByMinDuration(allLocks, stream.DistributeTo.Duration)
}

// Distribute distributes coins from an array of streams to all eligible locks.
func (k Keeper) Distribute(ctx sdk.Context, streams []types.Stream) (sdk.Coins, error) {
	distrInfo := newDistributionInfo()

	locksByDenomCache := make(map[string][]lockuptypes.PeriodLock)
	totalDistributedCoins := sdk.Coins{}
	for _, stream := range streams {
		filteredLocks := k.getDistributeToBaseLocks(ctx, stream, locksByDenomCache)
		// send based on synthetic lockup coins if it's distributing to synthetic lockups
		var streamDistributedCoins sdk.Coins
		var err error
		if lockuptypes.IsSyntheticDenom(stream.DistributeTo.Denom) {
			streamDistributedCoins, err = k.distributeSyntheticInternal(ctx, stream, filteredLocks, &distrInfo)
		} else {
			streamDistributedCoins, err = k.distributeInternal(ctx, stream, filteredLocks, &distrInfo)
		}
		if err != nil {
			return nil, err
		}
		totalDistributedCoins = totalDistributedCoins.Add(streamDistributedCoins...)
	}

	err := k.doDistributionSends(ctx, &distrInfo)
	if err != nil {
		return nil, err
	}
	k.hooks.AfterEpochDistribution(ctx)

	k.checkFinishDistribution(ctx, streams)
	return totalDistributedCoins, nil
}

// checkFinishDistribution checks if all non perpetual streams provided have completed their required distributions.
// If complete, move the stream from an active to a finished status.
func (k Keeper) checkFinishDistribution(ctx sdk.Context, streams []types.Stream) {
	for _, stream := range streams {
		// filled epoch is increased in this step and we compare with +1
		if !stream.IsPerpetual && stream.NumEpochsPaidOver <= stream.FilledEpochs+1 {
			if err := k.moveActiveStreamToFinishedStream(ctx, stream); err != nil {
				panic(err)
			}
		}
	}
}

// GetModuleToDistributeCoins returns sum of coins yet to be distributed for all of the module.
func (k Keeper) GetModuleToDistributeCoins(ctx sdk.Context) sdk.Coins {
	activeStreamsDistr := k.getToDistributeCoinsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
	upcomingStreamsDistr := k.getToDistributeCoinsFromIterator(ctx, k.UpcomingStreamsIterator(ctx))
	return activeStreamsDistr.Add(upcomingStreamsDistr...)
}

// GetModuleDistributedCoins returns sum of coins that have been distributed so far for all of the module.
func (k Keeper) GetModuleDistributedCoins(ctx sdk.Context) sdk.Coins {
	activeStreamsDistr := k.getDistributedCoinsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
	finishedStreamsDistr := k.getDistributedCoinsFromIterator(ctx, k.FinishedStreamsIterator(ctx))
	return activeStreamsDistr.Add(finishedStreamsDistr...)
}
