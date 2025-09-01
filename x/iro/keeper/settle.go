package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// AfterTransfersEnabled called by the genesis transfer IBC module when a transfer is handled
// This is a rollapp module hook
func (k Keeper) AfterTransfersEnabled(ctx sdk.Context, rollappId, rollappIBCDenom string) error {
	return k.Settle(ctx, rollappId, rollappIBCDenom)
}

// Settle settles the iro plan with the given rollappId
//
// This function performs the following steps:
// - Validates that the "TotalAllocation.Amount" of the RA token are available in the module account.
// - Burns any unsold FUT tokens in the module account.
// - Marks the plan as settled, allowing users to claim tokens.
// - Starts the vesting schedule for the owner tokens.
// - Uses the raised liquidity and unsold tokens to bootstrap the rollapp's liquidity pool.
func (k Keeper) Settle(ctx sdk.Context, rollappId, rollappIBCDenom string) error {
	plan, found := k.GetPlanByRollapp(ctx, rollappId)
	if !found {
		return nil
	}

	if plan.IsSettled() {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInternal, types.ErrPlanSettled), "rollappId: %s", rollappId)
	}

	// validate the required funds are available in the module account
	// funds expected as it's validated in the genesis transfer handler
	balance := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), rollappIBCDenom)
	if !balance.Amount.Equal(plan.TotalAllocation.Amount) {
		return errorsmod.Wrapf(gerrc.ErrInternal, "required: %s, available: %s", plan.TotalAllocation.String(), balance.String())
	}

	// update the settled denom
	plan.SettledDenom = rollappIBCDenom

	var err error
	var poolID, gaugeID uint64
	// if already graduated, we need to swap the IRO asset tokens with the settled tokens
	if plan.IsGraduated() {
		poolID = plan.GraduatedPoolId
		// call SwapPoolAsset to swap the pool asset to the settled denom
		err := k.gk.ReplacePoolAsset(ctx, plan.GetAddress(), poolID, plan.GetIRODenom(), plan.SettledDenom)
		if err != nil {
			return err
		}
	} else {
		var incentives sdk.Coins
		poolID, incentives, err = k.createPoolForPlan(ctx, plan)
		if err != nil {
			return err
		}

		// add incentives to the pool
		gaugeID, err = k.addIncentivesToPool(ctx, plan, poolID, incentives)
		if err != nil {
			return errors.Join(types.ErrFailedBootstrapLiquidityPool, err)
		}

		plan.GraduatedPoolId = poolID
	}
	// commit the settled plan
	k.SetPlan(ctx, plan)

	// burn all the remaining IRO token.
	iroTokenBalance := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.TotalAllocation.Denom)
	err = k.BK.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(iroTokenBalance))
	if err != nil {
		return err
	}

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventSettle{
		PlanId:        fmt.Sprintf("%d", plan.Id),
		RollappId:     rollappId,
		IBCDenom:      rollappIBCDenom,
		PoolId:        poolID,
		GaugeId:       gaugeID,
		VestingAmount: plan.VestingPlan.Amount,
	})
	if err != nil {
		return err
	}

	return nil
}

// addIncentivesToPool adds incentives to the pool
func (k Keeper) addIncentivesToPool(ctx sdk.Context, plan types.Plan, poolID uint64, incentives sdk.Coins) (gaugeID uint64, err error) {
	poolDenom := gammtypes.GetPoolShareDenom(poolID)
	distrTo := lockuptypes.QueryCondition{
		Denom:    poolDenom,
		LockAge:  k.ik.GetParams(ctx).MinLockAge,
		Duration: k.ik.GetParams(ctx).MinLockDuration,
	}
	return k.ik.CreateAssetGauge(ctx,
		false,
		k.AK.GetModuleAddress(types.ModuleName),
		incentives,
		distrTo,
		ctx.BlockTime().Add(plan.IncentivePlanParams.StartTimeAfterSettlement),
		plan.IncentivePlanParams.NumEpochsPaidOver,
	)
}
