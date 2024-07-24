package types

import (
	"errors"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
	KeyDisputePeriodInBlocks            = []byte("DisputePeriodInBlocks")
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	disputePeriodInBlocks uint64,
	registrationFee sdk.Coin,
) Params {
	return Params{
		DisputePeriodInBlocks: disputePeriodInBlocks,
		RegistrationFee:       registrationFee,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDisputePeriodInBlocks, DefaultRegistrationFee)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyRegistrationFee, &p.RegistrationFee, validateRegistrationFee),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return err
	}

	return validateRegistrationFee(p.RegistrationFee)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
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
		return fmt.Errorf("invalid registration fee: %s", registrationFee)
	}

	if registrationFee.Denom != appparams.BaseDenom {
		return fmt.Errorf("invalid registration fee denom: %s", registrationFee.Denom)
	}

	return nil
}
