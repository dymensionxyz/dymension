package types

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams takes an epoch distribution identifier, then returns an incentives Params struct.
func NewParams(distrEpochIdentifier string, createGaugeFee, addToGaugeFee, addDenomFee math.Int, minValueForDistr sdk.Coin, rollappGaugesMode Params_RollappGaugesModes) Params {
	return Params{
		DistrEpochIdentifier:    distrEpochIdentifier,
		CreateGaugeBaseFee:      createGaugeFee,
		AddToGaugeBaseFee:       addToGaugeFee,
		AddDenomFee:             addDenomFee,
		MinValueForDistribution: minValueForDistr,
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
		RollappGaugesMode:       DefaultRollappGaugesMode,
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}
	if err := validateCreateGaugeFeeInterface(p.CreateGaugeBaseFee); err != nil {
		return err
	}
	if err := validateAddToGaugeFeeInterface(p.AddToGaugeBaseFee); err != nil {
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
	return nil
}

func validateCreateGaugeFeeInterface(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return gerrc.ErrInvalidArgument.Wrapf("must be >= 0, got %s", v)
	}
	return nil
}

func validateAddToGaugeFeeInterface(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return gerrc.ErrInvalidArgument.Wrapf("must be >= 0, got %s", v)
	}
	return nil
}

func validateAddDenomFee(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return gerrc.ErrInvalidArgument.Wrapf("must be >= 0, got %s", v)
	}
	return nil
}

func validateMinValueForDistr(i interface{}) error {
	_, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateRollappGaugesMode(i interface{}) error {
	_, ok := i.(Params_RollappGaugesModes)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
