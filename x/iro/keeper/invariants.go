package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/invar"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

var invs = invar.NamedFuncsList[Keeper]{
	{"plan", InvariantPlan},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantPlan(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		lastPlanID := k.GetLastPlanId(ctx)
		plans := k.GetAllPlans(ctx, false)
		if len(plans) == 0 {
			return nil, false
		}

		if lastPlanID != plans[len(plans)-1].Id {
			return fmt.Errorf("last plan id mismatch: lastPlanID: %d, lastPlanInListID: %d", lastPlanID, plans[len(plans)-1].Id), true
		}

		for _, plan := range plans {
			if err := plan.ValidateBasic(); err != nil {
				return fmt.Errorf("plan validate basic: planID: %d, err: %w", plan.Id, err), true
			}

			if plan.TotalAllocation.Amount.LT(plan.SoldAmt) {
				return fmt.Errorf("total allocation less than sold amount: planID: %d, totalAllocation: %s, soldAmt: %s", plan.Id, plan.TotalAllocation.Amount, plan.SoldAmt), true
			}

			if plan.TotalAllocation.Amount.LT(plan.ClaimedAmt) {
				return fmt.Errorf("total allocation less than claimed amount: planID: %d, totalAllocation: %s, claimedAmt: %s", plan.Id, plan.TotalAllocation.Amount, plan.ClaimedAmt), true
			}

			if plan.ClaimedAmt.GT(plan.SoldAmt) {
				return fmt.Errorf("claimed amount greater than sold amount: planID: %d, claimedAmt: %s, soldAmt: %s", plan.Id, plan.ClaimedAmt, plan.SoldAmt), true
			}
		}

		return nil, false
	}
}
