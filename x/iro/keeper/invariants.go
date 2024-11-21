package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "plan", Func: InvariantPlan},
	{Name: "accounting", Func: InvariantAccounting},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

// the plan should validate and bookkeeping for tokens should be sensible
func InvariantPlan(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		plans := k.GetAllPlans(ctx, false)

		if len(plans) == 0 {
			return nil
		}

		lastPlanID := k.GetLastPlanId(ctx)
		if lastPlanID != plans[len(plans)-1].Id {
			return fmt.Errorf("last plan id mismatch: lastPlanID: %d, lastPlanInListID: %d", lastPlanID, plans[len(plans)-1].Id)
		}

		var errs []error
		for _, plan := range plans {
			err := checkPlan(plan)
			err = errorsmod.Wrapf(err, "planID: %d", plan.Id)
			errs = append(errs, err)
		}
		if err := errors.Join(errs...); err != nil {
			return errorsmod.Wrap(err, "check plans")
		}

		return nil
	})
}

func checkPlan(plan types.Plan) error {
	if err := plan.ValidateBasic(); err != nil {
		return fmt.Errorf("plan validate basic: planID: %d, err: %w", plan.Id, err)
	}

	if plan.TotalAllocation.Amount.LT(plan.SoldAmt) {
		return fmt.Errorf("total allocation less than sold amount: planID: %d, totalAllocation: %s, soldAmt: %s", plan.Id, plan.TotalAllocation.Amount, plan.SoldAmt)
	}

	if plan.TotalAllocation.Amount.LT(plan.ClaimedAmt) {
		return fmt.Errorf("total allocation less than claimed amount: planID: %d, totalAllocation: %s, claimedAmt: %s", plan.Id, plan.TotalAllocation.Amount, plan.ClaimedAmt)
	}

	if plan.ClaimedAmt.GT(plan.SoldAmt) {
		return fmt.Errorf("claimed amount greater than sold amount: planID: %d, claimedAmt: %s, soldAmt: %s", plan.Id, plan.ClaimedAmt, plan.SoldAmt)
	}
	return nil
}

/*
For all plans

	if plan is settled, no IRO tokens should be left
	if plan is settled, no DYM should be left in the plan account
	module should have enough RA tokens to cover the claimable amount
*/
func InvariantAccounting(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		plans := k.GetAllPlans(ctx, false)
		var errs []error

		for _, plan := range plans {
			if plan.IsSettled() {
				// Check if no IRO tokens are left
				iroBalance := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.GetIRODenom())
				if !iroBalance.IsZero() {
					errs = append(errs, fmt.Errorf("iro tokens left: planID: %d, balance: %s", plan.Id, iroBalance))
				}

				// Check if no DYM is left in the plan account
				dymBalance := k.BK.GetBalance(ctx, plan.GetAddress(), appparams.BaseDenom)
				if !dymBalance.IsZero() {
					errs = append(errs, fmt.Errorf("dym tokens left: planID: %d, balance: %s", plan.Id, dymBalance))
				}
			}

			// Check if module has enough RA tokens to cover the claimable amount
			claimableAmount := plan.TotalAllocation.Amount.Sub(plan.ClaimedAmt)
			moduleBalance := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.TotalAllocation.Denom)
			if moduleBalance.Amount.LT(claimableAmount) {
				errs = append(errs, fmt.Errorf("insufficient RA tokens: planID: %d, required: %s, available: %s",
					plan.Id, claimableAmount, moduleBalance.Amount))
			}
		}

		if err := errors.Join(errs...); err != nil {
			return errorsmod.Wrap(err, "accounting check")
		}

		return nil
	})
}
