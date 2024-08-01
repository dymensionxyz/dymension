package types

import (
	fmt "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// MinBond types.Coin `protobuf:"bytes,1,opt,name=min_bond,json=minBond,proto3" json:"min_bond,omitempty"`
	// UnbondingTime time.Duration `protobuf:"bytes,2,opt,name=unbonding_time,json=unbondingTime,proto3,stdduration" json:"unbonding_time"`

	// MinBond is the minimum bond required to be a validator
	DefaultMinBond uint64 = 1000000
	// UnbondingTime is the time duration for unbonding
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 2 // 2 weeks
	// DefaultNoticePeriod is the time duration for notice period
	DefaultNoticePeriod time.Duration = time.Hour * 24 * 7 // 1 week

	// KeyMinBond is store's key for MinBond Params
	KeyMinBond = []byte("MinBond")
	// KeyUnbondingTime is store's key for UnbondingTime Params
	KeyUnbondingTime = []byte("UnbondingTime")
	// KeyNoticePeriod is store's key for NoticePeriod Params
	KeyNoticePeriod = []byte("NoticePeriod")
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(minBond sdk.Coin, unbondingPeriod, noticePeriod time.Duration) Params {
	return Params{
		MinBond:       minBond,
		UnbondingTime: unbondingPeriod,
		NoticePeriod:  noticePeriod,
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
		minBond, DefaultUnbondingTime, DefaultNoticePeriod,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinBond, &p.MinBond, validateMinBond),
		paramtypes.NewParamSetPair(KeyUnbondingTime, &p.UnbondingTime, validateTime),
		paramtypes.NewParamSetPair(KeyNoticePeriod, &p.NoticePeriod, validateTime),
	}
}

func validateTime(i interface{}) error {
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

	if err := validateTime(p.UnbondingTime); err != nil {
		return err
	}

	if err := validateTime(p.NoticePeriod); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
