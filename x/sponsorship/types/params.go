package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyMinAllocationWeight = []byte("MinAllocationWeight")
	KeyMinVotingPower      = []byte("MinVotingPower")
)

func DefaultParams() Params {
	return Params{
		MinAllocationWeight: DefaultMinAllocationWeight,
		MinVotingPower:      DefaultMinVotingPower,
	}
}

func (p Params) Validate() error {
	if p.MinAllocationWeight.IsNegative() {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be >= 0, got %s", p.MinAllocationWeight)
	}
	if p.MinAllocationWeight.GT(MaxAllocationWeight) {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be <= 100 * 10^18, got %s", p.MinAllocationWeight)
	}
	if p.MinVotingPower.IsNegative() {
		return ErrInvalidParams.Wrapf("MinVotingPower must be >= 0, got %s", p.MinVotingPower)
	}
	return nil
}

// ParamKeyTable for the x/sponsorship module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet. Params must have a pointer receiver since it is registered as
// a pointer in the ParamKeyTable method.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinAllocationWeight, &p.MinAllocationWeight, validateMinAllocationWeight),
		paramtypes.NewParamSetPair(KeyMinVotingPower, &p.MinVotingPower, validateMinVotingPower),
	}
}

func validateMinAllocationWeight(i interface{}) error {
	value, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value.IsNegative() {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be >= 0, got %s", value)
	}
	if value.GT(MaxAllocationWeight) {
		return ErrInvalidParams.Wrapf("MinAllocationWeight must be <= 100 * 10^18, got %s", value)
	}

	return nil
}

func validateMinVotingPower(i interface{}) error {
	value, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value.IsNegative() {
		return ErrInvalidParams.Wrapf("MinVotingPower must be >= 0, got %s", value)
	}

	return nil
}
