package types

import (
	"errors"
	"fmt"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
	"gopkg.in/yaml.v2"

	"github.com/dymensionxyz/dymension/v3/app/params"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyDisputePeriodInBlocks is store's key for DisputePeriodInBlocks Params
	KeyDisputePeriodInBlocks = []byte("DisputePeriodInBlocks")

	KeyLivenessSlashBlocks   = []byte("LivenessSlashBlocks")
	KeyLivenessSlashInterval = []byte("LivenessSlashInterval")
	KeyLivenessJailBlocks    = []byte("LivenessJailBlocks")

	// KeyAppRegistrationFee defines the key to store the cost of the app
	KeyAppRegistrationFee = []byte("KeyAppRegistrationFee")

	// DYM is 1dym
	DYM                       = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	DefaultAppRegistrationFee = sdk.NewCoin(params.BaseDenom, DYM)
)

const (
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1

	DefaultLivenessSlashBlocks   = uint64(7200)  // 12 hours at 1 block per 6 seconds
	DefaultLivenessSlashInterval = uint64(3600)  // 1 hour at 1 block per 6 seconds
	DefaultLivenessJailBlocks    = uint64(28800) // 48 hours at 1 block per 6 seconds
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	disputePeriodInBlocks uint64,
	livenessSlashBlocks uint64,
	livenessSlashInterval uint64,
	livenessJailBlocks uint64,
	appRegistrationFee sdk.Coin,
) Params {
	return Params{
		DisputePeriodInBlocks: disputePeriodInBlocks,
		LivenessSlashBlocks:   livenessSlashBlocks,
		LivenessSlashInterval: livenessSlashInterval,
		LivenessJailBlocks:    livenessJailBlocks,
		AppRegistrationFee:    appRegistrationFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDisputePeriodInBlocks,
		DefaultLivenessSlashBlocks,
		DefaultLivenessSlashInterval,
		DefaultLivenessJailBlocks,
		DefaultAppRegistrationFee,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyLivenessSlashBlocks, &p.LivenessSlashBlocks, validateLivenessSlashBlocks),
		paramtypes.NewParamSetPair(KeyLivenessSlashInterval, &p.LivenessSlashInterval, validateLivenessSlashInterval),
		paramtypes.NewParamSetPair(KeyLivenessJailBlocks, &p.LivenessJailBlocks, validateLivenessJailBlocks),
		paramtypes.NewParamSetPair(KeyAppRegistrationFee, &p.AppRegistrationFee, validateAppRegistrationFee),
	}
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

func (p Params) WithLivenessJailBlocks(x uint64) Params {
	p.LivenessJailBlocks = x
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

	if err := validateAppRegistrationFee(p.AppRegistrationFee); err != nil {
		return errorsmod.Wrap(err, "app registration fee")
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateLivenessSlashBlocks(i interface{}) error {
	return uparam.ValidatePositiveUint64(i)
}

func validateLivenessSlashInterval(i interface{}) error {
	return uparam.ValidatePositiveUint64(i)
}

func validateLivenessJailBlocks(i interface{}) error {
	return uparam.ValidatePositiveUint64(i)
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

func validateAppRegistrationFee(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !v.IsValid() {
		return fmt.Errorf("invalid app creation cost: %s", v)
	}

	return nil
}
