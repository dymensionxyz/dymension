package incentives

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "incentives"
)

// Params defines the parameters for the incentives module
type Params struct {
	DistrEpochIdentifier    string                    `json:"distr_epoch_identifier" yaml:"distr_epoch_identifier"`
	CreateGaugeBaseFee      math.Int                  `json:"create_gauge_base_fee" yaml:"create_gauge_base_fee"`
	AddToGaugeBaseFee       math.Int                  `json:"add_to_gauge_base_fee" yaml:"add_to_gauge_base_fee"`
	AddDenomFee             math.Int                  `json:"add_denom_fee" yaml:"add_denom_fee"`
	MinValueForDistribution sdk.Coin                  `json:"min_value_for_distribution" yaml:"min_value_for_distribution"`
	RollappGaugesMode       Params_RollappGaugesModes `json:"rollapp_gauges_mode" yaml:"rollapp_gauges_mode"`
}

// Params_RollappGaugesModes defines the modes for rollapp gauges
type Params_RollappGaugesModes int32

const (
	// Params_ACTIVE_ROLLAPPS_ONLY means only active rollapps will be included in gauges
	Params_ACTIVE_ROLLAPPS_ONLY Params_RollappGaugesModes = 0
	// Params_ALL_ROLLAPPS means all rollapps will be included in gauges
	Params_ALL_ROLLAPPS Params_RollappGaugesModes = 1
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
	KeyCreateGaugeFee       = []byte("CreateGaugeFee")
	KeyAddToGaugeFee        = []byte("AddToGaugeFee")
	KeyAddDenomFee          = []byte("AddDenomFee")
	KeyMinValueForDistr     = []byte("MinValueForDistr")
	KeyRollappGaugesMode    = []byte("RollappGaugesMode")
)

const (
	DefaultDistrEpochIdentifier = "hour"
	DefaultRollappGaugesMode    = Params_ACTIVE_ROLLAPPS_ONLY
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

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
		CreateGaugeBaseFee:      math.NewInt(1000000),
		AddToGaugeBaseFee:       math.NewInt(100000),
		AddDenomFee:             math.NewInt(100000),
		MinValueForDistribution: sdk.NewCoin("adym", math.NewInt(1000000)),
		RollappGaugesMode:       DefaultRollappGaugesMode,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, validateDistrEpochIdentifier),
		paramtypes.NewParamSetPair(KeyCreateGaugeFee, &p.CreateGaugeBaseFee, validateCreateGaugeFee),
		paramtypes.NewParamSetPair(KeyAddToGaugeFee, &p.AddToGaugeBaseFee, validateAddToGaugeFee),
		paramtypes.NewParamSetPair(KeyAddDenomFee, &p.AddDenomFee, validateAddDenomFee),
		paramtypes.NewParamSetPair(KeyMinValueForDistr, &p.MinValueForDistribution, validateMinValueForDistr),
		paramtypes.NewParamSetPair(KeyRollappGaugesMode, &p.RollappGaugesMode, validateRollappGaugesMode),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDistrEpochIdentifier(p.DistrEpochIdentifier); err != nil {
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
	return nil
}

func validateDistrEpochIdentifier(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == "" {
		return fmt.Errorf("distribution epoch identifier cannot be empty")
	}
	return nil
}

func validateCreateGaugeFee(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("create gauge fee cannot be negative")
	}
	return nil
}

func validateAddToGaugeFee(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("add to gauge fee cannot be negative")
	}
	return nil
}

func validateAddDenomFee(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("add denom fee cannot be negative")
	}
	return nil
}

func validateMinValueForDistr(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.Amount.IsNegative() {
		return fmt.Errorf("min value for distribution cannot be negative")
	}
	return nil
}

func validateRollappGaugesMode(i interface{}) error {
	v, ok := i.(Params_RollappGaugesModes)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v != Params_ACTIVE_ROLLAPPS_ONLY && v != Params_ALL_ROLLAPPS {
		return fmt.Errorf("invalid rollapp gauges mode")
	}
	return nil
}
