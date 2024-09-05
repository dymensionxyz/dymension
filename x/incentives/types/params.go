package types

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Incentives parameters key store.
var (
	KeyDistrEpochIdentifier          = []byte("DistrEpochIdentifier")
	KeyCreateGaugeFee                = []byte("CreateGaugeFee")
	KeyAddToGaugeFee                 = []byte("AddToGaugeFee")
	KeyBaseGasFeeForCreateGauge      = []byte("BaseGasFeeForCreateGauge")
	KeyBaseGasFeeForAddRewardToGauge = []byte("BaseGasFeeForAddRewardToGauge")
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams takes an epoch distribution identifier, then returns an incentives Params struct.
func NewParams(
	distrEpochIdentifier string,
	createGaugeFee, addToGaugeFee math.Int,
	baseGasFeeForCreateGauge, baseGasFeeForAddRewardToGauge uint64,
) Params {
	return Params{
		DistrEpochIdentifier:          distrEpochIdentifier,
		CreateGaugeFee:                createGaugeFee,
		AddToGaugeFee:                 addToGaugeFee,
		BaseGasFeeForCreateGauge:      baseGasFeeForCreateGauge,
		BaseGasFeeForAddRewardToGauge: baseGasFeeForAddRewardToGauge,
	}
}

// DefaultParams returns the default incentives module parameters.
func DefaultParams() Params {
	return Params{
		DistrEpochIdentifier:          DefaultDistrEpochIdentifier,
		CreateGaugeFee:                DefaultCreateGaugeFee,
		AddToGaugeFee:                 DefaultAddToGaugeFee,
		BaseGasFeeForCreateGauge:      DefaultBaseGasFeeForCreateGauge,
		BaseGasFeeForAddRewardToGauge: DefaultBaseGasFeeForAddRewardToGauge,
	}
}

// Validate checks that the incentives module parameters are valid.
func (p Params) Validate() error {
	if err := epochtypes.ValidateEpochIdentifierInterface(p.DistrEpochIdentifier); err != nil {
		return err
	}
	if err := validateCreateGaugeFeeInterface(p.CreateGaugeFee); err != nil {
		return err
	}
	if err := validateAddToGaugeFeeInterface(p.AddToGaugeFee); err != nil {
		return err
	}
	if err := validateBaseGasFeeForCreateGauge(p.BaseGasFeeForCreateGauge); err != nil {
		return err
	}
	if err := validateBaseGasFeeForAddRewardToGauge(p.BaseGasFeeForAddRewardToGauge); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs takes the parameter struct and associates the paramsubspace key and field of the parameters as a KVStore.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyCreateGaugeFee, &p.CreateGaugeFee, validateCreateGaugeFeeInterface),
		paramtypes.NewParamSetPair(KeyAddToGaugeFee, &p.AddToGaugeFee, validateAddToGaugeFeeInterface),
		paramtypes.NewParamSetPair(KeyBaseGasFeeForCreateGauge, &p.BaseGasFeeForCreateGauge, validateBaseGasFeeForCreateGauge),
		paramtypes.NewParamSetPair(KeyBaseGasFeeForAddRewardToGauge, &p.BaseGasFeeForAddRewardToGauge, validateBaseGasFeeForAddRewardToGauge),
	}
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

func validateBaseGasFeeForCreateGauge(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateBaseGasFeeForAddRewardToGauge(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
