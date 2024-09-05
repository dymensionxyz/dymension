package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

var AllocationSellLimit = math.LegacyNewDecWithPrec(999, 3) // 99.9%

// Buy implements types.MsgServer.
func (m msgServer) Buy(ctx context.Context, req *types.MsgBuy) (*types.MsgBuyResponse, error) {
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.Buy(sdk.UnwrapSDKContext(ctx), req.PlanId, buyer, req.Amount.Amount, req.ExpectedOutAmount.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{}, nil
}

// Sell implements types.MsgServer.
func (m msgServer) Sell(ctx context.Context, req *types.MsgSell) (*types.MsgSellResponse, error) {
	seller, err := sdk.AccAddressFromBech32(req.Seller)
	if err != nil {
		return nil, err
	}
	err = m.Keeper.Sell(sdk.UnwrapSDKContext(ctx), req.PlanId, seller, req.Amount.Amount, req.ExpectedOutAmount.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellResponse{}, nil
}

// Buy buys allocation with price according to the price curve
func (k Keeper) Buy(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountTokensToBuy, maxCost math.Int) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return types.ErrPlanNotFound
	}

	err := k.validateIROTradeable(ctx, plan, buyer.String())
	if err != nil {
		return err
	}

	// validate we have enough tokens to sell
	// protocol will apply max limit (99.9%) to enforce initial token liquidity
	maxSellAmt := plan.TotalAllocation.Amount.ToLegacyDec().Mul(AllocationSellLimit).TruncateInt()
	if plan.SoldAmt.Add(amountTokensToBuy).GT(maxSellAmt) {
		return types.ErrInsufficientTokens
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

	// Charge taker fee
	costC, err := k.chargeTakerFee(ctx, sdk.NewCoin(appparams.BaseDenom, cost), buyer)
	if err != nil {
		return err
	}

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(costC))
	if err != nil {
		return err
	}

	// send allocated tokens from the plan to the buyer
	err = k.BK.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToBuy)))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Add(amountTokensToBuy)
	k.SetPlan(ctx, plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventBuy{
		Buyer:     buyer.String(),
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
func (k Keeper) Sell(ctx sdk.Context, planId string, seller sdk.AccAddress, amountTokensToSell, minCost math.Int) error {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return errorsmod.Wrapf(types.ErrPlanNotFound, "planId: %s", planId)
	}

	err := k.validateIROTradeable(ctx, plan, seller.String())
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
	err = k.BK.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	// Charge taker fee
	costC, err := k.chargeTakerFee(ctx, sdk.NewCoin(appparams.BaseDenom, cost), seller)
	if err != nil {
		return err
	}

	// send DYM from the plan to the seller. DYM managed by the plan's module account
	err = k.BK.SendCoins(ctx, plan.GetAddress(), seller, sdk.NewCoins(costC))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Sub(amountTokensToSell)
	k.SetPlan(ctx, plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventSell{
		Seller:    seller.String(),
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    amountTokensToSell,
	})
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) validateIROTradeable(ctx sdk.Context, plan types.Plan, trader string) error {
	if plan.IsSettled() {
		return errorsmod.Wrapf(types.ErrPlanSettled, "planId: %d", plan.Id)
	}

	// Validate start time started (unless the trader is the owner)
	if ctx.BlockTime().Before(plan.StartTime) && k.rk.MustGetRollapp(ctx, plan.RollappId).Owner != trader {
		return errorsmod.Wrapf(types.ErrPlanNotStarted, "planId: %d", plan.Id)
	}

	return nil
}

func (k Keeper) chargeTakerFee(ctx sdk.Context, cost sdk.Coin, sender sdk.AccAddress) (sdk.Coin, error) {
	newAmt, takerFeeCoin := k.calcTakerFee(cost, k.GetParams(ctx).TakerFee)
	if newAmt.IsZero() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidCost, "no tokens left after taker fee")
	}

	if takerFeeCoin.IsZero() {
		return sdk.Coin{}, errorsmod.Wrapf(types.ErrInvalidCost, "taker fee is zero")
	}

	err := k.BK.SendCoinsFromAccountToModule(ctx, sender, txfeestypes.ModuleName, sdk.NewCoins(takerFeeCoin))
	if err != nil {
		return sdk.Coin{}, err
	}

	return newAmt, nil
}

// Returns remaining amount in to swap, and takerFeeCoins.
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func (k Keeper) calcTakerFee(amt sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	newAmt := math.LegacyNewDecFromInt(amt.Amount).MulTruncate(sdk.OneDec().Sub(takerFee)).TruncateInt()
	takerFeeAmt := amt.Amount.Sub(newAmt)
	return sdk.NewCoin(amt.Denom, newAmt), sdk.NewCoin(amt.Denom, takerFeeAmt)
}
