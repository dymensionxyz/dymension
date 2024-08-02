package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// DefaultMinBond is the minimum bond required to be a validator
	DefaultMinBond uint64 = 1000000
	// DefaultUnbondingTime is the time duration for unbonding
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 2 // 2 weeks

	// KeyMinBond is store's key for MinBond Params
	KeyMinBond = []byte("MinBond")
	// KeyUnbondingTime is store's key for UnbondingTime Params
	KeyUnbondingTime = []byte("UnbondingTime")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(minBond sdk.Coin, unbondingPeriod time.Duration) Params {
	return Params{
		MinBond:       minBond,
		UnbondingTime: unbondingPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	denom, err := sdk.GetBaseDenom()
	if err != nil {
		panic(err)
	}
	minBond := sdk.NewCoin(denom, sdk.NewIntFromUint64(DefaultMinBond))
	return NewParams(
		minBond, DefaultUnbondingTime,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinBond, &p.MinBond, validateMinBond),
		paramtypes.NewParamSetPair(KeyUnbondingTime, &p.UnbondingTime, validateUnbondingTime),
	}
}

func validateUnbondingTime(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("unbonding time must be positive: %d", v)
	}

	return nil
}

func validateMinBond(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() || v.IsZero() {
		return nil
	}

	if !v.IsValid() {
		return fmt.Errorf("invalid coin: %s", v)
	}
	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMinBond(p.MinBond); err != nil {
		return err
	}

	if err := validateUnbondingTime(p.UnbondingTime); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
