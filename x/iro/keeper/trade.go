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

var AllocationSellLimit = math.LegacyNewDecWithPrec(999, 3) // 99.9%

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
	// protocol will apply max limit (99.9%) to enforce initial token liquidity
	maxSellAmt := plan.TotalAllocation.Amount.ToLegacyDec().Mul(AllocationSellLimit).TruncateInt()
	if plan.SoldAmt.Add(amountTokensToBuy).GT(maxSellAmt) {
		return types.ErrInsufficientTokens
	}

	// Calculate cost for buying amountTokensToBuy over the price curve
	cost := plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(amountTokensToBuy))
	costPlusTakerFee, takerFee, err := k.ApplyTakerFee(cost, k.GetParams(ctx).TakerFee, true)
	if err != nil {
		return err
	}

	// Validate expected out amount
	if costPlusTakerFee.GT(maxCostAmt) {
		return errorsmod.Wrapf(types.ErrInvalidExpectedOutAmount, "maxCost: %s, cost: %s, fee: %s", maxCostAmt.String(), cost.String(), takerFee.String())
	}

	// Charge taker fee
	takerFeeC := sdk.NewCoin(appparams.BaseDenom, takerFee)
	err = k.chargeTakerFee(ctx, takerFeeC, buyer)
	if err != nil {
		return err
	}

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	costC := sdk.NewCoin(appparams.BaseDenom, cost)
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
	k.SetPlan(ctx, *plan)

	// Emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventBuy{
		Buyer:     buyer.String(),
		PlanId:    planId,
		RollappId: plan.RollappId,
		Amount:    amountTokensToBuy,
		Cost:      cost,
		TakerFee:  takerFee,
	})
	if err != nil {
		return err
	}

	return nil
}

// BuyExactSpend uses fixed amount of DYM to buy as many tokens as possible
func (k Keeper) BuyExactSpend(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountToSpend, minTokensAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, buyer.String())
	if err != nil {
		return err
	}

	// deduct taker fee from the amount to spend
	toSpendLessTakerFee, takerFee, err := k.ApplyTakerFee(amountToSpend, k.GetParams(ctx).TakerFee, false)
	if err != nil {
		return err
	}

	// calculate the amount of tokens possible to buy with the amount to spend
	tokensOutAmt := plan.BondingCurve.TokensForExactDYM(plan.SoldAmt, toSpendLessTakerFee)

	// Validate expected out amount
	if tokensOutAmt.LT(minTokensAmt) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minTokens: %s, tokens: %s, fee: %s", minTokensAmt.String(), tokensOutAmt.String(), takerFee.String())
	}

	// validate the IRO have enough tokens to sell
	// protocol will apply max limit (99.9%) to enforce initial token liquidity
	maxSellAmt := plan.TotalAllocation.Amount.ToLegacyDec().Mul(AllocationSellLimit).TruncateInt()
	if plan.SoldAmt.Add(tokensOutAmt).GT(maxSellAmt) {
		return types.ErrInsufficientTokens
	}

	// Charge taker fee
	takerFeeC := sdk.NewCoin(appparams.BaseDenom, takerFee)
	err = k.chargeTakerFee(ctx, takerFeeC, buyer)
	if err != nil {
		return err
	}

	// send DYM from buyer to the plan. DYM sent directly to the plan's module account
	costC := sdk.NewCoin(appparams.BaseDenom, toSpendLessTakerFee)
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(costC))
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
		Cost:      toSpendLessTakerFee,
		TakerFee:  takerFee,
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
	cost := plan.BondingCurve.Cost(plan.SoldAmt.Sub(amountTokensToSell), plan.SoldAmt)
	costMinusTakerFee, takerFee, err := k.ApplyTakerFee(cost, k.GetParams(ctx).TakerFee, false)
	if err != nil {
		return err
	}

	// Validate expected out amount
	if costMinusTakerFee.LT(minIncomeAmt) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minCost: %s, cost: %s, fee: %s", minIncomeAmt.String(), cost.String(), takerFee.String())
	}

	// Charge taker fee
	takerFeeC := sdk.NewCoin(appparams.BaseDenom, takerFee)
	err = k.chargeTakerFee(ctx, takerFeeC, seller)
	if err != nil {
		return err
	}

	// send allocated tokens from seller to the plan
	err = k.BK.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	// send DYM from the plan to the seller. DYM managed by the plan's module account
	costC := sdk.NewCoin(appparams.BaseDenom, cost)
	err = k.BK.SendCoins(ctx, plan.GetAddress(), seller, sdk.NewCoins(costC))
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
		Revenue:   cost,
		TakerFee:  takerFee,
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
