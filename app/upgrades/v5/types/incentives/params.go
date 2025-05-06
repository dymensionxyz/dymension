package incentives

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	epochtypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	ModuleName = "incentives"
)

// Params holds parameters for the incentives module
type Params struct {
	// distr_epoch_identifier is what epoch type distribution will be triggered by
	// (day, week, etc.)
	DistrEpochIdentifier string `protobuf:"bytes,1,opt,name=distr_epoch_identifier,json=distrEpochIdentifier,proto3" json:"distr_epoch_identifier,omitempty" yaml:"distr_epoch_identifier"`
	// CreateGaugeBaseFee is a base fee required to create a new gauge. The final
	// fee is calculated as
	// Fee = CreateGaugeBaseFee + AddDenomFee * (len(Denoms) + len(GaugeDenoms)).
	CreateGaugeBaseFee math.Int `protobuf:"bytes,2,opt,name=create_gauge_base_fee,json=createGaugeBaseFee,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"create_gauge_base_fee"`
	// AddToGaugeBaseFee is a base fee required to add to gauge. The final
	// fee is calculated as
	// Fee = AddToGaugeBaseFee + AddDenomFee * len(Denoms).
	AddToGaugeBaseFee math.Int `protobuf:"bytes,3,opt,name=add_to_gauge_base_fee,json=addToGaugeBaseFee,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"add_to_gauge_base_fee"`
	// AddDenomFee is a fee charged for adding every new denom to the gauge.
	AddDenomFee math.Int `protobuf:"bytes,4,opt,name=add_denom_fee,json=addDenomFee,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Int" json:"add_denom_fee"`
}

// Incentives parameters key store.
var (
	KeyDistrEpochIdentifier = []byte("DistrEpochIdentifier")
	KeyCreateGaugeFee       = []byte("CreateGaugeFee")
	KeyAddToGaugeFee        = []byte("AddToGaugeFee")
	KeyAddDenomFee          = []byte("AddDenomFee")
)

// ParamKeyTable returns the key table for the incentive module's parameters.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
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
	return nil
}

// ParamSetPairs takes the parameter struct and associates the paramsubspace key and field of the parameters as a KVStore.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDistrEpochIdentifier, &p.DistrEpochIdentifier, epochtypes.ValidateEpochIdentifierInterface),
		paramtypes.NewParamSetPair(KeyCreateGaugeFee, &p.CreateGaugeBaseFee, validateCreateGaugeFeeInterface),
		paramtypes.NewParamSetPair(KeyAddToGaugeFee, &p.AddToGaugeBaseFee, validateAddToGaugeFeeInterface),
		paramtypes.NewParamSetPair(KeyAddDenomFee, &p.AddDenomFee, validateAddDenomFee),
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
