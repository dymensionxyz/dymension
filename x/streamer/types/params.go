package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	KeyMaxIterationsPerBlock = "MaxIterationsPerBlock"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maxIterationsPerBlock uint64) Params {
	return Params{
		MaxIterationsPerBlock: maxIterationsPerBlock,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxIterationsPerBlock)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte(KeyMaxIterationsPerBlock), &p.MaxIterationsPerBlock, validateMaxIterationsPerBlock),
	}
}

// validateDisputePeriodInBlocks validates the DisputePeriodInBlocks param
func validateMaxIterationsPerBlock(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}
	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}
