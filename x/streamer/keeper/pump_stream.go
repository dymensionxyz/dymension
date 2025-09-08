package keeper

import (
	"fmt"
	"math/big"
	"slices"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/rand"
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
	pumpParams types.PumpParams,
	epochBlocks math.Int,
) (math.Int, error) {
	if pumpParams.NumPumps == 0 {
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
	if randomInRange < pumpParams.NumPumps {
		return PumpAmt(ctx, pumpParams)
	}

	return math.ZeroInt(), nil
}

// PumpAmt computes the random pump amount.
func PumpAmt(ctx sdk.Context, pumpParams types.PumpParams) (math.Int, error) {
	numPumps := math.NewIntFromUint64(pumpParams.NumPumps)

	if pumpParams.EpochBudget.LT(numPumps) {
		// The budget is too small to use it for pumping. It might happen
		// close to the stream end if the epoch budget is too low, but it's
		// not probable since EpochBudget ≈ 10^20, numPumps ≈ 10^5.
		return math.ZeroInt(), nil
	}

	randBig := new(big.Int)

	switch pumpParams.PumpDistr {
	case types.PumpDistr_PUMP_DISTR_UNIFORM:
		// Draw a Uniform(0; 2*B/N) value
		// Mean is B/N
		modulo := pumpParams.EpochBudget.MulRaw(2).Quo(numPumps)
		randBig = rand.GenerateUniformRandomMod(ctx, modulo.BigIntMut())

	case types.PumpDistr_PUMP_DISTR_EXPONENTIAL:
		// Draw an Exp(N/B) value
		// Mean is B/N
		randBig = rand.GenerateExpRandomLambda(ctx, numPumps.BigIntMut(), pumpParams.EpochBudget.BigInt())

	case types.PumpDistr_PUMP_DISTR_UNSPECIFIED:
		return math.ZeroInt(), fmt.Errorf("pump distribution not specified")
	}

	r := math.NewIntFromBigIntMut(randBig)
	return math.MinInt(r, pumpParams.EpochBudgetLeft), nil
}

// ExecutePump performs the pump operation by buying tokens for a specific rollapp.
// CONTRACT: pumpAmt is always in base denom.
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

	buyer := k.ak.GetModuleAddress(types.ModuleName)

	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		if plan.IsSettled() {
			// IRO is settled -> use AMM:
			//   1. RA token exists
			//   2. Every RA token has an AMM route to DYM
			//   3. Do an AMM swap: PumpDenom -> SettledDenom
			//   4. Return Coin(SettledDenom; AMM token out)
			tokenOutDenom := plan.SettledDenom
			feeToken, err := k.txFeesKeeper.GetFeeToken(ctx, tokenOutDenom)
			if err != nil {
				return fmt.Errorf("get fee token for denom %s: %w", tokenOutDenom, err)
			}

			reverseRoute := reverseInRoute(feeToken.Route, tokenOutDenom)
			tokenOutAmt, err := k.poolManagerKeeper.RouteExactAmountIn(
				ctx,
				buyer,
				reverseRoute,
				pumpCoin,       // token in
				math.ZeroInt(), // no slippage
			)
			if err != nil {
				return fmt.Errorf("route exact amount in: target denom: %s, error: %w", tokenOutDenom, err)
			}

			tokenOut = sdk.NewCoin(tokenOutDenom, tokenOutAmt)
			return nil
		} else {
			// IRO is not settled:
			//   1. Need to buy IRO using LiquidityDenom
			//   2. if (PumpDenom != LiquidityDenom), do an AMM swap: PumpDenom -> LiquidityDenom
			//   3. Buy IRO: LiquidityDenom -> IRO denom
			//   4. Return Coin(IRO denom; IRO token out)
			pumpAmt := pumpCoin.Amount
			if pumpCoin.Denom != plan.LiquidityDenom {
				feeToken, err := k.txFeesKeeper.GetFeeToken(ctx, plan.LiquidityDenom)
				if err != nil {
					return fmt.Errorf("get fee token for denom %s: %w", plan.LiquidityDenom, err)
				}

				reverseRoute := reverseInRoute(feeToken.Route, plan.LiquidityDenom)
				pumpAmt, err = k.poolManagerKeeper.RouteExactAmountIn(
					ctx,
					buyer,
					reverseRoute,
					pumpCoin,       // token in
					math.ZeroInt(), // no slippage
				)
				if err != nil {
					return fmt.Errorf("route exact amount in: target denom: %s, error: %w", plan.LiquidityDenom, err)
				}
			}

			tokenOutAmt, err := k.iroKeeper.BuyExactSpend(
				ctx,
				fmt.Sprintf("%d", plan.Id),
				buyer,
				pumpAmt,        // amountToSpend
				math.ZeroInt(), // no slippage
			)
			if err != nil {
				return fmt.Errorf("buy from IRO %d: %w", plan.Id, err)
			}

			tokenOut = sdk.NewCoin(plan.GetIRODenom(), tokenOutAmt)
			return nil
		}
	})
	if err != nil {
		return sdk.Coin{}, err
	}

	return tokenOut, nil
}

// DistributePumpStreams processes all pump streams and executes pumps if conditions are met
func (k Keeper) DistributePumpStreams(ctx sdk.Context, pumpStreams []types.Stream) error {
	// All bought tokens should be burned
	toBurn := make(sdk.Coins, 0)

	sponsorshipDistr, err := k.sk.GetDistribution(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sponsorship distribution: %w", err)
	}

	// Always use base denom for budget
	baseDenom, err := k.txFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return fmt.Errorf("get base denom: %w", err)
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
			*stream.PumpParams,
			epochBlocks,
		)
		if err != nil {
			return fmt.Errorf("failed to calculate pump amount: %w", err)
		}

		if pumpAmt.IsZero() {
			// Shouldn't pump on this iteration
			continue
		}

		// Get top N rollapps by cast voting power
		pressure := k.TopRollapps(ctx, sponsorshipDistr.Gauges, pumpAmt, &stream.PumpParams.NumTopRollapps)

		totalPumped := sdk.NewCoins()
		events := make([]types.EventPumped_Pump, 0)

		// Distribute pump amount proportionally to each rollapp
		for _, p := range pressure {
			if p.Pressure.IsZero() {
				continue
			}
			pumpCoin := sdk.NewCoin(baseDenom, p.Pressure)

			tokenOut, err := k.ExecutePump(ctx, pumpCoin, p.RollappId)
			if err != nil {
				k.Logger(ctx).Error("failed to execute pump", "streamID", stream.Id, "rollappID", p.RollappId, "error", err)
				// Continue with other rollapps even if one fails
				continue
			}

			totalPumped = totalPumped.Add(pumpCoin)
			toBurn = toBurn.Add(tokenOut)
			events = append(events, types.EventPumped_Pump{
				RollappId: p.RollappId,
				PumpAmt:   p.Pressure,
				TokenOut:  tokenOut,
			})
		}

		// Update the stream if needed
		if !totalPumped.IsZero() {
			stream.PumpParams.EpochBudgetLeft = stream.PumpParams.EpochBudgetLeft.Sub(totalPumped.AmountOfNoDenomValidation(baseDenom))
			stream.AddDistributedCoins(totalPumped)

			err = k.SetStream(ctx, &stream)
			if err != nil {
				return fmt.Errorf("failed to update stream after pump: %w", err)
			}

			err = uevent.EmitTypedEvent(ctx, &types.EventPumped{StreamId: stream.Id, Pumps: events})
			if err != nil {
				return fmt.Errorf("emit EventPumped: %w", err)
			}
		}
	}

	if toBurn.Len() != 0 {
		err = k.bk.BurnCoins(ctx, types.ModuleName, toBurn)
		if err != nil {
			return fmt.Errorf("failed to burn coins: %w", err)
		}
	}

	return nil
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
