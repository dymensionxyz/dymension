package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Settle settles the iro plan with the given rollappId
func (k Keeper) Settle(ctx sdk.Context, rollappId string) error {
	/*
			validate the required funds are available in the module account //funds validated by the `genesistransfer` handler
			- rollapp token is known on the rollapp object
		* "claims" the unsold FUT token
		* move the funds to the plan's module account
		* mark the plan as `settled`, allowing users to claim tokens

		* uses the raised DYM and unsold tokens to bootstrap the rollapp's liquidity pool
	*/
	return nil
}

// Claim claims the FUT token for the real RA token
func (k Keeper) Claim(ctx sdk.Context, planId, claimer string) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return errorsmod.Wrapf(types.ErrPlanNotFound, "planId: %s", planId)
	}

	if !plan.Settled {
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
