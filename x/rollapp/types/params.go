package types

import (
	"errors"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"gopkg.in/yaml.v2"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyRegistrationFee is store's key for RegistrationFee Params
	KeyRegistrationFee = []byte("RegistrationFee")
	// DYM is the integer representation of 1 DYM
	DYM = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	// DefaultRegistrationFee is the default registration fee
	DefaultRegistrationFee = sdk.NewCoin(appparams.BaseDenom, DYM.Mul(sdk.NewInt(10))) // 10DYM
	// KeyDisputePeriodInBlocks is store's key for DisputePeriodInBlocks Params
	KeyDisputePeriodInBlocks = []byte("DisputePeriodInBlocks")

	KeyLivenessSlashBlocks   = []byte("LivenessSlashBlocks")
	KeyLivenessSlashInterval = []byte("LivenessSlashInterval")
	KeyLivenessJailBlocks    = []byte("LivenessJailBlocks")
)

const (
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1

	DefaultLivenessSlashBlocks   = uint64(7200)  // 12 hours at 6 blocks per second
	DefaultLivenessSlashInterval = uint64(3600)  // 1 hour at 6 blocks per second
	DefaultLivenessJailBlocks    = uint64(28800) // 48 hours at 6 blocks per second
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	disputePeriodInBlocks uint64,
	registrationFee sdk.Coin,
	livenessSlashBlocks uint64,
	livenessSlashInterval uint64,
	livenessJailBlocks uint64,
) Params {
	return Params{
		DisputePeriodInBlocks: disputePeriodInBlocks,
		RegistrationFee:       registrationFee,
		LivenessSlashBlocks:   livenessSlashBlocks,
		LivenessSlashInterval: livenessSlashInterval,
		LivenessJailBlocks:    livenessJailBlocks,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDisputePeriodInBlocks, DefaultRegistrationFee,
		DefaultLivenessSlashBlocks,
		DefaultLivenessSlashInterval,
		DefaultLivenessJailBlocks,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyRegistrationFee, &p.RegistrationFee, validateRegistrationFee),
		paramtypes.NewParamSetPair(KeyLivenessSlashBlocks, &p.LivenessSlashBlocks, validateLivenessSlashBlocks),
		paramtypes.NewParamSetPair(KeyLivenessSlashInterval, &p.LivenessSlashInterval, validateLivenessSlashInterval),
		paramtypes.NewParamSetPair(KeyLivenessJailBlocks, &p.LivenessJailBlocks, validateLivenessJailBlocks),
	}
}

func (p Params) WithDisputePeriodInBlocks(x uint64) Params {
	p.DisputePeriodInBlocks = x
	return p
}

func (p Params) WithRegFee(x sdk.Coin) Params {
	p.RegistrationFee = x
	return p
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return errorsmod.Wrap(err, "dispute period")
	}

	if err := validateLivenessSlashBlocks(p.LivenessSlashBlocks); err != nil {
		return errorsmod.Wrap(err, "liveness slash blocks")
	}
	if err := validateLivenessSlashInterval(p.LivenessSlashInterval); err != nil {
		return errorsmod.Wrap(err, "liveness slash interval")
	}
	if err := validateLivenessJailBlocks(p.LivenessJailBlocks); err != nil {
		return errorsmod.Wrap(err, "liveness jail blocks")
	}

	return validateRegistrationFee(p.RegistrationFee)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateLivenessSlashBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("must be positive: %d", v)
	}
	return nil
}

func validateLivenessSlashInterval(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("must be positive: %d", v)
	}
	return nil
}

func validateLivenessJailBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("must be positive: %d", v)
	}
	return nil
}

// validateDisputePeriodInBlocks validates the DisputePeriodInBlocks param
func validateDisputePeriodInBlocks(v interface{}) error {
	disputePeriodInBlocks, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if disputePeriodInBlocks < MinDisputePeriodInBlocks {
		return errors.New("dispute period cannot be lower than 1 block")
	}

	return nil
}

// validateRegistrationFee validates the RegistrationFee param
func validateRegistrationFee(v interface{}) error {
	registrationFee, ok := v.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if !registrationFee.IsValid() {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "registration fee: %s", registrationFee)
	}

	if registrationFee.Denom != appparams.BaseDenom {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "registration fee denom: %s", registrationFee)
	}

	return nil
}
