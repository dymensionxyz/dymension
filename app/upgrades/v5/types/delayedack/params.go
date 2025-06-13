package delayedack

import (
	"fmt"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "delayedack"
)

// Params defines the parameters for the delayedack module
type Params struct {
	EpochIdentifier         string         `json:"epoch_identifier" yaml:"epoch_identifier"`
	BridgingFee             math.LegacyDec `json:"bridging_fee" yaml:"bridging_fee"`
	DeletePacketsEpochLimit int32          `json:"delete_packets_epoch_limit" yaml:"delete_packets_epoch_limit"`
}

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

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		defaultEpochIdentifier,
		math.LegacyNewDecWithPrec(1, 3), // 0.1%
		defaultDeletePacketsEpochLimit,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEpochIdentifier, &p.EpochIdentifier, validateEpochIdentifier),
		paramtypes.NewParamSetPair(KeyBridgeFee, &p.BridgingFee, validateBridgingFee),
		paramtypes.NewParamSetPair(KeyDeletePacketsEpochLimit, &p.DeletePacketsEpochLimit, validateDeletePacketsEpochLimit),
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
	if err := validateBridgingFee(p.BridgingFee); err != nil {
		return err
	}
	if err := validateDeletePacketsEpochLimit(p.DeletePacketsEpochLimit); err != nil {
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

func validateBridgingFee(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("bridging fee cannot be negative")
	}
	return nil
}

func validateDeletePacketsEpochLimit(i interface{}) error {
	v, ok := i.(int32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("delete packets epoch limit must be positive")
	}
	return nil
}
