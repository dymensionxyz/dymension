package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyForceUnlockAllowedAddresses = []byte("ForceUnlockAllowedAddresses")
	KeyLockCreationFee             = []byte("LockCreationFee")

	_ paramtypes.ParamSet = &Params{}
)

// ParamKeyTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(forceUnlockAllowedAddresses []string, lockCreationFee math.Int) Params {
	return Params{
		ForceUnlockAllowedAddresses: forceUnlockAllowedAddresses,
		LockCreationFee:             lockCreationFee,
	}
}

// DefaultParams returns default lockup module parameters.
func DefaultParams() Params {
	return Params{
		ForceUnlockAllowedAddresses: []string{},
		LockCreationFee:             DefaultLockFee,
	}
}

// Validate validates params.
func (p Params) Validate() error {
	if err := validateAddresses(p.ForceUnlockAllowedAddresses); err != nil {
		return err
	}
	if err := validateLockCreationFee(p.LockCreationFee); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyForceUnlockAllowedAddresses, &p.ForceUnlockAllowedAddresses, validateAddresses),
		paramtypes.NewParamSetPair(KeyLockCreationFee, &p.LockCreationFee, validateLockCreationFee),
	}
}

func validateAddresses(i interface{}) error {
	addresses, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, address := range addresses {
		_, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateLockCreationFee(i interface{}) error {
	fee, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if !fee.IsNil() && fee.IsNegative() {
		return fmt.Errorf("lock creation fee must be non-negative: %d", fee.Int64())
	}

	return nil
}
