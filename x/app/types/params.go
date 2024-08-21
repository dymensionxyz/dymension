package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyAppCost defines the key to store the cost of the app
	KeyAppCost = []byte("KeyAppCost")
)

var (
	DefaultAppCost = sdk.NewCoin(params.BaseDenom, types.DYM)
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	cost sdk.Coin,
) Params {
	return Params{
		Cost: cost,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultAppCost)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyAppCost, &p.Cost, validateCost),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateCost(p.Cost); err != nil {
		return err
	}
	return nil
}

func validateCost(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !v.IsValid() {
		return fmt.Errorf("invalid cost: %s", v)
	}

	return nil
}
