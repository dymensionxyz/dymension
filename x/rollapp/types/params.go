package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
	"gopkg.in/yaml.v2"
)

var (
	DefaultAppRegistrationFee         = commontypes.Dym(math.NewInt(1))
	DefaultMinSequencerBondGlobalCoin = commontypes.Dym(math.NewInt(100))
)

const (
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1

	DefaultLivenessSlashBlocks   = uint64(7200) // 12 hours worth of blocks at 1 block per 6 seconds
	DefaultLivenessSlashInterval = uint64(600)  // 1 hour worth of blocks at 1 block per 6 seconds
)

// NewParams creates a new Params instance
func NewParams(
	disputePeriodInBlocks uint64,
	livenessSlashBlocks uint64,
	livenessSlashInterval uint64,
	appRegistrationFee sdk.Coin,
	minSequencerBondGlobal sdk.Coin,
) Params {
	return Params{
		DisputePeriodInBlocks:  disputePeriodInBlocks,
		LivenessSlashBlocks:    livenessSlashBlocks,
		LivenessSlashInterval:  livenessSlashInterval,
		AppRegistrationFee:     appRegistrationFee,
		MinSequencerBondGlobal: minSequencerBondGlobal,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultDisputePeriodInBlocks,
		DefaultLivenessSlashBlocks,
		DefaultLivenessSlashInterval,
		DefaultAppRegistrationFee,
		DefaultMinSequencerBondGlobalCoin,
	)
}

func (p Params) WithDisputePeriodInBlocks(x uint64) Params {
	p.DisputePeriodInBlocks = x
	return p
}

func (p Params) WithLivenessSlashBlocks(x uint64) Params {
	p.LivenessSlashBlocks = x
	return p
}

func (p Params) WithLivenessSlashInterval(x uint64) Params {
	p.LivenessSlashInterval = x
	return p
}

// Validate validates the set of params
func (p Params) ValidateBasic() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return errorsmod.Wrap(err, "dispute period")
	}

	if err := validateLivenessSlashBlocks(p.LivenessSlashBlocks); err != nil {
		return errorsmod.Wrap(err, "liveness slash blocks")
	}
	if err := validateLivenessSlashInterval(p.LivenessSlashInterval); err != nil {
		return errorsmod.Wrap(err, "liveness slash interval")
	}

	if err := validateAppRegistrationFee(p.AppRegistrationFee); err != nil {
		return errorsmod.Wrap(err, "app registration fee")
	}
	if err := uparam.ValidateCoin(p.MinSequencerBondGlobal); err != nil {
		return errorsmod.Wrap(err, "min sequencer bond")
	}
	return nil
}

// Validate implements the ParamSet interface
func (p Params) Validate() error {
	return p.ValidateBasic()
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateLivenessSlashBlocks(v uint64) error {
	return uparam.ValidatePositiveUint64(v)
}

func validateLivenessSlashInterval(v uint64) error {
	return uparam.ValidatePositiveUint64(v)
}

// validateDisputePeriodInBlocks validates the DisputePeriodInBlocks param
func validateDisputePeriodInBlocks(disputePeriodInBlocks uint64) error {
	if disputePeriodInBlocks < MinDisputePeriodInBlocks {
		return errors.New("dispute period cannot be lower than 1 block")
	}

	return nil
}

func validateAppRegistrationFee(v sdk.Coin) error {
	if !v.IsValid() {
		return fmt.Errorf("invalid app creation cost: %s", v)
	}

	return nil
}
