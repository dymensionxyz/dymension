package types

import (
	"fmt"

	"cosmossdk.io/math"
	"gopkg.in/yaml.v2"
)

const (
	defaultEpochIdentifier         = "hour"
	defaultDeletePacketsEpochLimit = 1000_000
)

// NewParams creates a new Params instance
func NewParams(epochIdentifier string, bridgingFee math.LegacyDec, deletePacketsEpochLimit int) Params {
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
		math.LegacyNewDecWithPrec(1, 3), // 0.1%
		defaultDeletePacketsEpochLimit,
	)
}

func validateBridgingFee(fee math.LegacyDec) error {
	if fee.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", fee)
	}
	if fee.IsNegative() {
		return fmt.Errorf("bridging fee must be positive: %s", fee)
	}

	if fee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("bridging fee too large: %s", fee)
	}

	return nil
}

func validateEpochIdentifier(i string) error {
	if i == "" {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	return nil
}

func validateDeletePacketsEpochLimit(v int32) error {
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
