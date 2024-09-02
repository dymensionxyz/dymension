package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Buy implements types.MsgServer.
func (m msgServer) Buy(ctx context.Context, req *types.MsgBuy) (*types.MsgBuyResponse, error) {
	err := m.Keeper.Buy(sdk.UnwrapSDKContext(ctx), req.PlanId, req.Buyer, req.Amount.Amount, req.ExpectedOutAmount.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{}, nil
}

// Sell implements types.MsgServer.
func (m msgServer) Sell(ctx context.Context, req *types.MsgSell) (*types.MsgSellResponse, error) {
	err := m.Keeper.Sell(sdk.UnwrapSDKContext(ctx), req.PlanId, req.Seller, req.Amount.Amount, req.ExpectedOutAmount.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellResponse{}, nil
}

// Buy buys allocation with price according to the price curve
func (k Keeper) Buy(ctx sdk.Context, planId, buyer string, amountTokensToBuy, maxCost math.Int) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return errorsmod.Wrapf(types.ErrPlanNotFound, "planId: %s", planId)
	}

	err := k.validateTradeable(ctx, plan, buyer)
	if err != nil {
		return err
	}

	// Calculate cost over fixed price curve
	cost := plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(amountTokensToBuy))

	// validate cost is positive
	if !cost.IsPositive() {
		return errorsmod.Wrapf(types.ErrInvalidCost, "cost: %s", cost.String())
	}

	// Validate expected out amount
	if cost.GT(maxCost) {
		return errorsmod.Wrapf(types.ErrInvalidExpectedOutAmount, "maxCost: %s, cost: %s", maxCost.String(), cost.String())
	}

	//FIXME: Charge taker fee

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	err = k.bk.SendCoins(ctx, sdk.MustAccAddressFromBech32(buyer), plan.GetAddress(), sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, cost)))
	if err != nil {
		return err
	}

	// send allocated tokens from the plan to the buyer
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(buyer), sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToBuy)))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Add(amountTokensToBuy)
	k.SetPlan(ctx, plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventBuy{
		Buyer:     buyer,
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    amountTokensToBuy,
	})
	if err != nil {
		return err
	}

	return nil
}

// Sell sells allocation with price according to the price curve
func (k Keeper) Sell(ctx sdk.Context, planId, seller string, amountTokensToSell, minCost math.Int) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return errorsmod.Wrapf(types.ErrPlanNotFound, "planId: %s", planId)
	}

	err := k.validateTradeable(ctx, plan, seller)
	if err != nil {
		return err
	}

	// Calculate cost over fixed price curve
	cost := plan.BondingCurve.Cost(plan.SoldAmt.Sub(amountTokensToSell), plan.SoldAmt)
	// Validate expected out amount
	if cost.LT(minCost) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minCost: %s, cost: %s", minCost.String(), cost.String())
	}

	// send allocated tokens from seller to the plan
	err = k.bk.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(seller), types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	//FIXME: Charge taker fee

	// send DYM from the plan to the seller. DYM managed by the plan's module account
	err = k.bk.SendCoins(ctx, plan.GetAddress(), sdk.MustAccAddressFromBech32(seller), sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, cost)))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Sub(amountTokensToSell)
	k.SetPlan(ctx, plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventSell{
		Seller:    seller,
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    amountTokensToSell,
	})
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) validateTradeable(ctx sdk.Context, plan types.Plan, trader string) error {
	if plan.IsSettled() {
		return errorsmod.Wrapf(types.ErrPlanSettled, "planId: %d", plan.Id)
	}

	// Validate start time started (unless the trader is the owner)
	if ctx.BlockTime().Before(plan.StartTime) && k.rk.MustGetRollapp(ctx, plan.RollappId).Owner != trader {
		return errorsmod.Wrapf(types.ErrPlanNotStarted, "planId: %d", plan.Id)
	}

	return nil
}
