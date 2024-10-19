package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Buy implements types.MsgServer.
func (m msgServer) Buy(ctx context.Context, req *types.MsgBuy) (*types.MsgBuyResponse, error) {
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.Buy(sdk.UnwrapSDKContext(ctx), req.PlanId, buyer, req.Amount, req.MaxCostAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{}, nil
}

// BuyExactSpend implements types.MsgServer.
func (m msgServer) BuyExactSpend(ctx context.Context, req *types.MsgBuyExactSpend) (*types.MsgBuyResponse, error) {
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, err
	}

	err = m.Keeper.BuyExactSpend(sdk.UnwrapSDKContext(ctx), req.PlanId, buyer, req.Spend, req.MinOutTokensAmount)
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
	err = m.Keeper.Sell(sdk.UnwrapSDKContext(ctx), req.PlanId, seller, req.Amount, req.MinIncomeAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellResponse{}, nil
}

// Buy buys fixed amount of allocation with price according to the price curve
func (k Keeper) Buy(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountTokensToBuy, maxCostAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, buyer.String())
	if err != nil {
		return err
	}

	// validate the IRO have enough tokens to sell
	if plan.SoldAmt.Add(amountTokensToBuy).GT(plan.TotalAllocation.Amount) {
		return types.ErrInsufficientTokens
	}

	// Calculate costAmt for buying amountTokensToBuy over the price curve
	costAmt := plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(amountTokensToBuy))
	costPlusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(costAmt, k.GetParams(ctx).TakerFee, true)
	if err != nil {
		return err
	}

	// Validate expected out amount
	if costPlusTakerFeeAmt.GT(maxCostAmt) {
		return errorsmod.Wrapf(types.ErrInvalidExpectedOutAmount, "maxCost: %s, cost: %s, fee: %s", maxCostAmt.String(), costAmt.String(), takerFeeAmt.String())
	}

	// Charge taker fee
	takerFee := sdk.NewCoin(appparams.BaseDenom, takerFeeAmt)
	err = k.chargeTakerFee(ctx, takerFee, buyer)
	if err != nil {
		return err
	}

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	cost := sdk.NewCoin(appparams.BaseDenom, costAmt)
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(cost))
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
	k.SetPlan(ctx, *plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventBuy{
		Buyer:     buyer.String(),
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    amountTokensToBuy,
		Cost:      costAmt,
		TakerFee:  takerFeeAmt,
	})
	if err != nil {
		return err
	}

	return nil
}

// BuyExactSpend uses exact amount of DYM to buy tokens on the curve
func (k Keeper) BuyExactSpend(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountToSpend, minTokensAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, buyer.String())
	if err != nil {
		return err
	}

	// deduct taker fee from the amount to spend
	toSpendMinusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(amountToSpend, k.GetParams(ctx).TakerFee, false)
	if err != nil {
		return err
	}

	// calculate the amount of tokens possible to buy with the amount to spend
	tokensOutAmt := plan.BondingCurve.TokensForExactDYM(plan.SoldAmt, toSpendMinusTakerFeeAmt)

	// Validate expected out amount
	if tokensOutAmt.LT(minTokensAmt) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minTokens: %s, tokens: %s, fee: %s", minTokensAmt.String(), tokensOutAmt.String(), takerFeeAmt.String())
	}

	// validate the IRO have enough tokens to sell
	if plan.SoldAmt.Add(tokensOutAmt).GT(plan.TotalAllocation.Amount) {
		return types.ErrInsufficientTokens
	}

	// Charge taker fee
	takerFee := sdk.NewCoin(appparams.BaseDenom, takerFeeAmt)
	err = k.chargeTakerFee(ctx, takerFee, buyer)
	if err != nil {
		return err
	}

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	cost := sdk.NewCoin(appparams.BaseDenom, toSpendMinusTakerFeeAmt)
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(cost))
	if err != nil {
		return err
	}

	// send allocated tokens from the plan to the buyer
	err = k.BK.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, tokensOutAmt)))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Add(tokensOutAmt)
	k.SetPlan(ctx, *plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventBuy{
		Buyer:     buyer.String(),
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    tokensOutAmt,
		Cost:      toSpendMinusTakerFeeAmt,
		TakerFee:  takerFeeAmt,
	})
	if err != nil {
		return err
	}

	return nil
}

// Sell sells allocation with price according to the price curve
func (k Keeper) Sell(ctx sdk.Context, planId string, seller sdk.AccAddress, amountTokensToSell, minIncomeAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, seller.String())
	if err != nil {
		return err
	}

	// Calculate the value of the tokens to sell according to the price curve
	costAmt := plan.BondingCurve.Cost(plan.SoldAmt.Sub(amountTokensToSell), plan.SoldAmt)
	costMinusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(costAmt, k.GetParams(ctx).TakerFee, false)
	if err != nil {
		return err
	}

	// Validate expected out amount
	if costMinusTakerFeeAmt.LT(minIncomeAmt) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minCost: %s, cost: %s, fee: %s", minIncomeAmt.String(), costAmt.String(), takerFeeAmt.String())
	}

	// Charge taker fee
	takerFee := sdk.NewCoin(appparams.BaseDenom, takerFeeAmt)
	err = k.chargeTakerFee(ctx, takerFee, seller)
	if err != nil {
		return err
	}

	// send allocated tokens from seller to the plan
	err = k.BK.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	// send DYM from the plan to the seller. DYM managed by the plan's module account
	cost := sdk.NewCoin(appparams.BaseDenom, costAmt)
	err = k.BK.SendCoins(ctx, plan.GetAddress(), seller, sdk.NewCoins(cost))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Sub(amountTokensToSell)
	k.SetPlan(ctx, *plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventSell{
		Seller:    seller.String(),
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    amountTokensToSell,
		Revenue:   costAmt,
		TakerFee:  takerFeeAmt,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetTradeableIRO returns the tradeable IRO plan
// - plan must exist
// - plan must not be settled
// - plan must have started (unless the trader is the owner)
func (k Keeper) GetTradeableIRO(ctx sdk.Context, planId string, trader string) (*types.Plan, error) {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return nil, types.ErrPlanNotFound
	}

	if plan.IsSettled() {
		return nil, errorsmod.Wrapf(types.ErrPlanSettled, "planId: %d", plan.Id)
	}

	// Validate start time started (unless the trader is the owner)
	if ctx.BlockTime().Before(plan.StartTime) && k.rk.MustGetRollapp(ctx, plan.RollappId).Owner != trader {
		return nil, errorsmod.Wrapf(types.ErrPlanNotStarted, "planId: %d", plan.Id)
	}

	return &plan, nil
}

// chargeTakerFee charges taker fee from the sender
// takerFee sent to the txfees module
func (k Keeper) chargeTakerFee(ctx sdk.Context, takerFee sdk.Coin, sender sdk.AccAddress) error {
	return k.BK.SendCoinsFromAccountToModule(ctx, sender, txfeestypes.ModuleName, sdk.NewCoins(takerFee))
}

// ApplyTakerFee applies taker fee to the cost
// isAdd: true if adding fee to the cost, false if subtracting fee from the cost
// returns new cost and fee. both must be positive
func (k Keeper) ApplyTakerFee(amount math.Int, takerFee math.LegacyDec, isAdd bool) (totalAmt, takerFeeAmt math.Int, err error) {
	if !amount.IsPositive() {
		return math.Int{}, math.Int{}, errorsmod.Wrapf(types.ErrInvalidCost, "amt: %s", amount.String())
	}

	feeAmt := math.LegacyNewDecFromInt(amount).Mul(takerFee).TruncateInt()

	var newAmt math.Int
	if isAdd {
		newAmt = amount.Add(feeAmt)
	} else {
		newAmt = amount.Sub(feeAmt)
	}

	if !newAmt.IsPositive() || !feeAmt.IsPositive() {
		return math.Int{}, math.Int{}, errorsmod.Wrapf(types.ErrInvalidCost, "taking fee resulted in negative amount: %s, fee: %s", newAmt.String(), feeAmt.String())
	}

	return newAmt, feeAmt, nil
}
