package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Claim implements types.MsgServer.
func (m msgServer) Claim(ctx context.Context, req *types.MsgClaim) (*types.MsgClaimResponse, error) {
	err := m.Keeper.Claim(sdk.UnwrapSDKContext(ctx), req.PlanId, req.Claimer)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimResponse{}, nil
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
	if availableTokens.IsZero() {
		return nil
	}
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
