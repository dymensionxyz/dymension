package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
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

	// non fair launched plans need to set the pre launch time
	// fair launched plans will allow launch once graduated
	if !plan.FairLaunched {
		k.rk.SetPreLaunchTime(ctx, &rollapp, plan.PreLaunchTime)
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
func (k Keeper) Buy(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountTokensToBuy, maxCostAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, buyer)
	if err != nil {
		return err
	}

	// validate the IRO have enough tokens to sell
	if plan.SoldAmt.Add(amountTokensToBuy).GT(plan.MaxAmountToSell) {
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
	takerFee := sdk.NewCoin(plan.LiquidityDenom, takerFeeAmt)
	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)
	err = k.chargeTakerFee(ctx, takerFee, buyer, &owner)
	if err != nil {
		return err
	}

	// Send liquidity token from buyer to the plan. The liquidity token sent directly to the plan's module account
	cost := sdk.NewCoin(plan.LiquidityDenom, costAmt)
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
		return err
	}

	// if all tokens are sold, we need to graduate the plan
	if plan.SoldAmt.Equal(plan.MaxAmountToSell) {
		poolID, _, err := k.GraduatePlan(ctx, planId)
		if err != nil {
			return err
		}

		err = uevent.EmitTypedEvent(ctx, &types.EventGraduation{
			PlanId:    planId,
			RollappId: plan.RollappId,
			PoolId:    poolID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// BuyExactSpend uses exact amount of liquidity to buy tokens on the curve
func (k Keeper) BuyExactSpend(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountToSpend, minTokensAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, buyer)
	if err != nil {
		return err
	}

	// deduct taker fee from the amount to spend
	toSpendMinusTakerFeeAmt, takerFeeAmt, err := k.ApplyTakerFee(amountToSpend, k.GetParams(ctx).TakerFee, false)
	if err != nil {
		return err
	}

	// calculate the amount of tokens possible to buy with the amount to spend
	tokensOutAmt, err := plan.BondingCurve.TokensForExactInAmount(plan.SoldAmt, toSpendMinusTakerFeeAmt)
	if err != nil {
		return err
	}

	// Validate expected out amount
	if tokensOutAmt.LT(minTokensAmt) {
		return errorsmod.Wrapf(types.ErrInvalidMinCost, "minTokens: %s, tokens: %s, fee: %s", minTokensAmt.String(), tokensOutAmt.String(), takerFeeAmt.String())
	}

	// validate the IRO have enough tokens to sell
	if plan.SoldAmt.Add(tokensOutAmt).GT(plan.MaxAmountToSell) {
		return types.ErrInsufficientTokens
	}

	// Charge taker fee
	takerFee := sdk.NewCoin(plan.LiquidityDenom, takerFeeAmt)
	owner := k.rk.MustGetRollappOwner(ctx, plan.RollappId)
	err = k.chargeTakerFee(ctx, takerFee, buyer, &owner)
	if err != nil {
		return err
	}

	// Send liquidity token from buyer to the plan. The liquidity token sent directly to the plan's module account
	cost := sdk.NewCoin(plan.LiquidityDenom, toSpendMinusTakerFeeAmt)
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
		return err
	}

	return nil
}

// Sell sells allocation with price according to the price curve
func (k Keeper) Sell(ctx sdk.Context, planId string, seller sdk.AccAddress, amountTokensToSell, minIncomeAmt math.Int) error {
	plan, err := k.GetTradeableIRO(ctx, planId, seller)
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

	// send allocated tokens from seller to the plan
	err = k.BK.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sdk.NewCoins(sdk.NewCoin(plan.TotalAllocation.Denom, amountTokensToSell)))
	if err != nil {
		return err
	}

	// Send liquidity token from the plan to the seller. The liquidity token managed by the plan's module account
	cost := sdk.NewCoin(plan.LiquidityDenom, costAmt)
	err = k.BK.SendCoins(ctx, plan.GetAddress(), seller, sdk.NewCoins(cost))
	if err != nil {
		return err
	}

	// Update plan
	plan.SoldAmt = plan.SoldAmt.Sub(amountTokensToSell)
	k.SetPlan(ctx, *plan)

	// Charge taker fee
	takerFee := sdk.NewCoin(plan.LiquidityDenom, takerFeeAmt)
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
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "planId: %d, status: %s", plan.Id, plan.GetGraduationStatus().String())
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

	if !newAmt.IsPositive() || !feeAmt.IsPositive() {
		return math.Int{}, math.Int{}, errorsmod.Wrapf(types.ErrInvalidCost, "taking fee resulted in negative amount: %s, fee: %s", newAmt.String(), feeAmt.String())
	}

	return newAmt, feeAmt, nil
}
