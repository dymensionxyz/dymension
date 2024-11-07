package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/ucoin"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultMinBond is the minimum bond required to be a validator
	DefaultMinBond = ucoin.SimpleMul(rollapptypes.OneDymCoin, 100)
	// DefaultKickThreshold is the minimum bond required to be a validator
	DefaultKickThreshold = rollapptypes.OneDymCoin
	// DefaultNoticePeriod is the time duration for notice period
	DefaultNoticePeriod = time.Hour * 24 * 7 // 1 week
	// DefaultLivenessSlashMultiplier gives the amount of tokens to slash if the sequencer is liable for a liveness failure
	DefaultLivenessSlashMultiplier = sdk.MustNewDecFromStr("0.01")
	// DefaultLivenessSlashMinAbsolute will be slashed if the multiplier amount is too small
	DefaultLivenessSlashMinAbsolute = rollapptypes.OneDymCoin
)

// NewParams creates a new Params instance
func NewParams(minBond sdk.Coin, noticePeriod time.Duration, livenessSlashMul sdk.Dec, livenessSlashAbs sdk.Coin, kickThreshold sdk.Coin) Params {
	return Params{
		MinBond:                    minBond,
		NoticePeriod:               noticePeriod,
		LivenessSlashMinMultiplier: livenessSlashMul,
		LivenessSlashMinAbsolute:   livenessSlashAbs,
		KickThreshold:              kickThreshold,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMinBond, DefaultNoticePeriod, DefaultLivenessSlashMultiplier, DefaultLivenessSlashMinAbsolute, DefaultKickThreshold,
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
		return fmt.Errorf("min bond must be positive: %s", v)
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

	if err := validateTime(p.NoticePeriod); err != nil {
		return err
	}

	if err := validateLivenessSlashMultiplier(p.LivenessSlashMinMultiplier); err != nil {
		return err
	}

	if err := uparam.ValidateCoin(p.LivenessSlashMinAbsolute); err != nil {
		return err
	}

	if err := uparam.ValidateCoin(p.KickThreshold); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
