package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
