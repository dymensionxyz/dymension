package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

// CreateAssetGauge creates a gauge and sends coins to the gauge.
func (k Keeper) CreateAssetGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error) {
	// Ensure that this gauge's duration is one of the allowed durations on chain
	durations := k.GetLockableDurations(ctx)
	if distrTo.LockQueryType != lockuptypes.ByDuration {
		return 0, fmt.Errorf("invalid lock query type: %s", distrTo.LockQueryType)
	}
	durationOk := false
	for _, duration := range durations {
		if duration == distrTo.Duration {
			durationOk = true
			break
		}
	}
	if !durationOk {
		return 0, fmt.Errorf("invalid duration: %d", distrTo.Duration)
	}

	// Ensure that the denom this gauge pays out to exists on-chain
	if !k.bk.HasSupply(ctx, distrTo.Denom) {
		return 0, fmt.Errorf("denom does not exist: %s", distrTo.Denom)
	}

	gauge := types.NewAssetGauge(k.GetLastGaugeID(ctx)+1, isPerpetual, distrTo, coins, startTime, numEpochsPaidOver)
	if err := k.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, gauge.Coins); err != nil {
		return 0, err
	}

	err := k.setGauge(ctx, &gauge)
	if err != nil {
		return 0, err
	}
	k.SetLastGaugeID(ctx, gauge.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingGauges, getTimeKey(gauge.StartTime))
	activeOrUpcomingGauge := true

	err = k.CreateGaugeRefKeys(ctx, &gauge, combinedKeys, activeOrUpcomingGauge)
	if err != nil {
		return 0, err
	}
	k.hooks.AfterCreateGauge(ctx, gauge.Id)
	return gauge.Id, nil
}

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

// DistributionValueCache caches minimum values for token distribution
type DistributionValueCache struct {
	minDistrValue      sdk.Coin
	denomToMinValueMap map[string]math.Int
}

// calculateAssetGaugeRewards computes the reward distribution for an asset gauge based on lock amounts.
// It calculates rewards for each qualifying lock and tracks them in the distribution tracker.
// Returns the total coins allocated for distribution.
func (k Keeper) calculateAssetGaugeRewards(
	ctx sdk.Context,
	gauge types.Gauge,
	locks []lockuptypes.PeriodLock,
	tracker *RewardDistributionTracker,
	minDistrValueCache *DistributionValueCache,
) (sdk.Coins, error) {
	assetDist := gauge.GetAsset()
	if assetDist == nil {
		return sdk.Coins{}, fmt.Errorf("gauge %d is not an asset gauge", gauge.Id)
	}

	denom := assetDist.Denom
	lockSum := lockuptypes.SumLocksByDenom(locks, denom)

	if lockSum.IsZero() {
		return sdk.Coins{}, nil
	}

	// if it's a perpetual gauge, we set remaining epochs to 1.
	// otherwise it is a non perpetual gauge and we determine how many epoch payouts are left
	remainEpochs := int64(1)
	if !gauge.IsPerpetual {
		// this should never happen in practice since gauge passed in should always be an active gauge.
		if gauge.NumEpochsPaidOver <= gauge.FilledEpochs {
			return sdk.Coins{}, fmt.Errorf("gauge %d is not active. num_epochs_paid_over: %d, filled_epochs: %d", gauge.Id, gauge.NumEpochsPaidOver, gauge.FilledEpochs)
		}
		remainEpochs = int64(gauge.NumEpochsPaidOver - gauge.FilledEpochs) //nolint:gosec
	}

	remainCoins := gauge.Coins.Sub(gauge.DistributedCoins...)
	if remainCoins.Empty() {
		ctx.Logger().Error(fmt.Sprintf("gauge %d is empty, skipping", gauge.Id))
		return sdk.Coins{}, nil
	}

	totalDistrCoins := sdk.NewCoins()

	for _, lock := range locks {
		lockRewardCoins := sdk.Coins{}
		lockedAmt := lock.Coins.AmountOfNoDenomValidation(denom)

		for _, coin := range remainCoins {
			// Check the cached minimum value for this denom
			minTokenRequired, ok := minDistrValueCache.denomToMinValueMap[coin.Denom]
			if !ok {
				// get the minimal amount allowed for distribution for this coin
				minAmtForNewCoin, err := k.tk.CalcBaseInCoin(ctx, minDistrValueCache.minDistrValue, coin.Denom)
				if err != nil {
					k.Logger(ctx).Debug("failed to calculate minimal value for denom", "denom", coin.Denom, "error", err)
					minDistrValueCache.denomToMinValueMap[coin.Denom] = math.OneInt().Neg()
					// send unknown denoms to txfees module
					err := k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, txfeestypes.ModuleName, sdk.NewCoins(coin))
					if err != nil {
						k.Logger(ctx).Error("failed to send unknown denom to txfees module", "denom", coin.Denom, "error", err)
					} else {
						// mark this denom as distributed
						totalDistrCoins = totalDistrCoins.Add(coin)
					}
					continue
				}
				minDistrValueCache.denomToMinValueMap[coin.Denom] = minAmtForNewCoin.Amount
				minTokenRequired = minAmtForNewCoin.Amount
			}
			// unsupported reward denom
			if minTokenRequired.IsNegative() {
				continue
			}

			// reward for the lock: (lock_amount / total_lock_amount) * (rewards / remain_epochs)
			// to minimize truncation effects, we use
			// (lock_amount * rewards) / (total_lock_amount * remain_epochs)
			amt := lockedAmt.Mul(coin.Amount).ToLegacyDec().QuoInt(lockSum).QuoInt64(remainEpochs).TruncateInt()

			// Check if the amount is worth distributing based on the minimum distribution value
			if amt.LT(minTokenRequired) {
				continue
			}

			if amt.IsPositive() {
				newlyDistributedCoin := sdk.Coin{Denom: coin.Denom, Amount: amt}
				lockRewardCoins = lockRewardCoins.Add(newlyDistributedCoin)
			}
		}

		if lockRewardCoins.Empty() {
			continue
		}

		// update the amount for that address
		lockRewardCoins = lockRewardCoins.Sort()
		err := tracker.addLockRewards(lock.Owner, gauge.Id, lockRewardCoins)
		if err != nil {
			return sdk.Coins{}, err
		}

		totalDistrCoins = totalDistrCoins.Add(lockRewardCoins...)
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
		cache[distributeBaseDenom] = k.lk.GetLocksLongerThanDurationDenom(ctx, asset.Denom, asset.Duration)
	}
	// get this from memory instead of hitting iterators / underlying stores.
	// due to many details of cacheKVStore, iteration will still cause expensive IAVL reads.
	allLocks := cache[distributeBaseDenom]
	return FilterLocksByMinDuration(allLocks, asset.Duration)
}
