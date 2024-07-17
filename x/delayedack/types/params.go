package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyEpochIdentifier is the key for the epoch identifier
	KeyEpochIdentifier = []byte("EpochIdentifier")

	// KeyBridgeFee is the key for the bridge fee
	KeyBridgeFee = []byte("BridgeFee")

	// KeyDeletePacketsEpochLimit is the key for the delete packets epoch limit
	KeyDeletePacketsEpochLimit = []byte("DeletePacketsEpochLimit")
)

const (
	defaultEpochIdentifier         = "hour"
	defaultDeletePacketsEpochLimit = 1000_000
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(epochIdentifier string, bridgingFee sdk.Dec, deletePacketsEpochLimit int) Params {
	return Params{
		EpochIdentifier:         epochIdentifier,
		BridgingFee:             bridgingFee,
		DeletePacketsEpochLimit: int32(deletePacketsEpochLimit),
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		defaultEpochIdentifier,
		sdk.NewDecWithPrec(1, 3), // 0.1%
		defaultDeletePacketsEpochLimit,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, validateEpochIdentifier),
		paramtypes.NewParamSetPair(KeyBridgeFee, &p.BridgingFee, validateBridgingFee),
		paramtypes.NewParamSetPair(KeyDeletePacketsEpochLimit, &p.DeletePacketsEpochLimit, validateDeletePacketsEpochLimit),
	}
}

func validateBridgingFee(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("bridging fee must be positive: %s", v)
	}

	if v.GTE(sdk.OneDec()) {
		return fmt.Errorf("bridging fee too large: %s", v)
	}

	return nil
}

func validateEpochIdentifier(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(v) == 0 {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	return nil
}

func validateDeletePacketsEpochLimit(i interface{}) error {
	v, ok := i.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v < 0 {
		return fmt.Errorf("delete packet epoch limit must not be negative: %d", v)
	}
	return nil
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateBridgingFee(p.BridgingFee); err != nil {
		return err
	}
	if err := validateEpochIdentifier(p.EpochIdentifier); err != nil {
		return err
	}
	if err := validateDeletePacketsEpochLimit(p.DeletePacketsEpochLimit); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
