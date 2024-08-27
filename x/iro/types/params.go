package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

// Default parameter values

var (
	DefaultTakeFee     = "0.02"                                                     // 2%
	DefaultCreationFee = sdk.NewCoin(params.BaseDenom, sdk.NewInt(10).MulRaw(1e18)) /* DYM */
)

// NewParams creates a new Params object
func NewParams(takerFee sdk.Dec, creationFee sdk.Coin) Params {
	return Params{
		TakerFee:    takerFee,
		CreationFee: creationFee,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		TakerFee:    sdk.MustNewDecFromStr(DefaultTakeFee),
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
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() || v.IsNegative() {
		return fmt.Errorf("taker fee must be a non-negative decimal: %s", v)
	}

	if v.GTE(sdk.OneDec()) {
		return fmt.Errorf("taker fee must be less than 1: %s", v)
	}

	return nil
}

func validateCreationFee(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid coin: %s", v)
	}

	if v.IsZero() {
		return fmt.Errorf("creation fee must be non-zero: %s", v)
	}

	return nil
}
