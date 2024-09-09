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

	err = m.Keeper.Buy(sdk.UnwrapSDKContext(ctx), req.PlanId, buyer, req.Amount, req.ExpectedOutAmount)
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
	err = m.Keeper.Sell(sdk.UnwrapSDKContext(ctx), req.PlanId, seller, req.Amount, req.ExpectedOutAmount)
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

	// validate the IRO have enough tokens to sell
	// protocol will apply max limit (99.9%) to enforce initial token liquidity
	maxSellAmt := plan.TotalAllocation.Amount.ToLegacyDec().Mul(AllocationSellLimit).TruncateInt()
	if plan.SoldAmt.Add(amountTokensToBuy).GT(maxSellAmt) {
		return types.ErrInsufficientTokens
	}

	// Calculate cost for buying amountTokensToBuy over fixed price curve
	cost := plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(amountTokensToBuy))
	if !cost.IsPositive() {
		return errorsmod.Wrapf(types.ErrInvalidCost, "cost: %s", cost.String())
	}
	costCoin := sdk.NewCoin(appparams.BaseDenom, cost)

	totalCost, takerFeeCoin := k.AddTakerFee(costCoin, k.GetParams(ctx).TakerFee)
	if !totalCost.IsPositive() || !takerFeeCoin.IsPositive() {
		return errorsmod.Wrapf(types.ErrInvalidCost, "totalCost: %s, takerFeeCoin: %s", totalCost.String(), takerFeeCoin.String())
	}

	// Validate expected out amount
	if totalCost.Amount.GT(maxCost) {
		return errorsmod.Wrapf(types.ErrInvalidExpectedOutAmount, "maxCost: %s, cost: %s", maxCost.String(), cost.String())
	}

	// Charge taker fee
	err = k.chargeTakerFee(ctx, takerFeeCoin, buyer)
	if err != nil {
		return err
	}

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(costCoin))
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
	if !cost.IsPositive() {
		return errorsmod.Wrapf(types.ErrInvalidCost, "cost: %s", cost.String())
	}
	costCoin := sdk.NewCoin(appparams.BaseDenom, cost)

	totalCost, takerFeeCoin := k.SubtractTakerFee(costCoin, k.GetParams(ctx).TakerFee)
	if !totalCost.IsPositive() || !takerFeeCoin.IsPositive() {
		return errorsmod.Wrapf(types.ErrInvalidCost, "totalCost: %s, takerFeeCoin: %s", totalCost.String(), takerFeeCoin.String())
	}

	// Validate expected out amount
	if totalCost.Amount.LT(minCost) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minCost: %s, cost: %s", minCost.String(), totalCost.String())
	}

	// Charge taker fee
	err = k.chargeTakerFee(ctx, takerFeeCoin, seller)
	if err != nil {
		return err
	}

	// send allocated tokens from seller to the plan
	err = k.BK.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	// send DYM from the plan to the seller. DYM managed by the plan's module account
	err = k.BK.SendCoins(ctx, plan.GetAddress(), seller, sdk.NewCoins(costCoin))
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

func (k Keeper) chargeTakerFee(ctx sdk.Context, takerFee sdk.Coin, sender sdk.AccAddress) error {
	return k.BK.SendCoinsFromAccountToModule(ctx, sender, txfeestypes.ModuleName, sdk.NewCoins(takerFee))
}

// AddTakerFee returns the remaining amount after subtracting the taker fee and the taker fee amount
// returns (1 + takerFee) * tokenIn, takerFee * tokenIn
func (k Keeper) AddTakerFee(amt sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	takerFeeAmt := math.LegacyNewDecFromInt(amt.Amount).Mul(takerFee).TruncateInt()
	newAmt := amt.Amount.Add(takerFeeAmt)
	return sdk.NewCoin(amt.Denom, newAmt), sdk.NewCoin(amt.Denom, takerFeeAmt)
}

// SubtractTakerFee returns the remaining amount after subtracting the taker fee and the taker fee amount
// returns (1 - takerFee) * tokenIn, takerFee * tokenIn
func (k Keeper) SubtractTakerFee(amt sdk.Coin, takerFee sdk.Dec) (sdk.Coin, sdk.Coin) {
	takerFeeAmt := math.LegacyNewDecFromInt(amt.Amount).Mul(takerFee).TruncateInt()
	newAmt := amt.Amount.Sub(takerFeeAmt)
	return sdk.NewCoin(amt.Denom, newAmt), sdk.NewCoin(amt.Denom, takerFeeAmt)
}
