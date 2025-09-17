package keeper

import (
	"fmt"
	"math/big"
	"slices"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/rand"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// TopRollapps calculated how much DYM each RA gets if the given budget is
// applied using the given weight distribution. Returns a sorted list of orders
// in descending order by pressure. Pressure for rollapps is normalized
// over the total amount of pressure. Example:
//
//	RA1 gets 30%
//	RA2 gets 50%
//	RA3 gets 20% but is not selected to pump
//	=> Normalize:
//	RA1 gets 3/8 = 37.5%
//	RA2 gets 5/8 = 62.5%
//
// CONTRACT: all sponsorshiptypes.Distribution gauges are rollapp gauges.
func (k Keeper) TopRollapps(ctx sdk.Context, gauges []sponsorshiptypes.Gauge, pumpBudget math.Int, limit *uint32) []types.PumpPressure {
	// Sort in descending order
	slices.SortFunc(gauges, func(left, right sponsorshiptypes.Gauge) int {
		return right.Power.BigIntMut().Cmp(left.Power.BigIntMut())
	})

	if limit != nil && int(*limit) < len(gauges) {
		gauges = gauges[:int(*limit)]
	}

	totalWeight := math.ZeroInt()
	for _, gauge := range gauges {
		totalWeight = totalWeight.Add(gauge.Power)
	}

	return k.PumpPressure(ctx, gauges, pumpBudget, totalWeight)
}

func (k Keeper) PumpPressure(ctx sdk.Context, gauges []sponsorshiptypes.Gauge, pumpBudget, totalWeight math.Int) []types.PumpPressure {
	var rollappRecords []types.PumpPressure
	for _, gauge := range gauges {
		g, err := k.ik.GetGaugeByID(ctx, gauge.GaugeId)
		if err != nil {
			k.Logger(ctx).Error("failed to get gauge", "gaugeID", gauge.GaugeId, "error", err)
			continue
		}
		if ra := g.GetRollapp(); ra != nil {
			rollappRecords = append(rollappRecords, types.PumpPressure{
				RollappId: ra.RollappId,
				// Don't pre-calculate 'pumpBudget / totalWeight' bc it loses precision
				Pressure: gauge.Power.Mul(pumpBudget).Quo(totalWeight),
			})
		}
	}

	return rollappRecords
}

// TotalPumpBudget is the total number of DYM that all pump streams hold.
func (k Keeper) TotalPumpBudget(ctx sdk.Context) math.Int {
	totalBudget := math.ZeroInt()
	for _, stream := range k.GetActiveStreams(ctx) {
		if stream.IsPumpStream() {
			totalBudget = totalBudget.Add(stream.Coins[0].Amount)
		}
	}
	return totalBudget
}

// ShouldPump decides if the pump should happen in this block. It uses block
// hash and block time as entropy.
//
// Example:
//
// seed = 123,456,789,012,345,678,901,234,567,890 (derived from the block context)
// epochBlocks = 1000
// pumpNum = 25
//
//	=> 2.5% of blocks should perform pumps
//
// Scale down the seed to [0, epochBlocks):
// randomInRange = randomBig % 1000 = 890
//
// Since we want 25 pumps out of 1000 blocks:
// - Numbers 0-24 → pump (25 numbers)
// - Numbers 25-999 → no pump (975 numbers)
//
// 890 ≥ 25, so no pump this block
func ShouldPump(
	ctx sdk.Context,
	budget math.Int,
	left math.Int,
	numPumps uint64,
	pumpDistr types.PumpDistr,
	epochBlocks math.Int,
) (math.Int, error) {
	if numPumps <= 0 {
		// Should not pump at all
		return math.ZeroInt(), nil
	}
	if epochBlocks.IsZero() {
		return math.ZeroInt(), fmt.Errorf("epochBlocks cannot be zero")
	}

	// Draw a random value in range [0, epochBlocks)
	randomInRangeBig := rand.GenerateUniformRandomMod(ctx, epochBlocks.BigInt())

	// If NumPumps >= epochBlocks => we should pump on every block
	// If NumPumps < epochBlocks => pump is probabilistic
	// In any case, epochBlocks should fall in uint64 since NumPumps is uint64
	// Therefore, it's safe to cast big.Int to uint64
	randomInRange := randomInRangeBig.Uint64()

	// Check if the random value falls within the pump probability:
	// success if random value < pumpNum
	if randomInRange < numPumps {
		return PumpAmt(ctx, budget, left, math.NewIntFromUint64(numPumps), pumpDistr)
	}

	return math.ZeroInt(), nil
}

// PumpAmt computes the random pump amount.
func PumpAmt(
	ctx sdk.Context,
	budget math.Int,
	left math.Int,
	numPumps math.Int,
	pumpDistr types.PumpDistr,
) (math.Int, error) {
	if budget.LT(numPumps) {
		// The budget is too small to use it for pumping. It might happen
		// close to the stream end if the epoch budget is too low, but it's
		// not probable since EpochBudget ≈ 10^20, numPumps ≈ 10^5.
		return math.ZeroInt(), nil
	}

	randBig := new(big.Int)

	switch pumpDistr {
	case types.PumpDistr_PUMP_DISTR_UNIFORM:
		// Draw a Uniform(0; 2*B/N) value
		// Mean is B/N
		modulo := budget.MulRaw(2).Quo(numPumps)
		randBig = rand.GenerateUniformRandomMod(ctx, modulo.BigIntMut())

	case types.PumpDistr_PUMP_DISTR_EXPONENTIAL:
		// Draw an Exp(N/B) value
		// Mean is B/N
		randBig = rand.GenerateExpRandomLambda(ctx, numPumps.BigIntMut(), budget.BigInt())

	case types.PumpDistr_PUMP_DISTR_UNSPECIFIED:
		return math.ZeroInt(), fmt.Errorf("pump distribution not specified")
	}

	r := math.NewIntFromBigIntMut(randBig)
	return math.MinInt(r, left), nil
}

// ExecutePump performs the pump operation by buying tokens for a specific rollapp.
func (k Keeper) ExecutePump(
	ctx sdk.Context,
	pumpCoin sdk.Coin,
	rollappID string,
) (tokenOut sdk.Coin, err error) {
	rollapp, found := k.rollappKeeper.GetRollapp(ctx, rollappID)
	if !found {
		return sdk.Coin{}, fmt.Errorf("rollapp not found: %s", rollappID)
	}

	plan, found := k.iroKeeper.GetPlanByRollapp(ctx, rollapp.RollappId)
	if !found {
		return sdk.Coin{}, fmt.Errorf("IRO plan not found for rollapp: %s", rollappID)
	}

	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		liquidityAmt, err := k.swapPumpCoinToLiquidityDenom(ctx, pumpCoin, plan.LiquidityDenom)
		if err != nil {
			return err
		}

		switch plan.GetGraduationStatus() {
		case irotypes.GraduationStatus_PRE_GRADUATION:
			// IRO is not graduated:
			//   - How many IRO tokens can buy with pumpAmt considering taker fee?
			//   - If tokenOut is greater than plan's MaxAmountToSell buy only the remaining part: MaxAmountToSell - SoldAmt
			//   - If the plan graduates and there's any leftover, spend it in the respective AMM pool
			//   - Return Coin(IRO denom; IRO token out)
			tokenOut, err = k.executePumpPreGraduation(ctx, liquidityAmt, plan)
			if err != nil {
				return fmt.Errorf("execute pump pre-graduation: %w", err)
			}
		case irotypes.GraduationStatus_GRADUATED:
			// IRO is graduated:
			//   - Buy IRO tokens from AMM using GraduatedPoolId
			//   - Return Coin(IRO denom; IRO token out)
			tokenOut, err = k.executePumpGraduated(ctx, liquidityAmt, plan)
			if err != nil {
				return fmt.Errorf("execute pump graduated: %w", err)
			}
		case irotypes.GraduationStatus_SETTLED:
			// IRO is settled:
			//   - Buy RA tokens from AMM using GraduatedPoolId
			//   - Return Coin(settled denom; RA token out)
			tokenOut, err = k.executePumpSettled(ctx, liquidityAmt, plan)
			if err != nil {
				return fmt.Errorf("execute pump settled: %w", err)
			}
		default:
			// Should never happen
			return fmt.Errorf("unknown graduation status: %v", plan.GetGraduationStatus())
		}

		return nil
	})

	return tokenOut, err
}

// swapPumpCoinToLiquidityDenom swaps pump tokens to the plan's liquidity denomination if needed
func (k Keeper) swapPumpCoinToLiquidityDenom(
	ctx sdk.Context,
	pumpCoin sdk.Coin,
	liquidityDenom string,
) (math.Int, error) {
	if pumpCoin.Denom == liquidityDenom {
		return pumpCoin.Amount, nil
	}

	feeToken, err := k.txFeesKeeper.GetFeeToken(ctx, liquidityDenom)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("get fee token for denom %s: %w", liquidityDenom, err)
	}

	reverseRoute := reverseInRoute(feeToken.Route, liquidityDenom)
	liquidityAmt, err := k.poolManagerKeeper.RouteExactAmountIn(
		ctx,
		k.ak.GetModuleAddress(types.ModuleName),
		reverseRoute,
		pumpCoin,
		math.ZeroInt(),
	)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("route exact amount in: target denom: %s, error: %w", liquidityDenom, err)
	}

	return liquidityAmt, nil
}

// CONTRACT: plan is pre-graduated
func (k Keeper) executePumpPreGraduation(
	ctx sdk.Context,
	amountToSpend math.Int,
	plan irotypes.Plan,
) (sdk.Coin, error) {
	// Find how many tokens we can buy with pumpAmt
	// Subtract taker fee -> get net amount to spend
	toSpendMinusTakerFeeAmt, _, err := k.iroKeeper.ApplyTakerFee(amountToSpend, k.iroKeeper.GetParams(ctx).TakerFee, false)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("apply taker fee: %w", err)
	}

	tokenOutAmt, err := plan.BondingCurve.TokensForExactInAmount(plan.SoldAmt, toSpendMinusTakerFeeAmt)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("tokens for exact in amount: %w", err)
	}

	remaining := plan.MaxAmountToSell.Sub(plan.SoldAmt)

	if tokenOutAmt.LT(remaining) {
		// IRO has enough tokens to sell, use `BuyExactSpend` for accuracy
		boughtAmt, err := k.iroKeeper.BuyExactSpend(
			ctx,
			plan.GetID(),
			k.ak.GetModuleAddress(types.ModuleName),
			amountToSpend,
			math.ZeroInt(), // no min tokens
		)
		if err != nil {
			return sdk.Coin{}, fmt.Errorf("buy from IRO %d: %w", plan.Id, err)
		}
		return sdk.NewCoin(plan.GetIRODenom(), boughtAmt), nil
	} else {
		// If too many tokens, buy only the remaining and trigger graduation (use `Buy`).
		//
		// amountToSpend covers costs for tokenOutAmt and tokenOutAmt >= remaining,
		// thus amountToSpend covers costs for the remaining, and actuallySpent < amountToSpend.
		actuallySpent, err := k.iroKeeper.Buy(
			ctx,
			plan.GetID(),
			k.ak.GetModuleAddress(types.ModuleName),
			remaining,
			amountToSpend,
		)
		if err != nil {
			return sdk.Coin{}, fmt.Errorf("buy from IRO %d: %w", plan.Id, err)
		}

		plan = k.iroKeeper.MustGetPlan(ctx, plan.GetID())
		if !plan.IsGraduated() {
			// Sanity check. Should never happen.
			return sdk.Coin{}, fmt.Errorf("plan is not graduated after buying max number of IRO tokens")
		}

		tokenOut := sdk.NewCoin(plan.GetIRODenom(), remaining)

		// We bought the max number of IRO tokens, and the plan should be graduated now.
		// In that case, we want to buy the leftover from the AMM pool (if any).
		leftover := amountToSpend.Sub(actuallySpent)
		if !leftover.IsZero() {
			graduatedTokenOut, err := k.executePumpGraduated(ctx, leftover, plan)
			if err != nil {
				return sdk.Coin{}, fmt.Errorf("execute pump graduated: %w", err)
			}
			tokenOut = tokenOut.Add(graduatedTokenOut)
		}

		return tokenOut, nil
	}
}

// CONTRACT: plan is graduated
func (k Keeper) executePumpGraduated(ctx sdk.Context, amountToSpend math.Int, plan irotypes.Plan) (sdk.Coin, error) {
	return k.executePumpAmm(ctx, amountToSpend, plan.LiquidityDenom, plan.GetIRODenom(), plan.GraduatedPoolId)
}

// CONTRACT: plan is settled
func (k Keeper) executePumpSettled(ctx sdk.Context, amountToSpend math.Int, plan irotypes.Plan) (sdk.Coin, error) {
	return k.executePumpAmm(ctx, amountToSpend, plan.LiquidityDenom, plan.SettledDenom, plan.GraduatedPoolId)
}

func (k Keeper) executePumpAmm(
	ctx sdk.Context,
	amountToSpend math.Int,
	liquidityDenom string,
	tokenOutDenom string,
	poolId uint64,
) (sdk.Coin, error) {
	tokenOutAmt, err := k.poolManagerKeeper.RouteExactAmountIn(
		ctx,
		k.ak.GetModuleAddress(types.ModuleName),
		[]poolmanagertypes.SwapAmountInRoute{{
			PoolId:        poolId,
			TokenOutDenom: tokenOutDenom,
		}},
		sdk.NewCoin(liquidityDenom, amountToSpend),
		math.ZeroInt(),
	)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("route exact amount in: target denom: %s, error: %w", liquidityDenom, err)
	}

	return sdk.NewCoin(tokenOutDenom, tokenOutAmt), nil
}

// DistributePumpStreams processes all pump streams and executes pumps if conditions are met
func (k Keeper) DistributePumpStreams(ctx sdk.Context, pumpStreams []types.Stream) error {
	sponsorshipDistr, err := k.sk.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sponsorship distribution: %w", err)
	}

	for _, stream := range pumpStreams {
		if !stream.IsPumpStream() {
			// Skip non-pump streams
			continue
		}

		epochBlocks, err := k.EpochBlocks(ctx, stream.DistrEpochIdentifier)
		if err != nil {
			return fmt.Errorf("failed to get epoch blocks: %w", err)
		}

		pumpAmt, err := ShouldPump(
			ctx,
			stream.EpochCoins[0].Amount,
			stream.PumpParams.EpochCoinsLeft[0].Amount,
			stream.PumpParams.NumPumps,
			stream.PumpParams.PumpDistr,
			epochBlocks,
		)
		if err != nil {
			return fmt.Errorf("failed to calculate pump amount: %w", err)
		}

		if pumpAmt.IsZero() {
			// Shouldn't pump on this iteration
			continue
		}

		var totalPumped sdk.Coins
		event := types.EventPumped{StreamId: stream.Id}

		switch t := stream.PumpParams.Target.(type) {
		case *types.PumpParams_Pool:
			var e types.EventPumped_Pool
			totalPumped, e, err = k.DistributePool(
				ctx,
				pumpAmt,
				stream.Coins[0].Denom,
				*t.Pool,
			)
			if err != nil {
				return fmt.Errorf("failed to distribute pool: %w", err)
			}
			event.Pool = &e

		case *types.PumpParams_Rollapps:
			var e []types.EventPumped_Rollapp
			totalPumped, e, err = k.DistributeRollapps(
				ctx,
				pumpAmt,
				stream.Coins[0].Denom, // this denom is always the base denom when pumping rollapps
				sponsorshipDistr.Gauges,
				*t.Rollapps,
			)
			if err != nil {
				return fmt.Errorf("failed to distribute rollapps: %w", err)
			}
			event.Rollapps = e
		}

		// Update the stream if needed
		if !totalPumped.IsZero() {
			stream.PumpParams.EpochCoinsLeft = stream.PumpParams.EpochCoinsLeft.Sub(totalPumped...)
			stream.AddDistributedCoins(totalPumped)

			err = k.SetStream(ctx, &stream)
			if err != nil {
				return fmt.Errorf("failed to update stream after pump: %w", err)
			}

			err = uevent.EmitTypedEvent(ctx, &event)
			if err != nil {
				return fmt.Errorf("emit EventPumped: %w", err)
			}
		}
	}

	return nil
}

func (k Keeper) DistributeRollapps(
	ctx sdk.Context,
	pumpAmt math.Int,
	pumpDenom string,
	gauges sponsorshiptypes.Gauges,
	rollapps types.TargetTopRollapps,
) (distributed sdk.Coins, events []types.EventPumped_Rollapp, err error) {
	// Get top N rollapps by cast voting power
	pressure := k.TopRollapps(ctx, gauges, pumpAmt, &rollapps.NumTopRollapps)

	// Distribute pump amount proportionally to each rollapp
	for _, p := range pressure {
		if p.Pressure.IsZero() {
			continue
		}

		pumpCoin := sdk.NewCoin(pumpDenom, p.Pressure)

		tokenOut, err := k.ExecutePump(ctx, pumpCoin, p.RollappId)
		if err != nil {
			k.Logger(ctx).Error("failed to execute pump", "rollappID", p.RollappId, "error", err)
			// Continue with other rollapps even if one fails
			continue
		}

		distributed = distributed.Add(pumpCoin)
		events = append(events, types.EventPumped_Rollapp{
			RollappId: p.RollappId,
			PumpCoin:  pumpCoin,
			TokenOut:  tokenOut,
		})
	}

	return distributed, events, nil
}

func (k Keeper) DistributePool(
	ctx sdk.Context,
	pumpAmt math.Int,
	pumpDenom string,
	pool types.TargetPool,
) (distributed sdk.Coins, event types.EventPumped_Pool, err error) {
	tokenOutAmt, err := k.poolManagerKeeper.RouteExactAmountIn(
		ctx,
		k.ak.GetModuleAddress(types.ModuleName),
		[]poolmanagertypes.SwapAmountInRoute{{
			PoolId:        pool.PoolId,
			TokenOutDenom: pool.TokenOut,
		}},
		sdk.NewCoin(pumpDenom, pumpAmt),
		math.ZeroInt(),
	)
	if err != nil {
		return nil, types.EventPumped_Pool{}, fmt.Errorf("route exact amount in: target denom: %s, error: %w", pool.TokenOut, err)
	}
	pumpCoin := sdk.NewCoin(pumpDenom, pumpAmt)
	event = types.EventPumped_Pool{
		PoolId:   pool.PoolId,
		PumpCoin: pumpCoin,
		TokenOut: sdk.NewCoin(pool.TokenOut, tokenOutAmt),
	}
	return sdk.NewCoins(pumpCoin), event, nil
}

// Number of milliseconds in the year.
// 60 * 60 * 8766 is how the SDK defines it:
// https://github.com/cosmos/cosmos-sdk/blob/v0.50.14/x/mint/types/params.go#L33
const yearMs = 60 * 60 * 8766 * 1000

func (k Keeper) EpochBlocks(ctx sdk.Context, epochID string) (math.Int, error) {
	info := k.ek.GetEpochInfo(ctx, epochID)
	mintParams, err := k.mintParams.Get(ctx)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("get mint params: %w", err)
	}
	// info.Duration might be "hour", "day", or "week" and is defined as
	// an integer, so it's safe to cast it to uint64.
	var (
		year          = math.NewInt(yearMs)
		blocksPerYear = math.NewIntFromUint64(mintParams.BlocksPerYear)
		epochMs       = math.NewInt(info.Duration.Milliseconds())
	)
	return epochMs.Mul(blocksPerYear).Quo(year), nil
}

// copy of https://github.com/dymensionxyz/osmosis/blob/4e25bd944ed7b5d4b83b023715a141f0aa6cb4f8/x/txfees/keeper/fees.go#L236
func reverseInRoute(feeTokenRoute []poolmanagertypes.SwapAmountInRoute, denom string) []poolmanagertypes.SwapAmountInRoute {
	newInRoute := make([]poolmanagertypes.SwapAmountInRoute, len(feeTokenRoute))

	lstIdx := len(feeTokenRoute) - 1
	for i := lstIdx; i >= 0; i-- {
		inRoute := feeTokenRoute[i]
		var outDenom string
		if i > 0 {
			outDenom = feeTokenRoute[i-1].TokenOutDenom
		} else {
			outDenom = denom
		}

		j := lstIdx - i
		newInRoute[j] = poolmanagertypes.SwapAmountInRoute{
			PoolId:        inRoute.PoolId,
			TokenOutDenom: outDenom,
		}
	}

	return newInRoute
}
