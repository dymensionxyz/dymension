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
	{"plan", InvariantPlan},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantPlan(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		plans := k.GetAllPlans(ctx, false)
		if len(plans) == 0 {
			return nil
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

		lastPlanID := k.GetLastPlanId(ctx)
		if lastPlanID != plans[len(plans)-1].Id {
			return fmt.Errorf("last plan id mismatch: lastPlanID: %d, lastPlanInListID: %d", lastPlanID, plans[len(plans)-1].Id)
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
