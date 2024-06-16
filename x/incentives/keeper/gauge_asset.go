package keeper

import (
	"fmt"
	"time"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getLocksToDistributionWithMaxDuration returns locks that match the provided lockuptypes QueryCondition,
// are greater than the provided minDuration, AND have yet to be distributed to.
func (k Keeper) getLocksToDistributionWithMaxDuration(ctx sdk.Context, distrTo lockuptypes.QueryCondition, minDuration time.Duration) []lockuptypes.PeriodLock {
	switch distrTo.LockQueryType {
	case lockuptypes.ByDuration:
		denom := distrTo.Denom
		if distrTo.Duration > minDuration {
			return k.lk.GetLocksLongerThanDurationDenom(ctx, denom, minDuration)
		}
		return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, distrTo.Duration)
	case lockuptypes.ByTime:
		panic("Gauge by time is present, however is no longer supported. This should have been blocked in ValidateBasic")
	default:
	}
	return []lockuptypes.PeriodLock{}
}

// FilteredLocksDistributionEst estimates distribution amount of coins from gauge.
// It also applies an update for the gauge, handling the sending of the rewards.
// (Note this update is in-memory, it does not change state.)
func (k Keeper) FilteredLocksDistributionEst(ctx sdk.Context, gauge types.Gauge, filteredLocks []lockuptypes.PeriodLock) (types.Gauge, sdk.Coins, bool, error) {
	disTo, ok := gauge.DistributeTo.(*types.Gauge_Asset)
	if !ok {
		return types.Gauge{}, nil, false, fmt.Errorf("gauge %d is not an asset gauge", gauge.Id)
	}
	TotalAmtLocked := k.lk.GetPeriodLocksAccumulation(ctx, *disTo.Asset)
	if TotalAmtLocked.IsZero() {
		return types.Gauge{}, nil, false, nil
	}
	if TotalAmtLocked.IsNegative() {
		return types.Gauge{}, nil, true, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins...)
	// remainEpochs is the number of remaining epochs that the gauge will pay out its rewards.
	// for a perpetual gauge, it will pay out everything in the next epoch, and we don't make
	// an assumption of the rate at which it will get refilled at.
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}
	if remainEpochs == 0 {
		return gauge, sdk.Coins{}, false, nil
	}

	remainCoinsPerEpoch := sdk.Coins{}
	for _, coin := range remainCoins {
		// distribution amount per epoch = gauge_size / (remain_epochs)
		amt := coin.Amount.QuoRaw(int64(remainEpochs))
		remainCoinsPerEpoch = remainCoinsPerEpoch.Add(sdk.NewCoin(coin.Denom, amt))
	}

	// now we compute the filtered coins
	filteredDistrCoins := sdk.Coins{}
	if len(filteredLocks) == 0 {
		// if were doing no filtering, we want to calculate the total amount to distributed in
		// the next epoch.
		// distribution in next epoch = gauge_size  / (remain_epochs)
		filteredDistrCoins = remainCoinsPerEpoch
	}
	for _, lock := range filteredLocks {
		denomLockAmt := lock.Coins.AmountOf(disTo.Asset.Denom)

		for _, coin := range remainCoinsPerEpoch {
			// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
			// distribution amount = gauge_size_per_epoch * denom_lock_amount / total_denom_lock_amount
			amt := coin.Amount.Mul(denomLockAmt).Quo(TotalAmtLocked)
			filteredDistrCoins = filteredDistrCoins.Add(sdk.NewCoin(coin.Denom, amt))
		}
	}

	// increase filled epochs after distribution
	gauge.FilledEpochs += 1
	gauge.DistributedCoins = gauge.DistributedCoins.Add(remainCoinsPerEpoch...)

	return gauge, filteredDistrCoins, false, nil
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

// distributeInternal runs the distribution logic for a gauge, and adds the sends to
// the distrInfo struct. It also updates the gauge for the distribution.
// Locks is expected to be the correct set of lock recipients for this gauge.
func (k Keeper) distributeToAssetGauge(
	ctx sdk.Context, gauge types.Gauge, distrInfo *distributionInfo,
) (sdk.Coins, error) {
	totalDistrCoins := sdk.NewCoins()
	distributeTo, ok := gauge.DistributeTo.(*types.Gauge_Asset)
	if !ok {
		return totalDistrCoins, fmt.Errorf("gauge %d is not an asset gauge", gauge.Id)
	}

	locksByDenomCache := make(map[string][]lockuptypes.PeriodLock)
	locks := k.getDistributeToBaseLocks(ctx, gauge, locksByDenomCache)

	denom := distributeTo.Asset.Denom
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	if lockSum.IsZero() {
		return totalDistrCoins, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins...)
	// if its a perpetual gauge, we set remaining epochs to 1.
	// otherwise is is a non perpetual gauge and we determine how many epoch payouts are left
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}

	/* ---------------------------- defense in depth ---------------------------- */
	// this should never happen in practice since gauge passed in should always be an active gauge.
	if remainEpochs == uint64(0) {
		ctx.Logger().Error(fmt.Sprintf("gauge %d has no remaining epochs, skipping", gauge.Id))
		return totalDistrCoins, nil
	}

	// this should never happen in practice
	if remainCoins.Empty() {
		ctx.Logger().Error(fmt.Sprintf("gauge %d is empty, skipping", gauge.Id))
		err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
		return totalDistrCoins, err
	}

	for _, lock := range locks {
		distrCoins := sdk.Coins{}
		for _, coin := range remainCoins {
			// distribution amount = gauge_size * denom_lock_amount / (total_denom_lock_amount * remain_epochs)
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

	err := k.updateGaugePostDistribute(ctx, gauge, totalDistrCoins)
	return totalDistrCoins, err
}

// getDistributeToBaseLocks takes a gauge along with cached period locks by denom and returns locks that must be distributed to
func (k Keeper) getDistributeToBaseLocks(ctx sdk.Context, gauge types.Gauge, cache map[string][]lockuptypes.PeriodLock) []lockuptypes.PeriodLock {
	// if gauge is empty, don't get the locks
	if gauge.Coins.Empty() {
		return []lockuptypes.PeriodLock{}
	}

	distTo, ok := gauge.DistributeTo.(*types.Gauge_Asset)
	if !ok {
		panic(fmt.Sprintf("gauge %d is not an asset gauge", gauge.Id))
	}
	// All gauges have a precondition of being ByDuration.
	distributeBaseDenom := distTo.Asset.Denom
	if _, ok := cache[distributeBaseDenom]; !ok {
		cache[distributeBaseDenom] = k.getLocksToDistributionWithMaxDuration(
			ctx, *distTo.Asset, time.Millisecond)
	}
	// get this from memory instead of hitting iterators / underlying stores.
	// due to many details of cacheKVStore, iteration will still cause expensive IAVL reads.
	allLocks := cache[distributeBaseDenom]
	return FilterLocksByMinDuration(allLocks, distTo.Asset.Duration)
}
