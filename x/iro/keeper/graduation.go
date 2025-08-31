package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// createPoolForPlan creates a pool for the plan
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

// GraduatePlan graduates the plan into a pool
func (k Keeper) GraduatePlan(ctx sdk.Context, planId string) (uint64, sdk.Coins, error) {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return 0, nil, errorsmod.Wrapf(gerrc.ErrNotFound, "plan not found")
	}

	poolID, leftoverTokens, err := k.createPoolForPlan(ctx, plan)
	if err != nil {
		return 0, nil, errors.Join(types.ErrFailedBootstrapLiquidityPool, err)
	}

	// set the plan to the graduated status
	plan.GraduationStatus = types.GraduationStatus_POOL_CREATED
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
