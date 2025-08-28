package types

import (
	fmt "fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default parameter values

var (
	DefaultTakerFee                                     = "0.02"                      // 2%
	DefaultCreationFee                                  = math.NewInt(1).MulRaw(1e18) /* 1 Rollapp token */
	DefaultMinPlanDuration                              = 0 * time.Hour               // no enforced minimum by default
	DefaultIncentivePlanMinimumNumEpochsPaidOver        = uint64(364)                 // default: min 364 days (based on 1 day distribution epoch)
	DefaultIncentivePlanMinimumStartTimeAfterSettlement = 60 * time.Minute            // default: min 1 hour after settlement
	DefaultMinLiquidityPart                             = "0.4"                       // default: at least 40% goes to the liquidity pool
	DefaultMinVestingDuration                           = 7 * 24 * time.Hour          // default: min 7 days
	DefaultMinVestingStartTimeAfterSettlement           = 0 * time.Minute             // default: no enforced minimum by default

	DefaultFairLaunch = FairLaunch{
		AllocationAmount: math.NewInt(1e9).MulRaw(1e18),                      // 1B RA tokens
		TargetRaise:      sdk.NewCoin("adym", math.NewInt(2e4).MulRaw(1e18)), // 20K DYM
		CurveExp:         math.LegacyMustNewDecFromStr("1.5"),
		LiquidityPart:    math.LegacyMustNewDecFromStr("1.0"),
	}
)

// NewParams creates a new Params object
func NewParams(takerFee, liquidityPart math.LegacyDec, creationFee math.Int, minPlanDuration time.Duration, minIncentivePlanParams IncentivePlanParams, minVestingDuration, minVestingStartTimeAfterSettlement time.Duration, fairLaunch FairLaunch) Params {
	return Params{
		TakerFee:                              takerFee,
		CreationFee:                           creationFee,
		MinPlanDuration:                       minPlanDuration,
		IncentivesMinStartTimeAfterSettlement: minIncentivePlanParams.StartTimeAfterSettlement,
		IncentivesMinNumEpochsPaidOver:        minIncentivePlanParams.NumEpochsPaidOver,
		MinLiquidityPart:                      liquidityPart,
		MinVestingDuration:                    minVestingDuration,
		MinVestingStartTimeAfterSettlement:    minVestingStartTimeAfterSettlement,
		FairLaunch:                            fairLaunch,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		TakerFee:                              math.LegacyMustNewDecFromStr(DefaultTakerFee),
		CreationFee:                           DefaultCreationFee,
		MinPlanDuration:                       DefaultMinPlanDuration,
		IncentivesMinStartTimeAfterSettlement: DefaultIncentivePlanMinimumStartTimeAfterSettlement,
		IncentivesMinNumEpochsPaidOver:        DefaultIncentivePlanMinimumNumEpochsPaidOver,
		MinLiquidityPart:                      math.LegacyMustNewDecFromStr(DefaultMinLiquidityPart),
		MinVestingDuration:                    DefaultMinVestingDuration,
		MinVestingStartTimeAfterSettlement:    DefaultMinVestingStartTimeAfterSettlement,
		FairLaunch:                            DefaultFairLaunch,
	}
}

// Validate checks that the parameters have valid values.
func (p Params) ValidateBasic() error {
	if err := validateTakerFee(p.TakerFee); err != nil {
		return err
	}

	if err := validateCreationFee(p.CreationFee); err != nil {
		return err
	}

	if p.MinPlanDuration < 0 {
		return fmt.Errorf("minimum plan duration must be non-negative: %v", p.MinPlanDuration)
	}

	if p.IncentivesMinNumEpochsPaidOver < 1 {
		return fmt.Errorf("incentive plan num epochs paid over must be greater than 0: %d", p.IncentivesMinNumEpochsPaidOver)
	}

	if p.IncentivesMinStartTimeAfterSettlement <= 0 {
		return fmt.Errorf("incentive plan start time after settlement must be greater than 0: %v", p.IncentivesMinStartTimeAfterSettlement)
	}

	if !p.MinLiquidityPart.IsPositive() || p.MinLiquidityPart.GT(math.LegacyOneDec()) {
		return fmt.Errorf("min liquidity part must be positive: %s", p.MinLiquidityPart)
	}

	if p.MinVestingDuration < 0 {
		return fmt.Errorf("minimum vesting duration must be non-negative: %v", p.MinVestingDuration)
	}

	if !p.FairLaunch.AllocationAmount.IsPositive() {
		return fmt.Errorf("allocation amount must be positive: %s", p.FairLaunch.AllocationAmount)
	}

	if !p.FairLaunch.TargetRaise.IsValid() {
		return fmt.Errorf("target raise is not valid: %s", p.FairLaunch.TargetRaise)
	}

	if !p.FairLaunch.LiquidityPart.IsPositive() || p.FairLaunch.LiquidityPart.GT(math.LegacyOneDec()) {
		return fmt.Errorf("liquidity part must be positive and less than 1: %s", p.FairLaunch.LiquidityPart)
	}

	return nil
}

func validateTakerFee(v math.LegacyDec) error {
	if v.IsNil() || v.IsNegative() {
		return fmt.Errorf("taker fee must be a non-negative decimal: %s", v)
	}

	if v.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("taker fee must be less than 1: %s", v)
	}

	return nil
}

func validateCreationFee(v math.Int) error {
	// creation fee must be a positive integer greater than 1^18 (1 Rollapp token)
	if v.LT(math.NewIntWithDecimal(1, 18)) {
		return fmt.Errorf("creation fee must be a positive integer: %s", v)
	}

	return nil
}
