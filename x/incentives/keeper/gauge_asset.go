package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// distributionInfo stores all of the information for rewards distributions.
type distributionInfo struct {
	nextID            int
	lockOwnerAddrToID map[string]int
	idToBech32Addr    []string
	idToDecodedAddr   []sdk.AccAddress
	idToDistrCoins    []sdk.Coins
	// TODO: add totalDistrCoins to track total coins distributed
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

// getLocksToDistributionWithMaxDuration returns locks that match the provided lockuptypes QueryCondition,
// are greater than the provided minDuration, AND have yet to be distributed to.
func (k Keeper) getLocksToDistributionWithMaxDuration(ctx sdk.Context, distrTo lockuptypes.QueryCondition, minDuration time.Duration) []lockuptypes.PeriodLock {
	switch distrTo.LockQueryType {
	case lockuptypes.ByDuration:
		// TODO: what the meaning of minDuration here? it's set to time.Millisecond in the caller.
		duration := min(distrTo.Duration, minDuration)
		return k.lk.GetLocksLongerThanDurationDenom(ctx, distrTo.Denom, duration)
	case lockuptypes.ByTime:
		panic("Gauge by time is present, however is no longer supported. This should have been blocked in ValidateBasic")
	default:
	}
	return []lockuptypes.PeriodLock{}
}

// addLockRewards adds the provided rewards to the lockID mapped to the provided owner address.
func (d *distributionInfo) addLockRewards(owner string, rewards sdk.Coins) error {
	if id, ok := d.lockOwnerAddrToID[owner]; ok {
		oldDistrCoins := d.idToDistrCoins[id]
		d.idToDistrCoins[id] = rewards.Add(oldDistrCoins...)
	} else {
		id := d.nextID
		d.nextID++
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

// sendRewardsToLocks utilizes provided distributionInfo to send coins from the module account to various recipients.
func (k Keeper) sendRewardsToLocks(ctx sdk.Context, distrs *distributionInfo) error {
	numIDs := len(distrs.idToDecodedAddr)
	if len(distrs.idToDistrCoins) != numIDs {
		return fmt.Errorf("number of addresses and coins to distribute to must be equal")
	}
	ctx.Logger().Debug("Beginning distribution to users", "num_of_user", numIDs)

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
	ctx.Logger().Debug("Finished Distributing to users")
	return nil
}

// distributeToAssetGauge runs the distribution logic for a gauge, and adds the sends to
// the distrInfo struct. It also updates the gauge for the distribution.
// Locks is expected to be the correct set of lock recipients for this gauge.
func (k Keeper) distributeToAssetGauge(ctx sdk.Context, gauge types.Gauge, locks []lockuptypes.PeriodLock, currResult *distributionInfo) (sdk.Coins, error) {
	assetDist := gauge.GetAsset()
	if assetDist == nil {
		return sdk.Coins{}, fmt.Errorf("gauge %d is not an asset gauge", gauge.Id)
	}

	denom := assetDist.Denom
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	if lockSum.IsZero() {
		return sdk.Coins{}, nil
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins...)
	// if it's a perpetual gauge, we set remaining epochs to 1.
	// otherwise it is a non perpetual gauge and we determine how many epoch payouts are left
	remainEpochs := uint64(1)
	if !gauge.IsPerpetual {
		remainEpochs = gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}

	/* ---------------------------- defense in depth ---------------------------- */
	// this should never happen in practice since gauge passed in should always be an active gauge.
	if remainEpochs == 0 {
		ctx.Logger().Error(fmt.Sprintf("gauge %d has no remaining epochs, skipping", gauge.Id))
		return sdk.Coins{}, nil
	}

	// this should never happen in practice
	if remainCoins.Empty() {
		ctx.Logger().Error(fmt.Sprintf("gauge %d is empty, skipping", gauge.Id))
		return sdk.Coins{}, nil
	}

	totalDistrCoins := sdk.NewCoins()
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
		err := currResult.addLockRewards(lock.Owner, distrCoins)
		if err != nil {
			return sdk.Coins{}, err
		}

		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	return totalDistrCoins, nil
}

// GetDistributeToBaseLocks takes a gauge along with cached period locks by denom and returns locks that must be distributed to
func (k Keeper) GetDistributeToBaseLocks(ctx sdk.Context, gauge types.Gauge, cache map[string][]lockuptypes.PeriodLock) []lockuptypes.PeriodLock {
	// if gauge is empty, don't get the locks
	if gauge.Coins.Empty() {
		return []lockuptypes.PeriodLock{}
	}

	// All gauges have a precondition of being ByDuration.
	asset := gauge.GetAsset() // this should never be nil
	distributeBaseDenom := asset.Denom
	if _, ok := cache[distributeBaseDenom]; !ok {
		cache[distributeBaseDenom] = k.getLocksToDistributionWithMaxDuration(ctx, *asset, time.Millisecond)
	}
	// get this from memory instead of hitting iterators / underlying stores.
	// due to many details of cacheKVStore, iteration will still cause expensive IAVL reads.
	allLocks := cache[distributeBaseDenom]
	return FilterLocksByMinDuration(allLocks, asset.Duration)
}
