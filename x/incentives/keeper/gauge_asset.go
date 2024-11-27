package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// RewardDistributionTracker maintains the state of pending reward distributions,
// tracking both total rewards and per-gauge rewards for each recipient.
// It uses array-based storage for better cache locality during distribution.
type RewardDistributionTracker struct {
	nextID            int                    // Next available ID for new recipients
	lockOwnerAddrToID map[string]int         // Maps lock owner addresses to their array index
	idToBech32Addr    []string               // Recipient bech32 addresses indexed by ID
	idToDecodedAddr   []sdk.AccAddress       // Decoded recipient addresses indexed by ID
	idToDistrCoins    []sdk.Coins            // Total rewards per recipient indexed by ID
	idToGaugeRewards  []map[uint64]sdk.Coins // Per-gauge rewards for each recipient indexed by ID
}

// NewRewardDistributionTracker creates a new tracker for managing reward distributions
func NewRewardDistributionTracker() RewardDistributionTracker {
	return RewardDistributionTracker{
		nextID:            0,
		lockOwnerAddrToID: make(map[string]int),
		idToBech32Addr:    []string{},
		idToDecodedAddr:   []sdk.AccAddress{},
		idToDistrCoins:    []sdk.Coins{},
		idToGaugeRewards:  []map[uint64]sdk.Coins{},
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
func (d *RewardDistributionTracker) addLockRewards(owner string, gaugeID uint64, rewards sdk.Coins) error {
	if id, ok := d.lockOwnerAddrToID[owner]; ok {
		// Update total rewards
		oldDistrCoins := d.idToDistrCoins[id]
		d.idToDistrCoins[id] = rewards.Add(oldDistrCoins...)

		// Update gauge rewards (idToGaugeRewards[id] already initialized on first creation)
		if existing, ok := d.idToGaugeRewards[id][gaugeID]; ok {
			d.idToGaugeRewards[id][gaugeID] = existing.Add(rewards...)
		} else {
			d.idToGaugeRewards[id][gaugeID] = rewards
		}
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

		// Initialize and set gauge rewards
		gaugeRewards := make(map[uint64]sdk.Coins)
		gaugeRewards[gaugeID] = rewards
		d.idToGaugeRewards = append(d.idToGaugeRewards, gaugeRewards)
	}
	return nil
}

// GetEvents returns distribution events for all recipients.
// For each recipient, it creates a single event with attributes for each gauge's rewards.
func (d *RewardDistributionTracker) GetEvents() sdk.Events {
	events := make(sdk.Events, 0, len(d.idToBech32Addr))

	for id := 0; id < len(d.idToBech32Addr); id++ {
		attributes := []sdk.Attribute{
			sdk.NewAttribute(types.AttributeReceiver, d.idToBech32Addr[id]),
			sdk.NewAttribute(types.AttributeAmount, d.idToDistrCoins[id].String()),
		}

		// Add attributes for each gauge's rewards (events doesn't requires deterministic order)
		for gaugeID, gaugeRewards := range d.idToGaugeRewards[id] {
			attributes = append(attributes,
				sdk.NewAttribute(
					fmt.Sprintf("%s_%d", types.AttributeGaugeID, gaugeID),
					gaugeRewards.String(),
				),
			)
		}

		events = append(events, sdk.NewEvent(
			types.TypeEvtDistribution,
			attributes...,
		))
	}

	return events
}

// distributeTrackedRewards sends the tracked rewards from the module account to recipients
// and emits corresponding events for each gauge's rewards.
func (k Keeper) distributeTrackedRewards(ctx sdk.Context, tracker *RewardDistributionTracker) error {
	numIDs := len(tracker.idToDecodedAddr)
	if len(tracker.idToDistrCoins) != numIDs || len(tracker.idToGaugeRewards) != numIDs {
		return fmt.Errorf("number of addresses, coins, and gauge rewards to distribute must be equal")
	}
	ctx.Logger().Debug("Beginning distribution to users", "num_of_user", numIDs)

	// First send all rewards
	for id := 0; id < numIDs; id++ {
		err := k.bk.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			tracker.idToDecodedAddr[id],
			tracker.idToDistrCoins[id])
		if err != nil {
			return err
		}
	}

	// Emit all events
	ctx.EventManager().EmitEvents(tracker.GetEvents())

	ctx.Logger().Debug("Finished Distributing to users")
	return nil
}

// calculateAssetGaugeRewards computes the reward distribution for an asset gauge based on lock amounts.
// It calculates rewards for each qualifying lock and tracks them in the distribution tracker.
// Returns the total coins allocated for distribution.
func (k Keeper) calculateAssetGaugeRewards(ctx sdk.Context, gauge types.Gauge, locks []lockuptypes.PeriodLock, tracker *RewardDistributionTracker) (sdk.Coins, error) {
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
		err := tracker.addLockRewards(lock.Owner, gauge.Id, distrCoins)
		if err != nil {
			return sdk.Coins{}, err
		}

		totalDistrCoins = totalDistrCoins.Add(distrCoins...)
	}

	return totalDistrCoins, nil
}

// GetDistributeToBaseLocks takes a gauge along with cached period locks by denom and returns locks that must be distributed to
func (k Keeper) GetDistributeToBaseLocks(ctx sdk.Context, gauge types.Gauge, cache types.DenomLocksCache) []lockuptypes.PeriodLock {
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
