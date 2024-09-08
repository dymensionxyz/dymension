package types

import (
	fmt "fmt"
	"time"

	"cosmossdk.io/math"
)

// Default parameter values

var (
	DefaultTakerFee                   = "0.02"                       // 2%
	DefaultCreationFee                = math.NewInt(10).MulRaw(1e18) /* 10 DYM */
	DefaultMinPlanDuration            = 7 * 24 * time.Hour           // 7 days
	DefaultIncentivePlanMinimumParams = IncentivePlanParams{
		NumEpochsPaidOver:        10_080,           // default: min 7 days (based on 1 minute distribution epoch)
		StartTimeAfterSettlement: 60 * time.Minute, // default: min 1 hour after settlement
	}
)

// NewParams creates a new Params object
func NewParams(takerFee math.LegacyDec, creationFee math.Int, minPlanDuration time.Duration, minIncentivePlanParams IncentivePlanParams) Params {
	return Params{
		TakerFee:                   takerFee,
		CreationFee:                creationFee,
		MinPlanDuration:            minPlanDuration,
		IncentivePlanMinimumParams: minIncentivePlanParams,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		TakerFee:                   math.LegacyMustNewDecFromStr(DefaultTakerFee),
		CreationFee:                DefaultCreationFee,
		MinPlanDuration:            DefaultMinPlanDuration,
		IncentivePlanMinimumParams: DefaultIncentivePlanMinimumParams,
	}
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateTakerFee(p.TakerFee); err != nil {
		return err
	}

	if err := validateCreationFee(p.CreationFee); err != nil {
		return err
	}

	if p.MinPlanDuration <= 0 {
		return fmt.Errorf("minimum plan duration must be greater than 0: %s", p.MinPlanDuration)
	}

	if err := validateIncentivePlanParams(p.IncentivePlanMinimumParams); err != nil {
		return err
	}

	return nil
}

func validateTakerFee(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() || v.IsNegative() {
		return fmt.Errorf("taker fee must be a non-negative decimal: %s", v)
	}

	if v.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("taker fee must be less than 1: %s", v)
	}

	return nil
}

func validateCreationFee(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !v.IsPositive() {
		return fmt.Errorf("creation fee must be a positive integer: %s", v)
	}

	return nil
}

// validateIncentivePlanParams
func validateIncentivePlanParams(i interface{}) error {
	v, ok := i.(IncentivePlanParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.NumEpochsPaidOver < 1 {
		return fmt.Errorf("incentive plan num epochs paid over must be greater than 0: %d", v.NumEpochsPaidOver)
	}

	if v.StartTimeAfterSettlement <= 0 {
		return fmt.Errorf("incentive plan start time after settlement must be greater than 0: %s", v.StartTimeAfterSettlement)
	}

	return nil
}
