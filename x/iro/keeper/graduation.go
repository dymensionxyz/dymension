package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GraduatePlan graduates the plan into a pool
func (k Keeper) GraduatePlan(ctx sdk.Context, planId string) (uint64, error) {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return 0, errorsmod.Wrapf(gerrc.ErrNotFound, "plan not found")
	}

	// FIXME: check that the plan is in the pre-graduation status

	raisedLiquidityAmt := k.BK.GetBalance(ctx, plan.GetAddress(), plan.LiquidityDenom).Amount
	poolTokens := raisedLiquidityAmt.ToLegacyDec().Mul(plan.LiquidityPart).TruncateInt()
	ownerTokens := raisedLiquidityAmt.Sub(poolTokens)

	// start the vesting schedule for the owner tokens
	plan.VestingPlan.Amount = ownerTokens
	plan.VestingPlan.StartTime = ctx.BlockHeader().Time.Add(plan.VestingPlan.StartTimeAfterSettlement)
	plan.VestingPlan.EndTime = plan.VestingPlan.StartTime.Add(plan.VestingPlan.VestingDuration)

	// uses the raised liquidity and unsold tokens to bootstrap the rollapp's liquidity pool
	poolID, err := k.bootstrapLiquidityPool(ctx, plan, poolTokens)
	if err != nil {
		return 0, errors.Join(types.ErrFailedBootstrapLiquidityPool, err)
	}

	// set the plan to the graduated status
	plan.GraduationStatus = types.GraduationStatus_POOL_CREATED
	plan.GraduatedPoolId = poolID

	k.SetPlan(ctx, plan)

	return poolID, nil
}
