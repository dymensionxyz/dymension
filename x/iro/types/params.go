package types

import (
	fmt "fmt"
	"time"

	"cosmossdk.io/math"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// Default parameter values

var (
	DefaultTakerFee                                     = "0.02"                        // 2%
	DefaultCreationFee                                  = math.NewInt(1).MulRaw(1e18)   /* 1 Rollapp token */
	DefaultMinPlanDuration                              = 0 * time.Hour                 // no enforced minimum by default
	DefaultIncentivePlanMinimumNumEpochsPaidOver        = uint64(364)                   // default: min 364 days (based on 1 day distribution epoch)
	DefaultIncentivePlanMinimumStartTimeAfterSettlement = 60 * time.Minute              // default: min 1 hour after settlement
	DefaultMinLiquidityPart                             = "0.4"                         // default: at least 40% goes to the liquidity pool
	DefaultMinVestingDuration                           = 7 * 24 * time.Hour            // default: min 7 days
	DefaultMinVestingStartTimeAfterSettlement           = 0 * time.Minute               // default: no enforced minimum by default
	DefaultMinTradeAmount                               = math.NewIntWithDecimal(1, 16) // 0.01 DYM

	DefaultStandardLaunch = StandardLaunch{
		AllocationAmount: math.NewInt(1e9).MulRaw(1e18), // 1B RA tokens
		TargetRaise:      common.DymUint64(20_000),      // 20K DYM
		CurveExp:         math.LegacyMustNewDecFromStr("1.25"),
		InitialFdv:       common.DYM.MulRaw(10_000), // 10K DYM
	}
)

// NewParams creates a new Params object
func NewParams(takerFee, liquidityPart math.LegacyDec, creationFee, minTradeAmount math.Int, minPlanDuration time.Duration, minIncentivePlanParams IncentivePlanParams, minVestingDuration, minVestingStartTimeAfterSettlement time.Duration, standardLaunch StandardLaunch) Params {
	return Params{
		TakerFee:                              takerFee,
		CreationFee:                           creationFee,
		MinPlanDuration:                       minPlanDuration,
		IncentivesMinStartTimeAfterSettlement: minIncentivePlanParams.StartTimeAfterSettlement,
		IncentivesMinNumEpochsPaidOver:        minIncentivePlanParams.NumEpochsPaidOver,
		MinLiquidityPart:                      liquidityPart,
		MinVestingDuration:                    minVestingDuration,
		MinVestingStartTimeAfterSettlement:    minVestingStartTimeAfterSettlement,
		MinTradeAmount:                        minTradeAmount,
		StandardLaunch:                        standardLaunch,
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
		MinTradeAmount:                        DefaultMinTradeAmount,
		StandardLaunch:                        DefaultStandardLaunch,
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

	if err := validateMinTradeAmount(p.MinTradeAmount); err != nil {
		return err
	}

	if err := p.StandardLaunch.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

// validate standard launch params
func (p StandardLaunch) ValidateBasic() error {
	if !p.AllocationAmount.IsPositive() {
		return fmt.Errorf("allocation amount must be positive: %s", p.AllocationAmount)
	}
	if !p.TargetRaise.IsValid() || !p.TargetRaise.Amount.IsPositive() {
		return fmt.Errorf("target raise is not valid: %s", p.TargetRaise)
	}
	if !p.InitialFdv.IsPositive() {
		return fmt.Errorf("initial TVL must be positive: %s", p.InitialFdv)
	}
	if !p.CurveExp.IsPositive() {
		return fmt.Errorf("curve exponent must be positive: %s", p.CurveExp)
	}
	if p.CurveExp.GT(math.LegacyNewDec(MaxNValue)) {
		return fmt.Errorf("curve exponent exceeds maximum value of %d: %s", MaxNValue, p.CurveExp)
	}
	if !checkPrecision(p.CurveExp) {
		return fmt.Errorf("curve exponent must have at most %d decimal places", MaxNPrecision)
	}
	return nil
}

func validateMinTradeAmount(v math.Int) error {
	if v.IsNil() || v.IsNegative() {
		return fmt.Errorf("min trade amount must be non-negative: %s", v)
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
