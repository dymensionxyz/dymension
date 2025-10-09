package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// EnableTrading enables trading for a given plan.
// It checks that the plan exists, it is not already enabled, the submitter is the owner of the RollApp
// and the plan is not settled.
// If all preconditions are met, it sets the TradingEnabled flag to true and stores the plan back in the
// store.
func (k Keeper) EnableTrading(ctx sdk.Context, planId string, submitter sdk.AccAddress) error {
	plan, ok := k.GetPlan(ctx, planId)
	if !ok {
		return types.ErrPlanNotFound
	}

	if plan.TradingEnabled {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "trading already enabled")
	}

	rollapp, found := k.rk.GetRollapp(ctx, plan.RollappId)
	if !found {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "rollapp not found")
	}

	owner := sdk.MustAccAddressFromBech32(rollapp.Owner)
	if !owner.Equals(submitter) {
		return errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the RollApp")
	}

	if plan.IsSettled() {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "plan already settled")
	}

	plan.EnableTradingWithStartTime(ctx.BlockTime())
	k.SetPlan(ctx, plan)

	// non standard launched plans need to set the pre launch time
	// standard launched plans will allow launch once graduated
	if !plan.StandardLaunch {
		k.rk.SetPreLaunchTime(ctx, &rollapp, plan.StartTime.Add(plan.IroPlanDuration))
	}

	err := uevent.EmitTypedEvent(ctx, &types.EventTradingEnabled{
		PlanId:    planId,
		RollappId: plan.RollappId,
	})
	if err != nil {
		return err
	}

	return nil
}

// Buy buys fixed amount of allocation with price according to the price curve
func (k Keeper) Buy(
	ctx sdk.Context,
	planId string,
	buyer sdk.AccAddress,
	amountTokensToBuy, maxCostAmt math.Int,
) (tokensInAmt math.Int, err error) {
	params := k.GetParams(ctx)

	plan, err := k.GetTradeableIRO(ctx, planId, buyer)
	if err != nil {
		return math.ZeroInt(), err
	}

	// validate the IRO have enough tokens to sell
	if plan.SoldAmt.Add(amountTokensToBuy).GT(plan.MaxAmountToSell) {
		return math.ZeroInt(), types.ErrInsufficientTokens
	}

	// Calculate costAmt for buying amountTokensToBuy over the price curve
	costAmt := plan.BondingCurve.Cost(plan.SoldAmt, plan.SoldAmt.Add(amountTokensToBuy))
	costPlusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(costAmt, params.TakerFee, true)
	if err != nil {
		return math.ZeroInt(), err
	}
	cost := sdk.NewCoin(plan.LiquidityDenom, costAmt)
	takerFee := sdk.NewCoin(plan.LiquidityDenom, takerFeeAmt)

	// Validate expected out amount
	if costPlusTakerFeeAmt.GT(maxCostAmt) {
		return math.ZeroInt(), errorsmod.Wrapf(types.ErrInvalidExpectedOutAmount, "maxCost: %s, cost: %s, fee: %s", maxCostAmt.String(), costAmt.String(), takerFeeAmt.String())
	}

	// validate minimal trade amount
	tradeAmtBaseDenom, err := k.tk.CalcCoinInBaseDenom(ctx, sdk.NewCoin(plan.LiquidityDenom, costAmt))
	if err != nil {
		return math.ZeroInt(), errorsmod.Wrapf(err, "failed to convert trade amount to base denom")
	}
	if !tradeAmtBaseDenom.Amount.GTE(params.MinTradeAmount) {
		return math.ZeroInt(), types.ErrInsufficientTradeAmount
	}

	// validate the remaining tokens have a positive cost
	newSoldAmt := plan.SoldAmt.Add(amountTokensToBuy)
	remainingTokens := plan.MaxAmountToSell.Sub(newSoldAmt)
	if remainingTokens.IsPositive() && !plan.BondingCurve.Cost(newSoldAmt, plan.MaxAmountToSell).IsPositive() {
		return math.ZeroInt(), errorsmod.Wrapf(types.ErrInvalidCost, "remaining tokens would not be buyable: remaining=%s", remainingTokens.String())
	}

	// Charge taker fee
	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)
	err = k.chargeTakerFee(ctx, takerFee, buyer, &owner)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Send liquidity token from buyer to the plan. The liquidity token sent directly to the plan's module account
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(cost))
	if err != nil {
		return math.ZeroInt(), err
	}

	// send allocated tokens from the plan to the buyer
	err = k.BK.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToBuy)))
	if err != nil {
		return math.ZeroInt(), err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Add(amountTokensToBuy)
	k.SetPlan(ctx, *plan)

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventBuy{
		Buyer:        buyer.String(),
		PlanId:       planId,
		RollappId:    plan.RollappId,
		Amount:       sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToBuy),
		Cost:         sdk.NewCoin(plan.LiquidityDenom, costAmt),
		TakerFee:     takerFee,
		ClosingPrice: plan.SpotPrice(),
	})
	if err != nil {
		return math.ZeroInt(), err
	}

	// if all tokens are sold, we need to graduate the plan
	if plan.SoldAmt.Equal(plan.MaxAmountToSell) {
		// we make the graduation gas free, as it's not relevant to the specific user's action
		noGasCtx := ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
		poolID, _, err := k.GraduatePlan(noGasCtx, planId)
		if err != nil {
			return math.ZeroInt(), err
		}

		err = uevent.EmitTypedEvent(noGasCtx, &types.EventGraduation{
			PlanId:    planId,
			RollappId: plan.RollappId,
			PoolId:    poolID,
		})
		if err != nil {
			return math.ZeroInt(), err
		}
	}

	return costPlusTakerFeeAmt, nil
}

// BuyExactSpend uses exact amount of liquidity to buy tokens on the curve
func (k Keeper) BuyExactSpend(
	ctx sdk.Context,
	planId string,
	buyer sdk.AccAddress,
	amountToSpend, minTokensAmt math.Int,
) (tokensOutAmt math.Int, err error) {
	params := k.GetParams(ctx)

	plan, err := k.GetTradeableIRO(ctx, planId, buyer)
	if err != nil {
		return math.ZeroInt(), err
	}

	// deduct taker fee from the amount to spend
	toSpendMinusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(amountToSpend, params.TakerFee, false)
	if err != nil {
		return math.ZeroInt(), err
	}
	takerFee := sdk.NewCoin(plan.LiquidityDenom, takerFeeAmt)
	cost := sdk.NewCoin(plan.LiquidityDenom, toSpendMinusTakerFeeAmt)

	// validate minimal trade amount
	tradeAmtBaseDenom, err := k.tk.CalcCoinInBaseDenom(ctx, sdk.NewCoin(plan.LiquidityDenom, toSpendMinusTakerFeeAmt))
	if err != nil {
		return math.ZeroInt(), errorsmod.Wrapf(err, "failed to convert trade amount to base denom")
	}
	if !tradeAmtBaseDenom.Amount.GTE(params.MinTradeAmount) {
		return math.ZeroInt(), types.ErrInsufficientTradeAmount
	}

	// calculate the amount of tokens possible to buy with the amount to spend
	tokensOutAmt, err = plan.BondingCurve.TokensForExactInAmount(plan.SoldAmt, toSpendMinusTakerFeeAmt)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Validate expected out amount
	if tokensOutAmt.LT(minTokensAmt) {
		return math.ZeroInt(), errorsmod.Wrapf(types.ErrInvalidMinCost, "minTokens: %s, tokens: %s, fee: %s", minTokensAmt.String(), tokensOutAmt.String(), takerFeeAmt.String())
	}

	// validate the IRO have enough tokens to sell
	if plan.SoldAmt.Add(tokensOutAmt).GT(plan.MaxAmountToSell) {
		return math.ZeroInt(), types.ErrInsufficientTokens
	}

	// validate the remaining tokens have a positive cost
	newSoldAmt := plan.SoldAmt.Add(tokensOutAmt)
	remainingTokens := plan.MaxAmountToSell.Sub(newSoldAmt)
	if remainingTokens.IsPositive() && !plan.BondingCurve.Cost(newSoldAmt, plan.MaxAmountToSell).IsPositive() {
		return math.ZeroInt(), errorsmod.Wrapf(types.ErrInvalidCost, "remaining tokens would not be buyable: remaining=%s", remainingTokens.String())
	}

	// Charge taker fee
	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)
	err = k.chargeTakerFee(ctx, takerFee, buyer, &owner)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Send liquidity token from buyer to the plan. The liquidity token sent directly to the plan's module account
	err = k.BK.SendCoins(ctx, buyer, plan.GetAddress(), sdk.NewCoins(cost))
	if err != nil {
		return math.ZeroInt(), err
	}

	// send allocated tokens from the plan to the buyer
	err = k.BK.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyer, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, tokensOutAmt)))
	if err != nil {
		return math.ZeroInt(), err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Add(tokensOutAmt)
	k.SetPlan(ctx, *plan)

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventBuy{
		Buyer:        buyer.String(),
		PlanId:       planId,
		RollappId:    plan.RollappId,
		Amount:       sdk.NewCoin(plan.TotalAllocation.Denom, tokensOutAmt),
		Cost:         sdk.NewCoin(plan.LiquidityDenom, toSpendMinusTakerFeeAmt),
		TakerFee:     takerFee,
		ClosingPrice: plan.SpotPrice(),
	})
	if err != nil {
		return math.ZeroInt(), err
	}

	return tokensOutAmt, nil
}

// Sell sells allocation with price according to the price curve
func (k Keeper) Sell(ctx sdk.Context, planId string, seller sdk.AccAddress, amountTokensToSell, minIncomeAmt math.Int) error {
	params := k.GetParams(ctx)

	plan, err := k.GetTradeableIRO(ctx, planId, seller)
	if err != nil {
		return err
	}

	// Calculate the value of the tokens to sell according to the price curve
	costAmt := plan.BondingCurve.Cost(plan.SoldAmt.Sub(amountTokensToSell), plan.SoldAmt)
	costMinusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(costAmt, params.TakerFee, false)
	if err != nil {
		return err
	}
	cost := sdk.NewCoin(plan.LiquidityDenom, costAmt)
	takerFee := sdk.NewCoin(plan.LiquidityDenom, takerFeeAmt)

	// validate minimal trade amount
	tradeAmtBaseDenom, err := k.tk.CalcCoinInBaseDenom(ctx, sdk.NewCoin(plan.LiquidityDenom, costAmt))
	if err != nil {
		return errorsmod.Wrapf(err, "failed to convert trade amount to base denom")
	}
	if !tradeAmtBaseDenom.Amount.GTE(params.MinTradeAmount) {
		return types.ErrInsufficientTradeAmount
	}

	// Validate expected out amount
	if costMinusTakerFeeAmt.LT(minIncomeAmt) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minCost: %s, cost: %s, fee: %s", minIncomeAmt.String(), costAmt.String(), takerFeeAmt.String())
	}

	// send allocated tokens from seller to the plan
	err = k.BK.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	// Send liquidity token from the plan to the seller. The liquidity token managed by the plan's module account

	err = k.BK.SendCoins(ctx, plan.GetAddress(), seller, sdk.NewCoins(cost))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Sub(amountTokensToSell)
	k.SetPlan(ctx, *plan)

	// Charge taker fee
	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)
	err = k.chargeTakerFee(ctx, takerFee, seller, &owner)
	if err != nil {
		return err
	}

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventSell{
		Seller:       seller.String(),
		PlanId:       planId,
		RollappId:    plan.RollappId,
		Amount:       sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell),
		Revenue:      sdk.NewCoin(plan.LiquidityDenom, costAmt),
		TakerFee:     takerFee,
		ClosingPrice: plan.SpotPrice(),
	})
	if err != nil {
		return err
	}

	return nil
}

// GetTradeableIRO returns the tradeable IRO plan
// - plan must exist
// - plan must not be graduated or settled
// - plan must have started (unless the trader is the owner)
func (k Keeper) GetTradeableIRO(ctx sdk.Context, planId string, trader sdk.AccAddress) (*types.Plan, error) {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		return nil, types.ErrPlanNotFound
	}

	if !plan.PreGraduation() {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "planId: %d, status: %s", plan.Id, plan.GetGraduationStatus())
	}

	// Validate trading enabled and start time started (unless the trader is the owner)
	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)
	if owner.Equals(trader) {
		return &plan, nil
	}

	if !plan.TradingEnabled {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading disabled")
	}

	if ctx.BlockTime().Before(plan.StartTime) {
		return nil, errorsmod.Wrapf(types.ErrPlanNotStarted, "planId: %d", plan.Id)
	}
	return &plan, nil
}

// chargeTakerFee charges taker fee from the sender.
// The fee is sent to the txfees module and the beneficiary if presented.
func (k Keeper) chargeTakerFee(ctx sdk.Context, takerFeeCoin sdk.Coin, sender sdk.AccAddress, beneficiary *sdk.AccAddress) error {
	err := k.tk.ChargeFeesFromPayer(ctx, sender, takerFeeCoin, beneficiary)
	if err != nil {
		return fmt.Errorf("charge fees: sender: %s: fee: %s: %w", sender, takerFeeCoin, err)
	}
	return nil
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

	if !newAmt.IsPositive() {
		return math.Int{}, math.Int{}, errorsmod.Wrapf(types.ErrInvalidCost, "taking fee resulted in negative amount: %s, fee: %s", newAmt.String(), feeAmt.String())
	}

	return newAmt, feeAmt, nil
}
