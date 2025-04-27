package types

import (
	"fmt"

	"cosmossdk.io/math"
	"gopkg.in/yaml.v2"
)

const (
	defaultEpochIdentifier = "hour"
	defaultTimeoutFee      = "0.0015"
	defaultErrAckFee       = "0.0015"
)

// NewParams creates a new Params instance
func NewParams(epochIdentifier string, timeoutFee math.LegacyDec, errAckFee math.LegacyDec) Params {
	return Params{
		EpochIdentifier: epochIdentifier,
		TimeoutFee:      timeoutFee,
		ErrackFee:       errAckFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(defaultEpochIdentifier, math.LegacyMustNewDecFromStr(defaultTimeoutFee), math.LegacyMustNewDecFromStr(defaultErrAckFee))
}

// Validate validates the set of params
func (p Params) ValidateBasic() error {
	if err := validateEpochIdentifier(p.EpochIdentifier); err != nil {
		return fmt.Errorf("epoch identifier: %w", err)
	}
	if err := validateTimeoutFee(p.TimeoutFee); err != nil {
		return fmt.Errorf("timeout fee: %w", err)
	}
	if err := validateErrAckFee(p.ErrackFee); err != nil {
		return fmt.Errorf("error acknowledgement fee: %w", err)
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateEpochIdentifier(v string) error {
	if len(v) == 0 {
		return fmt.Errorf("epoch identifier cannot be empty")
	}
	return nil
}

func validateTimeoutFee(v math.LegacyDec) error {
	if v.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", v)
	}
	if v.IsNegative() {
		return ErrNegativeFee
	}

	if v.GTE(math.LegacyOneDec()) {
		return ErrFeeTooHigh
	}

	return nil
}

func validateErrAckFee(v math.LegacyDec) error {
	if v.IsNil() {
		return fmt.Errorf("invalid global pool params: %+v", v)
	}
	if v.IsNegative() {
		return ErrNegativeFee
	}

	if v.GTE(math.LegacyOneDec()) {
		return ErrFeeTooHigh
	}

	return nil
}
