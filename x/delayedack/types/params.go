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

// Validate validates the set of params
func (p Params) ValidateBasic() error {
	// validate bridging fee
	if p.BridgingFee.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", p.BridgingFee)
	}
	if p.BridgingFee.IsNegative() {
		return fmt.Errorf("bridging fee must be positive: %s", p.BridgingFee)
	}

	if p.BridgingFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("bridging fee too large: %s", p.BridgingFee)
	}

	// validate epoch identifier
	if p.EpochIdentifier == "" {
		return fmt.Errorf("epoch identifier cannot be empty")
	}

	// validate delete packets epoch limit
	if p.DeletePacketsEpochLimit < 0 {
		return fmt.Errorf("delete packet epoch limit must not be negative: %d", p.DeletePacketsEpochLimit)
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
