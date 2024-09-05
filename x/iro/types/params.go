package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// Default parameter values

var (
	DefaultTakerFee    = "0.02"                       // 2%
	DefaultCreationFee = math.NewInt(10).MulRaw(1e18) /* 10 DYM */
)

// NewParams creates a new Params object
func NewParams(takerFee math.LegacyDec, creationFee math.Int) Params {
	return Params{
		TakerFee:    takerFee,
		CreationFee: creationFee,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		TakerFee:    math.LegacyMustNewDecFromStr(DefaultTakerFee),
		CreationFee: DefaultCreationFee,
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
