package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// TransfersEnabled called by the genesis transfer IBC module when a transfer is handled
func (k Keeper) TransfersEnabled(ctx sdk.Context, rollappId string) error {
	return k.Settle(ctx, rollappId)
}

// Settle settles the iro plan with the given rollappId
func (k Keeper) Settle(ctx sdk.Context, rollappId string) error {
	rollapp := k.rk.MustGetRollapp(ctx, rollappId)

	// Get the plan
	plan, found := k.GetPlanByRollapp(ctx, rollappId)
	if !found {
		return nil
	}

	if plan.IsSettled() {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInternal, types.ErrPlanSettled), "rollappId: %s", rollappId)
	}

	// Get the rollapp's denom
	rollappDenom, err := rollapp.GetIBCDenom()
	if err != nil {
		return err
	}

	// validate the required funds are available in the module account
	// funds expected as it's validated in the genesis transfer handler
	balance := k.bk.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), rollappDenom)
	if balance.Amount.LT(plan.TotalAllocation.Amount) {
		return errorsmod.Wrapf(gerrc.ErrInternal, "required: %s, available: %s", plan.TotalAllocation.String(), balance.String())
	}

	// "claims" the unsold FUT token
	futBalance := k.bk.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.TotalAllocation.Denom)
	err = k.bk.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(futBalance))
	if err != nil {
		return err
	}

	// mark the plan as `settled`, allowing users to claim tokens
	plan.SettledDenom = rollappDenom
	k.SetPlan(ctx, plan)

	// FIXME: uses the raised DYM and unsold tokens to bootstrap the rollapp's liquidity pool

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

// Claim claims the FUT token for the real RA token
func (k Keeper) Claim(ctx sdk.Context, planId, claimer string) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return errorsmod.Wrapf(types.ErrPlanNotFound, "planId: %s", planId)
	}

	if !plan.IsSettled() {
		return errorsmod.Wrapf(types.ErrPlanSettled, "planId: %s", planId)
	}

	// Burn all the FUT tokens the user have
	availableTokens := k.bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(claimer), plan.TotalAllocation.Denom)
	err := k.bk.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(claimer), types.ModuleName, sdk.NewCoins(availableTokens))
	if err != nil {
		return err
	}
	err = k.bk.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(availableTokens))
	if err != nil {
		return err
	}

	// Give the user the RA token in return (same amount as the FUT token)
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(claimer), sdk.NewCoins(sdk.NewCoin(params.BaseDenom, availableTokens.Amount)))
	if err != nil {
		return err
	}

	// Update the plan
	plan.ClaimedAmt = plan.ClaimedAmt.Add(availableTokens.Amount)
	k.SetPlan(ctx, plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventClaim{
		Claimer:   claimer,
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    availableTokens.Amount,
	})
	if err != nil {
		return err
	}

	return nil
}
