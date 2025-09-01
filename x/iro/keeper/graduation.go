package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
)

// GraduatePlan graduates the plan into a pool
// it's called once an IRO plan reaches its maximum selling amount
func (k Keeper) GraduatePlan(ctx sdk.Context, planId string) (uint64, sdk.Coins, error) {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return 0, nil, errorsmod.Wrapf(gerrc.ErrNotFound, "plan not found")
	}

	poolID, leftoverTokens, err := k.createPoolForPlan(ctx, plan)
	if err != nil {
		return 0, nil, errors.Join(types.ErrFailedBootstrapLiquidityPool, err)
	}

	// set the pool ID to the plan
	plan.GraduatedPoolId = poolID
	k.SetPlan(ctx, plan)

	rollapp, found := k.rk.GetRollapp(ctx, plan.RollappId)
	if !found {
		return 0, nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp not found")
	}
	// graduated plans can be launched, thus we need to update the pre launch time
	if rollapp.PreLaunchTime.After(ctx.BlockTime()) {
		k.rk.SetPreLaunchTime(ctx, &rollapp, ctx.BlockTime())
	}

	return poolID, leftoverTokens, nil
}

// createPoolForPlan creates a pool for the plan
// can be called both from graduation and settle
func (k Keeper) createPoolForPlan(ctx sdk.Context, plan types.Plan) (uint64, sdk.Coins, error) {
	raisedLiquidityAmt := k.BK.GetBalance(ctx, plan.GetAddress(), plan.LiquidityDenom).Amount
	poolTokens := raisedLiquidityAmt.ToLegacyDec().Mul(plan.LiquidityPart).TruncateInt()
	ownerTokens := raisedLiquidityAmt.Sub(poolTokens)

	// start the vesting schedule for the owner tokens
	plan.VestingPlan.Amount = ownerTokens
	plan.VestingPlan.StartTime = ctx.BlockHeader().Time.Add(plan.VestingPlan.StartTimeAfterSettlement)
	plan.VestingPlan.EndTime = plan.VestingPlan.StartTime.Add(plan.VestingPlan.VestingDuration)

	// uses the raised liquidity and unsold tokens to bootstrap the rollapp's liquidity pool
	return k.bootstrapLiquidityPool(ctx, plan, poolTokens)
}

// bootstrapLiquidityPool bootstraps the liquidity pool with the raised liquidity and unsold tokens.
//
// This function performs the following steps:
// - Sends the raised liquidity to the IRO module to be used as the pool creator.
// - Determines the required pool liquidity amounts to fulfill the last price.
// - Creates a balancer pool with the determined tokens and liquidity.
func (k Keeper) bootstrapLiquidityPool(ctx sdk.Context, plan types.Plan, poolTokens math.Int) (uint64, sdk.Coins, error) {
	// claimable amount is kept in the module account and used for user's claims
	claimableAmt := plan.SoldAmt.Sub(plan.ClaimedAmt)

	// the remaining tokens are used to bootstrap the liquidity pool
	unallocatedTokens := plan.TotalAllocation.Amount.Sub(claimableAmt)

	// send the raised liquidity token to the iro module as it will be used as the pool creator
	err := k.BK.SendCoinsFromAccountToModule(ctx, plan.GetAddress(), types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.LiquidityDenom, poolTokens)))
	if err != nil {
		return 0, nil, err
	}

	// if the plan is settled, use the settled denom
	// otherwise use the IRO denom
	denom := plan.SettledDenom
	if denom == "" {
		denom = plan.GetIRODenom()
	}

	// find the raTokens needed to bootstrap the pool, to fulfill last price
	raTokens, liquidityTokens := types.CalcLiquidityPoolTokens(unallocatedTokens, poolTokens, plan.SpotPrice())
	rollappLiquidityCoin := sdk.NewCoin(denom, raTokens)
	baseLiquidityCoin := sdk.NewCoin(plan.LiquidityDenom, liquidityTokens)

	// create pool
	gammGlobalParams := k.gk.GetParams(ctx).GlobalFees
	poolParams := balancer.NewPoolParams(gammGlobalParams.SwapFee, gammGlobalParams.ExitFee, nil)
	balancerPool := balancer.NewMsgCreateBalancerPool(k.AK.GetModuleAddress(types.ModuleName), poolParams, []balancer.PoolAsset{
		{
			Token:  baseLiquidityCoin,
			Weight: math.OneInt(),
		},
		{
			Token:  rollappLiquidityCoin,
			Weight: math.OneInt(),
		},
	}, "")

	// we call the pool manager directly, instead of the gamm keeper, to avoid the pool creation fee
	poolId, err := k.pm.CreatePool(ctx, balancerPool)
	if err != nil {
		return 0, nil, err
	}

	leftovers := sdk.NewCoins(
		sdk.NewCoin(baseLiquidityCoin.Denom, poolTokens.Sub(baseLiquidityCoin.Amount)),
		sdk.NewCoin(rollappLiquidityCoin.Denom, unallocatedTokens.Sub(rollappLiquidityCoin.Amount)),
	)

	return poolId, leftovers, nil
}
