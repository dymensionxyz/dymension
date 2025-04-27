package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewParams(forceUnlockAllowedAddresses []string, lockCreationFee math.Int, minLockDuration time.Duration) Params {
	return Params{
		ForceUnlockAllowedAddresses: forceUnlockAllowedAddresses,
		LockCreationFee:             lockCreationFee,
		MinLockDuration:             minLockDuration,
	}
}

// DefaultParams returns default lockup module parameters.
func DefaultParams() Params {
	return Params{
		ForceUnlockAllowedAddresses: []string{},
		LockCreationFee:             DefaultLockFee,
		MinLockDuration:             0,
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

	if err := validateMinLockDuration(p.MinLockDuration); err != nil {
		return err
	}

	return nil
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

func validateMinLockDuration(i interface{}) error {
	duration, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if duration < 0 {
		return fmt.Errorf("duration should be non-negative: %d", duration)
	}

	return nil
}
