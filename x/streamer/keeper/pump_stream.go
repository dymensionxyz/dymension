package keeper

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"sort"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

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
	pumpNum uint64,
	epochBlocks uint64,
) (math.Int, error) {
	if pumpNum == 0 {
		// Should not pump at all
		return math.ZeroInt(), nil
	}
	if epochBlocks == 0 {
		return math.ZeroInt(), fmt.Errorf("epochBlocks cannot be zero")
	}
	if pumpNum > epochBlocks {
		return math.ZeroInt(), fmt.Errorf("pumpNum (%d) cannot be greater than epochBlocks (%d)", pumpNum, epochBlocks)
	}

	// Scale down the random value to range [0, epochBlocks)
	modulo := new(big.Int).SetUint64(epochBlocks)
	randomInRange := GenerateUnifiedRandom(ctx, modulo)

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
func (k Keeper) ExecutePump(
	ctx sdk.Context,
	pumpAmt math.Int,
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

	// Always use base denom for budget
	baseDenom, err := k.txFeesKeeper.GetBaseDenom(ctx)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("get base denom: %w", err)
	}

	var targetDenom string
	if plan.IsSettled() {
		targetDenom = plan.SettledDenom
	} else {
		targetDenom = plan.LiquidityDenom
	}

	// Get FeeToken for target denom to find routing to the base denom.
	// Every token must have a route to the base denom.
	feeToken, err := k.txFeesKeeper.GetFeeToken(ctx, targetDenom)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("get fee token for denom %s: %w", targetDenom, err)
	}

	buyer := k.ak.GetModuleAddress(types.ModuleName)
	var tokenOutAmt math.Int

	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		// Buy:
		// - RA tokens if IRO is over
		// - Liquidity tokens if IRO is in progress
		tokenOutAmt, err := k.poolManagerKeeper.RouteExactAmountIn(
			ctx,
			buyer,
			feeToken.Route,
			sdk.NewCoin(baseDenom, pumpAmt), // token in
			math.ZeroInt(),                  // no slippage)
		)
		if err != nil {
			return fmt.Errorf("route exact amount in: target denom: %s, error: %w", targetDenom, err)
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

	return sdk.NewCoin(targetDenom, tokenOutAmt), err
}

// getFeeTokenForDenom gets the FeeToken configuration for a given denom
// TODO: This is a placeholder - the actual implementation should use the txfees module
func (k Keeper) getFeeTokenForDenom(ctx sdk.Context, denom string) (*FeeToken, error) {
	// TODO: Implement using txfees keeper to get FeeToken with Route
	// For now, return a mock structure
	return &FeeToken{
		Route: []poolmanagertypes.SwapAmountInRoute{
			{
				PoolId:        1, // TODO: get proper pool ID from txfees
				TokenOutDenom: denom,
			},
		},
	}, nil
}

// FeeToken represents the fee token configuration with routing information
type FeeToken struct {
	Route []poolmanagertypes.SwapAmountInRoute
}

// DistributePumpStreams processes all pump streams and executes pumps if conditions are met
func (k Keeper) DistributePumpStreams(ctx sdk.Context, pumpStreams []types.Stream) error {
	for _, stream := range pumpStreams {
		// Skip non-pump streams
		if stream.PumpParams == nil {
			continue
		}

		// Calculate epoch blocks for randomization
		epochBlocks, err := k.EpochBlocks(ctx, stream.DistrEpochIdentifier)
		if err != nil {
			return fmt.Errorf("failed to get epoch blocks: %w", err)
		}

		// Use ShouldPump directly to determine pump amount
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

		// Get top rollapps from the stream's distribution
		rollapps := k.TopRollapps(ctx, stream.DistributeTo.Records, stream.PumpParams.NumTopRollapps)

		// Distribute pump amount proportionally to each rollapp
		for _, rollapp := range rollapps {
			// Calculate proportional pump amount based on rollapp weight
			pumpAmtRA := pumpAmt.Mul(rollapp.Weight).Quo(stream.DistributeTo.TotalWeight)

			if pumpAmtRA.IsZero() {
				continue
			}

			err := k.ExecutePump(ctx, pumpAmtRA, rollapp.RollappID)
			if err != nil {
				k.Logger(ctx).Error("failed to execute pump", "streamID", stream.Id, "rollappID", rollapp.RollappID, "error", err)
				// Continue with other rollapps even if one fails
				continue
			}
		}

		// Update the stream's epoch budget left
		updatedStream := stream
		updatedStream.PumpParams.EpochBudgetLeft = updatedStream.PumpParams.EpochBudgetLeft.Sub(pumpAmt)

		err = k.SetStream(ctx, &updatedStream)
		if err != nil {
			return fmt.Errorf("failed to update stream after pump: %w", err)
		}
	}
	return nil
}

type RollappWeight struct {
	RollappID string
	Weight    math.Int
}

// TopRollapps selects nop N rollapps and returns their IDs.
func (k Keeper) TopRollapps(ctx sdk.Context, records []types.DistrRecord, topN uint32) []RollappWeight {
	// Filter out non-rollapp gauges
	var rollappRecords []RollappWeight
	for _, record := range records {
		gauge, err := k.ik.GetGaugeByID(ctx, record.GaugeId)
		if err != nil {
			k.Logger(ctx).Error("failed to get gauge", "gaugeID", record.GaugeId, "error", err)
			continue
		}
		if ra := gauge.GetRollapp(); ra != nil {
			rollappRecords = append(rollappRecords, RollappWeight{
				RollappID: ra.RollappId,
				Weight:    record.Weight,
			})
		}
	}

	// Sort all records which are rollapp gauges by weight in descending order
	sort.Slice(rollappRecords, func(i, j int) bool {
		return rollappRecords[i].Weight.GT(rollappRecords[j].Weight)
	})

	if len(rollappRecords) <= int(topN) {
		return rollappRecords
	}
	return rollappRecords[:topN]
}

// Number of seconds in the year.
// 60 * 60 * 8766 is how the SDK defines it:
// https://github.com/cosmos/cosmos-sdk/blob/v0.50.14/x/mint/types/params.go#L33
const year = 60 * 60 * 8766

func (k Keeper) EpochBlocks(ctx sdk.Context, epochID string) (uint64, error) {
	info := k.ek.GetEpochInfo(ctx, epochID)
	mintParams, err := k.mintParams.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("get mint params: %w", err)
	}
	blocksPerSecond := mintParams.BlocksPerYear / year
	// info.Duration might be "hour", "day", or "week" and is defined as
	// an integer, so it's safe to cast it to uint64.
	return uint64(info.Duration.Seconds()) * blocksPerSecond, nil
}
