package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
func (p Params) ValidateBasic() error {
	if err := validateAddresses(p.ForceUnlockAllowedAddresses); err != nil {
		return err
	}
	if err := validateLockCreationFee(p.LockCreationFee); err != nil {
		return err
	}
	return nil
}

func validateAddresses(addresses []string) error {
	for _, address := range addresses {
		_, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateLockCreationFee(fee math.Int) error {
	if !fee.IsNil() && fee.IsNegative() {
		return fmt.Errorf("lock creation fee must be non-negative: %d", fee.Int64())
	}

	return nil
}
