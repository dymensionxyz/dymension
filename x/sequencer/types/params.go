package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultMinBond is the minimum bond required to be a validator
	DefaultMinBond uint64 = 1000000
	// DefaultUnbondingTime is the time duration for unbonding
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 2 // 2 weeks
	// DefaultNoticePeriod is the time duration for notice period
	DefaultNoticePeriod time.Duration = time.Hour * 24 * 7 // 1 week
	// DefaultLivenessSlashMultiplier gives the amount of tokens to slash if the sequencer is liable for a liveness failure
	DefaultLivenessSlashMultiplier sdk.Dec = sdk.MustNewDecFromStr("0.01907") // leaves 50% of original funds remaining after 48 slashes
)

// NewParams creates a new Params instance
func NewParams(minBond sdk.Coin, unbondingPeriod, noticePeriod time.Duration, livenessSlashMul sdk.Dec) Params {
	return Params{
		MinBond:                 minBond,
		UnbondingTime:           unbondingPeriod,
		NoticePeriod:            noticePeriod,
		LivenessSlashMultiplier: livenessSlashMul,
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
		minBond, DefaultUnbondingTime, DefaultNoticePeriod, DefaultLivenessSlashMultiplier,
	)
}

func validateTime(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("time must be positive: %d", v)
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

func validateLivenessSlashMultiplier(i interface{}) error {
	return uparam.ValidateZeroToOneDec(i)
}

// ValidateBasic validates the set of params
func (p Params) ValidateBasic() error {
	if err := validateMinBond(p.MinBond); err != nil {
		return err
	}

	if err := validateTime(p.UnbondingTime); err != nil {
		return err
	}

	if err := validateTime(p.NoticePeriod); err != nil {
		return err
	}

	if err := validateLivenessSlashMultiplier(p.LivenessSlashMultiplier); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
