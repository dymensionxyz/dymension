package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Claim implements types.MsgServer.
func (m msgServer) Claim(ctx context.Context, req *types.MsgClaim) (*types.MsgClaimResponse, error) {
	claimerAddr := sdk.MustAccAddressFromBech32(req.Claimer)
	err := m.Keeper.Claim(sdk.UnwrapSDKContext(ctx), req.PlanId, claimerAddr)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimResponse{}, nil
}

// Claim claims the FUT token for the real RA token
//
// This function allows a user to claim their RA tokens by burning their FUT tokens.
// It burns *all* the FUT tokens the claimer has, and sends the equivalent amount of RA tokens to the claimer.
func (k Keeper) Claim(ctx sdk.Context, planId string, claimer sdk.AccAddress) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return types.ErrPlanNotFound
	}

	if !plan.IsSettled() {
		return types.ErrPlanNotSettled
	}

	availableTokens := k.BK.GetBalance(ctx, claimer, plan.TotalAllocation.Denom)
	if availableTokens.IsZero() {
		return types.ErrNoTokensToClaim
	}

	// Burn all the FUT tokens the user have
	err := k.BK.SendCoinsFromAccountToModule(ctx, claimer, types.ModuleName, sdk.NewCoins(availableTokens))
	if err != nil {
		return err
	}
	err = k.BK.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(availableTokens))
	if err != nil {
		return err
	}

	// Give the user the RA token in return (same amount as the FUT token)
	err = k.BK.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimer, sdk.NewCoins(sdk.NewCoin(plan.SettledDenom, availableTokens.Amount)))
	if err != nil {
		return err
	}

	// Update the plan
	plan.ClaimedAmt = plan.ClaimedAmt.Add(availableTokens.Amount)
	k.SetPlan(ctx, plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventClaim{
		Claimer:   claimer.String(),
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    availableTokens.Amount,
	})
	if err != nil {
		return err
	}

	return nil
}
