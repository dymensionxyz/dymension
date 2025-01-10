package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultNoticePeriod is the time duration for notice period
	DefaultNoticePeriod = time.Hour * 24 * 7 // 1 week
	// DefaultLivenessSlashMultiplier gives the amount of tokens to slash if the sequencer is liable for a liveness failure
	DefaultLivenessSlashMultiplier = sdk.MustNewDecFromStr("0.01")
	// DefaultLivenessSlashMinAbsolute will be slashed if the multiplier amount is too small
	DefaultLivenessSlashMinAbsolute = commontypes.DYMCoin

	DefaultPenaltyStateUpdate   = uint64(1)
	DefaultPenaltyLiveness      = uint64(300)
	DefaultPenaltyKickThreshold = uint64(900)
)

// NewParams creates a new Params instance
func NewParams(noticePeriod time.Duration, livenessSlashMul sdk.Dec, livenessSlashAbs sdk.Coin,
	penaltyStateUpdate uint64,
	penaltyLiveness uint64,
	penaltyKickThreshold uint64,
) Params {
	return Params{
		NoticePeriod:               noticePeriod,
		LivenessSlashMinMultiplier: livenessSlashMul,
		LivenessSlashMinAbsolute:   livenessSlashAbs,
		PenaltyStateUpdate:         penaltyStateUpdate,
		PenaltyLiveness:            penaltyLiveness,
		PenaltyKickThreshold:       penaltyKickThreshold,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultNoticePeriod, DefaultLivenessSlashMultiplier, DefaultLivenessSlashMinAbsolute, DefaultPenaltyStateUpdate, DefaultPenaltyLiveness, DefaultPenaltyKickThreshold)
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

func validateLivenessSlashMultiplier(i interface{}) error {
	return uparam.ValidateZeroToOneDec(i)
}

// ValidateBasic validates the set of params
func (p Params) ValidateBasic() error {
	if err := validateTime(p.NoticePeriod); err != nil {
		return err
	}

	if err := validateLivenessSlashMultiplier(p.LivenessSlashMinMultiplier); err != nil {
		return err
	}

	if err := uparam.ValidateCoin(p.LivenessSlashMinAbsolute); err != nil {
		return err
	}

	if err := uparam.ValidateUint64(p.PenaltyKickThreshold); err != nil {
		return err
	}
	if err := uparam.ValidateUint64(p.PenaltyLiveness); err != nil {
		return err
	}
	if err := uparam.ValidateUint64(p.PenaltyKickThreshold); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
