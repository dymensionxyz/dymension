package types

import (
	"cosmossdk.io/math"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams takes an epoch distribution identifier, then returns an incentives Params struct.
func NewParams(distrEpochIdentifier string, createGaugeFee, addToGaugeFee, addDenomFee math.Int, minValueForDistr sdk.Coin, minLockAge, minLockDuration time.Duration, rollappGaugesMode Params_RollappGaugesModes) Params {
	return Params{
		DistrEpochIdentifier:    distrEpochIdentifier,
		CreateGaugeBaseFee:      createGaugeFee,
		AddToGaugeBaseFee:       addToGaugeFee,
		AddDenomFee:             addDenomFee,
		MinValueForDistribution: minValueForDistr,
		MinLockAge:              minLockAge,
		MinLockDuration:         minLockDuration,
		RollappGaugesMode:       rollappGaugesMode,
	}
}

// DefaultParams returns the default incentives module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier:    DefaultDistrEpochIdentifier,
		CreateGaugeBaseFee:      DefaultCreateGaugeFee,
		AddToGaugeBaseFee:       DefaultAddToGaugeFee,
		AddDenomFee:             DefaultAddDenomFee,
		MinValueForDistribution: DefaultMinValueForDistr,
		MinLockAge:              DefaultMinLockAge,
		MinLockDuration:         DefaultMinLockDuration,
		RollappGaugesMode:       DefaultRollappGaugesMode,
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) ValidateBasic() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}
	if err := validateCreateGaugeFee(p.CreateGaugeBaseFee); err != nil {
		return err
	}
	if err := validateAddToGaugeFee(p.AddToGaugeBaseFee); err != nil {
		return err
	}
	if err := validateAddDenomFee(p.AddDenomFee); err != nil {
		return err
	}

	if err := validateMinValueForDistr(p.MinValueForDistribution); err != nil {
		return err
	}

	if err := validateRollappGaugesMode(p.RollappGaugesMode); err != nil {
		return err
	}

	if p.MinLockAge < 0 {
		return gerrc.ErrInvalidArgument.Wrapf("min_lock_age must be >= 0, got %s", p.MinLockAge)
	}
	if p.MinLockDuration < 0 {
		return gerrc.ErrInvalidArgument.Wrapf("min_lock_duration must be >= 0, got %s", p.MinLockDuration)
	}

	return nil
}

func validateCreateGaugeFee(v math.Int) error {
	if v.IsNegative() {
		return gerrc.ErrInvalidArgument.Wrapf("must be >= 0, got %s", v)
	}
	return nil
}

func validateAddToGaugeFee(v math.Int) error {
	if v.IsNegative() {
		return gerrc.ErrInvalidArgument.Wrapf("must be >= 0, got %s", v)
	}
	return nil
}

func validateAddDenomFee(v math.Int) error {
	if v.IsNegative() {
		return gerrc.ErrInvalidArgument.Wrapf("must be >= 0, got %s", v)
	}
	return nil
}

func validateMinValueForDistr(_ sdk.Coin) error {
	return nil
}

func validateRollappGaugesMode(mode Params_RollappGaugesModes) error {
	if mode != Params_ActiveOnly && mode != Params_AllRollapps {
		return gerrc.ErrInvalidArgument.Wrapf("invalid RollappGaugesMode: %d", mode)
	}
	return nil
}
