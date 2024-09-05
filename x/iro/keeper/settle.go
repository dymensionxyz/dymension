package keeper

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	balancer "github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

// TransfersEnabled called by the genesis transfer IBC module when a transfer is handled
// This is a rollapp module hook
func (k Keeper) TransfersEnabled(ctx sdk.Context, rollappId, rollappIBCDenom string) error {
	return k.Settle(ctx, rollappId, rollappIBCDenom)
}

// Settle settles the iro plan with the given rollappId
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
	if balance.Amount.LT(plan.TotalAllocation.Amount) {
		return errorsmod.Wrapf(gerrc.ErrInternal, "required: %s, available: %s", plan.TotalAllocation.String(), balance.String())
	}

	// "claims" the unsold FUT token
	futBalance := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.TotalAllocation.Denom)
	err := k.BK.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(futBalance))
	if err != nil {
		return err
	}

	// mark the plan as `settled`, allowing users to claim tokens
	plan.SettledDenom = rollappIBCDenom
	k.SetPlan(ctx, plan)

	// uses the raised DYM and unsold tokens to bootstrap the rollapp's liquidity pool
	err = k.bootstrapLiquidityPool(ctx, plan)
	if err != nil {
		return errors.Join(types.ErrFailedBootstrapLiquidityPool, err)
	}

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventSettle{
		RollappId: rollappId,
		PlanId:    fmt.Sprintf("%d", plan.Id),
	})
	if err != nil {
		return err
	}

	return nil
}

// Bootstrap liquidity pool with the raised DYM and unsold tokens
func (k Keeper) bootstrapLiquidityPool(ctx sdk.Context, plan types.Plan) error {
	unallocatedTokens := plan.TotalAllocation.Amount.Sub(plan.SoldAmt)        // assumed > 0, as we enforce it in the Buy function
	raisedDYM := k.BK.GetBalance(ctx, plan.GetAddress(), appparams.BaseDenom) // assumed > 0, as we enforce it by IRO creation fee

	// send the raised DYM to the iro module
	err := k.BK.SendCoinsFromAccountToModule(ctx, plan.GetAddress(), types.ModuleName, sdk.NewCoins(raisedDYM))
	if err != nil {
		return err
	}

	// calculate last price
	lastPrice := plan.BondingCurve.SpotPrice(plan.SoldAmt)
	// find the tokens needed to bootstrap the pool, to fulfill last price
	tokens, dym := determineLimitingFactor(unallocatedTokens, raisedDYM.Amount, lastPrice)
	rollappLiquidityCoin := sdk.NewCoin(plan.SettledDenom, tokens)
	dymLiquidityCoin := sdk.NewCoin(appparams.BaseDenom, dym)

	// create pool
	gammGlobalParams := k.gk.GetParams(ctx).GlobalFees
	poolParams := balancer.NewPoolParams(gammGlobalParams.SwapFee, gammGlobalParams.ExitFee, nil)
	balancerPool := balancer.NewMsgCreateBalancerPool(k.AK.GetModuleAddress(types.ModuleName), poolParams, []balancer.PoolAsset{
		{
			Token:  dymLiquidityCoin,
			Weight: math.OneInt(),
		},
		{
			Token:  rollappLiquidityCoin,
			Weight: math.OneInt(),
		},
	}, "")

	poolId, err := k.gk.CreatePool(ctx, balancerPool) // TODO: skip the pool creation fee
	if err != nil {
		return err
	}

	// Add incentives
	poolDenom := gammtypes.GetPoolShareDenom(poolId)
	incentives := sdk.NewCoins(
		sdk.NewCoin(dymLiquidityCoin.Denom, raisedDYM.Amount.Sub(dymLiquidityCoin.Amount)),
		sdk.NewCoin(rollappLiquidityCoin.Denom, unallocatedTokens.Sub(rollappLiquidityCoin.Amount)),
	)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         poolDenom,
		Duration:      k.ik.GetLockableDurations(ctx)[0],
	}
	_, err = k.ik.CreateGauge(ctx, false, k.AK.GetModuleAddress(types.ModuleName), incentives, distrTo, ctx.BlockTime().Add(time.Hour), 30)
	if err != nil {
		return err
	}

	return nil
}

func determineLimitingFactor(unsoldTokens, raisedDYM math.Int, ratio math.LegacyDec) (tokens, dym math.Int) {
	requiredDYM := unsoldTokens.ToLegacyDec().Mul(ratio).TruncateInt()

	// if raisedDYM is less than requiredDYM, than DYM is the limiting factor
	// we use all the raisedDYM, and the corresponding amount of tokens
	if raisedDYM.LT(requiredDYM) {
		dym = raisedDYM
		tokens = raisedDYM.ToLegacyDec().Quo(ratio).TruncateInt()
		return
	}

	// if raisedDYM is more than requiredDYM, than tokens are the limiting factor
	// we use all the unsold tokens, and the corresponding amount of DYM
	tokens = unsoldTokens
	dym = requiredDYM

	// for the edge case where price is 0 (no tokens sold and initial price is 0)
	// and required DYM is 0, we set the raised DYM as the limiting factor
	if requiredDYM.IsZero() {
		dym = raisedDYM
	}
	return
}
