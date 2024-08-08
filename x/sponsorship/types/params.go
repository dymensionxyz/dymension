package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
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
		return errorsmod.Wrapf(ErrInvalidParams, "MinAllocationWeight must be >= 0, got %d", p.MinAllocationWeight.Int64())
	}
	if p.MinAllocationWeight.GT(hundred) {
		return errorsmod.Wrapf(ErrInvalidParams, "MinAllocationWeight must be <= 100, got %d", p.MinAllocationWeight.Int64())
	}
	if p.MinVotingPower.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidParams, "MinVotingPower must be >= 0, got %d", p.MinVotingPower.Int64())
	}
	return nil
}

// ParamKeyTable for the x/sponsorship module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements params.ParamSet.
func (p Params) ParamSetPairs() paramtypes.ParamSetPairs {
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
		return errorsmod.Wrapf(ErrInvalidParams, "MinAllocationWeight must be >= 0, got %d", value.Int64())
	}
	if value.GT(hundred) {
		return errorsmod.Wrapf(ErrInvalidParams, "MinAllocationWeight must be <= 100, got %d", value.Int64())
	}

	return nil
}

func validateMinVotingPower(i interface{}) error {
	value, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if value.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidParams, "MinVotingPower must be >= 0, got %d", value.Int64())
	}

	return nil
}
