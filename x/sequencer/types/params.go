package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
	"gopkg.in/yaml.v2"
)

var (
	// DefaultNoticePeriod is the time duration for notice period
	DefaultNoticePeriod = time.Hour * 24 * 7 // 1 week
	// DefaultLivenessSlashMultiplier gives the amount of tokens to slash if the sequencer is liable for a liveness failure
	DefaultLivenessSlashMultiplier = math.LegacyMustNewDecFromStr("0.01")
	// DefaultLivenessSlashMinAbsolute will be slashed if the multiplier amount is too small
	DefaultLivenessSlashMinAbsolute = commontypes.DYMCoin

	DefaultDishonorStateUpdate   = uint64(1)
	DefaultDishonorLiveness      = uint64(300)
	DefaultDishonorKickThreshold = uint64(900)
)

// NewParams creates a new Params instance
func NewParams(noticePeriod time.Duration, livenessSlashMul math.LegacyDec, livenessSlashAbs sdk.Coin,
	dishonorStateUpdate uint64,
	dishonorLiveness uint64,
	dishonorKickThreshold uint64,
) Params {
	return Params{
		NoticePeriod:               noticePeriod,
		LivenessSlashMinMultiplier: livenessSlashMul,
		LivenessSlashMinAbsolute:   livenessSlashAbs,
		DishonorStateUpdate:        dishonorStateUpdate,
		DishonorLiveness:           dishonorLiveness,
		DishonorKickThreshold:      dishonorKickThreshold,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultNoticePeriod, DefaultLivenessSlashMultiplier, DefaultLivenessSlashMinAbsolute, DefaultDishonorStateUpdate, DefaultDishonorLiveness, DefaultDishonorKickThreshold)
}

func validateTime(v time.Duration) error {
	if v <= 0 {
		return fmt.Errorf("time must be positive: %d", v)
	}

	return nil
}

func validateLivenessSlashMultiplier(v math.LegacyDec) error {
	return uparam.ValidateZeroToOneDec(v)
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

	if err := uparam.ValidateUint64(p.PenaltyLiveness()); err != nil {
		return err
	}
	if err := uparam.ValidateUint64(p.PenaltyReductionStateUpdate()); err != nil {
		return err
	}
	if err := uparam.ValidateUint64(p.PenaltyKickThreshold()); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p Params) PenaltyLiveness() uint64 {
	return p.DishonorLiveness
}

func (p Params) PenaltyReductionStateUpdate() uint64 {
	return p.DishonorStateUpdate
}

func (p Params) PenaltyKickThreshold() uint64 {
	return p.DishonorKickThreshold
}

func (p *Params) SetPenaltyLiveness(x uint64) {
	p.DishonorLiveness = x
}

func (p *Params) SetPenaltyReductionStateUpdate(x uint64) {
	p.DishonorStateUpdate = x
}

func (p *Params) SetPenaltyKickThreshold(x uint64) {
	p.DishonorKickThreshold = x
}
