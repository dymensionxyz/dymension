package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// the plan should validate and struct level bookkeeping for tokens should be sensible
func InvariantPlan(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		plans := k.GetAllPlans(ctx, false)

		if len(plans) == 0 {
			return nil
		}

		lastPlanID := k.GetLastPlanId(ctx)
		max_ := plans[0].Id
		for _, plan := range plans {
			max_ = max(plan.Id, max_)
		}
		if lastPlanID != max_ {
			return fmt.Errorf("last plan id mismatch: lastPlanID: %d, max: %d", lastPlanID, max_)
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

// the plan and module accounts should have sufficient and correct balances
func InvariantAccounting(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		plans := k.GetAllPlans(ctx, false)
		var errs []error

		for _, plan := range plans {
			if plan.IsSettled() {
				// module should have no more IRO
				iroBalance := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.GetIRODenom())
				if !iroBalance.IsZero() {
					errs = append(errs, fmt.Errorf("iro tokens left in module, settled: planID: %d, balance: %s", plan.Id, iroBalance))
				}

				// Check if module has enough RA tokens to cover the claimable amount
				claimable := plan.SoldAmt.Sub(plan.ClaimedAmt)
				moduleBal := k.BK.GetBalance(ctx, k.AK.GetModuleAddress(types.ModuleName), plan.SettledDenom)
				if moduleBal.Amount.LT(claimable) {
					errs = append(errs, fmt.Errorf("insufficient RA tokens: planID: %d, required: %s, available: %s",
						plan.Id, claimable, moduleBal.Amount))
				}
			}
		}

		return errors.Join(errs...)
	})
}
