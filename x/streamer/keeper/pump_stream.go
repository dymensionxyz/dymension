package keeper

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (k Keeper) PumpPressure(ctx sdk.Context, distr sponsorshiptypes.Distribution, pumpBudget math.Int) []types.PumpPressure {
	var rollappRecords []types.PumpPressure
	for _, gauge := range distr.Gauges {
		g, err := k.ik.GetGaugeByID(ctx, gauge.GaugeId)
		if err != nil {
			k.Logger(ctx).Error("failed to get gauge", "gaugeID", gauge.GaugeId, "error", err)
			continue
		}
		if ra := g.GetRollapp(); ra != nil {
			rollappRecords = append(rollappRecords, types.PumpPressure{
				RollappId: ra.RollappId,
				// Don't pre-calculate 'pumpBudget / distr.VotingPower' bc it loses precision
				Pressure: gauge.Power.Mul(pumpBudget).Quo(distr.VotingPower),
			})
		}
	}

	// Sort all records which are rollapp gauges by weight in descending order
	sort.Slice(rollappRecords, func(i, j int) bool {
		return rollappRecords[i].Pressure.GT(rollappRecords[j].Pressure)
	})

	return rollappRecords
}

// TotalPumpBudget is the total number of DYM that all pump streams hold.
func (k Keeper) TotalPumpBudget(ctx sdk.Context) math.Int {
	var totalBudget math.Int
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
// epochBlocks is consumed!
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
	pumpNum uint64,
	epochBlocks math.Int, // is consumed!
) (math.Int, error) {
	if pumpNum == 0 {
		// Should not pump at all
		return math.ZeroInt(), nil
	}
	if epochBlocks.IsZero() {
		return math.ZeroInt(), fmt.Errorf("epochBlocks cannot be zero")
	}
	if pumpNum > epochBlocks.Uint64() {
		return math.ZeroInt(), fmt.Errorf("pumpNum (%d) cannot be greater than epochBlocks (%d)", pumpNum, epochBlocks)
	}

	// Scale down the random value to range [0, epochBlocks)
	randomInRange := GenerateUnifiedRandom(ctx, epochBlocks.BigIntMut())

	// Check if the random value falls within the pump probability
	// For pumpNum pumps in epochBlocks: success if random value < pumpNum
	numeratorBig := big.NewInt(int64(pumpNum))

	// If randomInRange < numeratorBig
	if randomInRange.Cmp(numeratorBig) < 0 {
		return PumpAmt(ctx, budget, left, pumpNum)
	}

	return math.ZeroInt(), nil
}

// PumpAmt computes min(Uniform(0, 2 * Budget / PumpNum), Left).
func PumpAmt(ctx sdk.Context, budget math.Int, left math.Int, pumpNum uint64) (math.Int, error) {
	modulo := budget.MulRaw(2).QuoRaw(int64(pumpNum))
	if modulo.IsZero() {
		return math.ZeroInt(), fmt.Errorf("budget per pump is fractional: too small budget (%s) or too many pumps (%d)", budget, pumpNum)
	}
	randBig := GenerateUnifiedRandom(ctx, modulo.BigIntMut())
	rand := math.NewIntFromBigIntMut(randBig)
	return math.MinInt(rand, left), nil
}

// GenerateUnifiedRandom a unified random variable by modulo.
func GenerateUnifiedRandom(ctx sdk.Context, modulo *big.Int) *big.Int {
	h := sha256.New()
	h.Write(ctx.HeaderHash())
	// h.Write(blockTimeBytes)
	seed := h.Sum(nil)

	randomBig := new(big.Int).SetBytes(seed)
	return new(big.Int).Mod(randomBig, modulo)
}

// ExecutePump performs the pump operation by buying tokens for a specific rollapp.
// CONTRACT: pumpAmt is always in base denom.
func (k Keeper) ExecutePump(
	ctx sdk.Context,
	pumpAmt sdk.Coin,
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

	var targetDenom string
	if plan.IsSettled() {
		targetDenom = plan.SettledDenom
	} else {
		targetDenom = plan.LiquidityDenom
	}

	buyer := k.ak.GetModuleAddress(types.ModuleName)
	var tokenOutAmt math.Int

	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		// There are several cases:
		// - targetDenom == plan.SettledDenom =>
		//     IRO is settled =>
		//     RA token exists =>
		//     every RA token has an AMM route to DYM =>
		//     do an AMM swap using feetoken
		//
		// - targetDenom == plan.LiquidityDenom != baseDenom =>
		//     IRO is not settled =>
		//     need to buy IRO using liquidity tokens =>
		//     every liquidity token has an AMM route to DYM =>
		//     do an AMM swap using feetoken =>
		//     buy IRO using tokenOut of liquidity tokens
		//
		// - targetDenom == plan.LiquidityDenom == baseDenom =>
		//     IRO is not settled =>
		//     need to buy IRO using liquidity tokens =>
		//     pump amount is already in base denom =>
		//     buy IRO using pumpAmt of baseDenom

		if targetDenom == pumpAmt.Denom {
			tokenOutAmt = pumpAmt.Amount
		} else {
			feeToken, err := k.txFeesKeeper.GetFeeToken(ctx, targetDenom)
			if err != nil {
				return fmt.Errorf("get fee token for denom %s: %w", targetDenom, err)
			}

			reverseRoute := reverseInRoute(feeToken.Route, targetDenom)
			tokenOutAmt, err = k.poolManagerKeeper.RouteExactAmountIn(
				ctx,
				buyer,
				reverseRoute,
				pumpAmt,        // token in
				math.ZeroInt(), // no slippage
			)
			if err != nil {
				return fmt.Errorf("route exact amount in: target denom: %s, error: %w", targetDenom, err)
			}
		}

		if !plan.IsSettled() {
			// If IRO is in progress, use liquidity tokens to buy IRO tokens
			tokenOutAmt, err = k.iroKeeper.BuyExactSpend(
				ctx,
				fmt.Sprintf("%d", plan.Id),
				buyer,
				tokenOutAmt,    // amountToSpend
				math.ZeroInt(), // no slippage
			)
			if err != nil {
				return fmt.Errorf("buy from IRO %d: %w", plan.Id, err)
			}
		}

		return nil
	})
	if err != nil {
		return sdk.Coin{}, err
	}

	return sdk.NewCoin(targetDenom, tokenOutAmt), nil
}

// DistributePumpStreams processes all pump streams and executes pumps if conditions are met
func (k Keeper) DistributePumpStreams(ctx sdk.Context, pumpStreams []types.Stream) error {
	// All bought tokens should be burned
	toBurn := make(sdk.Coins, 0)
	event := make([]types.EventPumped_Pump, 0)

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
			stream.PumpParams.EpochBudget,
			stream.PumpParams.EpochBudgetLeft,
			stream.PumpParams.NumPumps,
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
		pressure := k.PumpPressure(ctx, sponsorshipDistr, pumpAmt)
		if len(pressure) > int(stream.PumpParams.NumTopRollapps) {
			pressure = pressure[:int(stream.PumpParams.NumTopRollapps)]
		}

		// Distribute pump amount proportionally to each rollapp
		pumpedAmt := math.ZeroInt()
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

			pumpedAmt = pumpedAmt.Add(pumpCoin.Amount)
			toBurn = toBurn.Add(tokenOut)
			event = append(event, types.EventPumped_Pump{
				RollappId: p.RollappId,
				StreamId:  stream.Id,
				PumpAmt:   p.Pressure,
				TokenOut:  tokenOut,
			})
		}

		// Update the stream if needed
		if !pumpedAmt.IsZero() {
			stream.PumpParams.EpochBudgetLeft = stream.PumpParams.EpochBudgetLeft.Sub(pumpedAmt)
			stream.AddDistributedCoins(sdk.NewCoins(sdk.NewCoin(baseDenom, pumpedAmt)))

			err = k.SetStream(ctx, &stream)
			if err != nil {
				return fmt.Errorf("failed to update stream after pump: %w", err)
			}
		}
	}

	if toBurn.Len() != 0 {
		err = k.bk.BurnCoins(ctx, types.ModuleName, toBurn)
		if err != nil {
			return fmt.Errorf("failed to burn coins: %w", err)
		}

		err = uevent.EmitTypedEvent(ctx, &types.EventPumped{Pumps: event})
		if err != nil {
			return fmt.Errorf("emit EventPumped: %w", err)
		}
	}

	return nil
}

// Number of seconds in the year.
// 60 * 60 * 8766 is how the SDK defines it:
// https://github.com/cosmos/cosmos-sdk/blob/v0.50.14/x/mint/types/params.go#L33
const yearSecs = 60 * 8766

func (k Keeper) EpochBlocks(ctx sdk.Context, epochID string) (math.Int, error) {
	info := k.ek.GetEpochInfo(ctx, epochID)
	mintParams, err := k.mintParams.Get(ctx)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("get mint params: %w", err)
	}
	// info.Duration might be "hour", "day", or "week" and is defined as
	// an integer, so it's safe to cast it to uint64.
	var (
		year          = math.NewInt(yearSecs)
		blocksPerYear = math.NewIntFromUint64(mintParams.BlocksPerYear)
		epochSecs     = math.NewInt(int64(info.Duration))
	)
	return epochSecs.Mul(blocksPerYear).Quo(year), nil
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
