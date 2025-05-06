package lockup

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "lockup"
)

// Params defines the parameters for the lockup module
type Params struct {
	ForceUnlockAllowedAddresses []string      `json:"force_unlock_allowed_addresses" yaml:"force_unlock_allowed_addresses"`
	LockCreationFee             math.Int      `json:"lock_creation_fee" yaml:"lock_creation_fee"`
	MinLockDuration             time.Duration `json:"min_lock_duration" yaml:"min_lock_duration"`
}

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyForceUnlockAllowedAddresses = []byte("ForceUnlockAllowedAddresses")
	KeyLockCreationFee             = []byte("LockCreationFee")
	KeyMinLockDuration             = []byte("MinLockDuration")
)

// ParamKeyTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(forceUnlockAllowedAddresses []string, lockCreationFee math.Int, minLockDuration time.Duration) Params {
	return Params{
		ForceUnlockAllowedAddresses: forceUnlockAllowedAddresses,
		LockCreationFee:             lockCreationFee,
		MinLockDuration:             minLockDuration,
	}
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyForceUnlockAllowedAddresses, &p.ForceUnlockAllowedAddresses, validateAddresses),
		paramtypes.NewParamSetPair(KeyLockCreationFee, &p.LockCreationFee, validateLockCreationFee),
		paramtypes.NewParamSetPair(KeyMinLockDuration, &p.MinLockDuration, validateMinLockDuration),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
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
	v, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, addr := range v {
		if addr == "" {
			return fmt.Errorf("address cannot be empty")
		}
	}
	return nil
}

func validateLockCreationFee(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("lock creation fee cannot be negative")
	}
	return nil
}

func validateMinLockDuration(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v < 0 {
		return fmt.Errorf("min lock duration cannot be negative")
	}
	return nil
}
