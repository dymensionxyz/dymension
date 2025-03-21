package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const IROTokenPrefix = "IRO/"

func IRODenom(rollappID string) string {
	return fmt.Sprintf("%s%s", IROTokenPrefix, rollappID)
}

func RollappIDFromIRODenom(denom string) (string, bool) {
	return strings.CutPrefix(denom, IROTokenPrefix)
}

var MinTokenAllocation = math.LegacyNewDec(10) // min allocation in decimal representation

func NewPlan(id uint64, rollappId string, liquidityDenom string, allocation sdk.Coin, curve BondingCurve, planDuration time.Duration, incentivesParams IncentivePlanParams, liquidityPart math.LegacyDec, vestingDuration, vestingStartTimeAfterSettlement time.Duration) Plan {
	eq := FindEquilibrium(curve, allocation.Amount, liquidityPart)
	// start time and pre-launch time are set later on
	plan := Plan{
		Id:                  id,
		RollappId:           rollappId,
		TotalAllocation:     allocation,
		BondingCurve:        curve,
		IroPlanDuration:     planDuration,
		SoldAmt:             math.ZeroInt(),
		ClaimedAmt:          math.ZeroInt(),
		IncentivePlanParams: incentivesParams,
		MaxAmountToSell:     eq,
		LiquidityPart:       liquidityPart,
		LiquidityDenom:      liquidityDenom,
		VestingPlan: IROVestingPlan{
			Amount:                   math.ZeroInt(),
			Claimed:                  math.ZeroInt(),
			VestingDuration:          vestingDuration,
			StartTimeAfterSettlement: vestingStartTimeAfterSettlement,
		},
	}
	plan.ModuleAccAddress = authtypes.NewModuleAddress(plan.ModuleAccName()).String()
	return plan
}

// ValidateBasic checks if the plan is valid
func (p Plan) ValidateBasic() error {
	if err := p.BondingCurve.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidBondingCurve, err)
	}
	// check that the allocation is greater than the minimal allowed token allocation
	allocationDec := ScaleFromBase(p.TotalAllocation.Amount, p.BondingCurve.SupplyDecimals())
	if !allocationDec.GT(MinTokenAllocation) {
		return ErrInvalidAllocation
	}
	if p.PreLaunchTime.Before(p.StartTime) {
		return ErrInvalidEndTime
	}
	if p.ModuleAccAddress == "" {
		return errors.New("module account address cannot be empty")
	}
	if p.SoldAmt.IsNegative() {
		return fmt.Errorf("sold amount cannot be negative: %s", p.SoldAmt.String())
	}
	if p.ClaimedAmt.IsNegative() {
		return fmt.Errorf("claimed amount cannot be negative: %s", p.ClaimedAmt.String())
	}
	if !p.MaxAmountToSell.IsPositive() {
		return fmt.Errorf("max amount to sell must be positive: %s", p.MaxAmountToSell.String())
	}
	if p.MaxAmountToSell.GT(p.TotalAllocation.Amount) {
		return fmt.Errorf("max amount to sell must be less than or equal to the total allocation: %s > %s", p.MaxAmountToSell.String(), p.TotalAllocation.Amount.String())
	}

	if p.LiquidityPart.IsNegative() || p.LiquidityPart.GT(math.LegacyOneDec()) {
		return errors.New("liquidity part must be between 0 and 1")
	}

	if err := p.IncentivePlanParams.ValidateBasic(); err != nil {
		return errors.Join(ErrInvalidIncentivePlanParams, err)
	}

	if err := p.VestingPlan.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "vesting plan")
	}

	if err := sdk.ValidateDenom(p.LiquidityDenom); err != nil {
		return errorsmod.Wrap(err, "invalid liquidity denom")
	}

	return nil
}

// SpotPrice returns the spot price of the plan
func (p Plan) SpotPrice() math.LegacyDec {
	return p.BondingCurve.SpotPrice(p.SoldAmt)
}

func (p Plan) IsSettled() bool {
	return p.SettledDenom != ""
}

func (p Plan) ModuleAccName() string {
	return ModuleName + "-" + p.RollappId
}

func (p Plan) GetAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(p.ModuleAccAddress)
	return addr
}

// GetIRODenom returns IRO token's denom
func (p Plan) GetIRODenom() string {
	return IRODenom(p.RollappId)
}

// EnableTradingWithStartTime enables trading for the plan and sets the start and pre-launch times
// based on the provided start time and plan duration
func (p *Plan) EnableTradingWithStartTime(startTime time.Time) {
	p.TradingEnabled = true
	p.StartTime = startTime
	p.PreLaunchTime = startTime.Add(p.IroPlanDuration)
}

func DefaultIncentivePlanParams() IncentivePlanParams {
	return IncentivePlanParams{
		NumEpochsPaidOver:        43200, // 1 month in minute epoch
		StartTimeAfterSettlement: DefaultIncentivePlanMinimumStartTimeAfterSettlement,
	}
}

func (i IncentivePlanParams) ValidateBasic() error {
	if i.NumEpochsPaidOver == 0 {
		return errors.New("number of epochs paid over cannot be zero")
	}
	return nil
}
