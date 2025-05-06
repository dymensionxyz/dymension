package eibc

import (
	"fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "eibc"
)

// Params defines the parameters for the eibc module
type Params struct {
	EpochIdentifier string         `json:"epoch_identifier" yaml:"epoch_identifier"`
	BridgeFee       math.LegacyDec `json:"bridge_fee" yaml:"bridge_fee"`
}

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyEpochIdentifier is the key for the epoch identifier
	KeyEpochIdentifier = []byte("EpochIdentifier")

	// KeyBridgeFee is the key for the bridge fee
	KeyBridgeFee = []byte("BridgeFee")
)

const (
	defaultEpochIdentifier = "hour"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(epochIdentifier string, bridgeFee math.LegacyDec) Params {
	return Params{
		EpochIdentifier: epochIdentifier,
		BridgeFee:       bridgeFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		defaultEpochIdentifier,
		math.LegacyNewDecWithPrec(1, 3), // 0.1%
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, validateEpochIdentifier),
		paramtypes.NewParamSetPair(KeyBridgeFee, &p.BridgeFee, validateBridgeFee),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEpochIdentifier(p.EpochIdentifier); err != nil {
		return err
	}
	if err := validateBridgeFee(p.BridgeFee); err != nil {
		return err
	}
	return nil
}

func validateEpochIdentifier(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	return nil
}

func validateBridgeFee(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("bridge fee cannot be negative")
	}
	return nil
}
