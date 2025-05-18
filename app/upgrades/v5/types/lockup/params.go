package lockup

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "lockup"
)

type Params struct {
	ForceUnlockAllowedAddresses []string `protobuf:"bytes,1,rep,name=force_unlock_allowed_addresses,json=forceUnlockAllowedAddresses,proto3" json:"force_unlock_allowed_addresses,omitempty" yaml:"force_unlock_allowed_address"`
}

// Parameter store keys.
var (
	KeyForceUnlockAllowedAddresses = []byte("ForceUnlockAllowedAddresses")

	_ paramtypes.ParamSet = &Params{}
)

// ParamKeyTable for lockup module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Validate validates params.
func (p Params) Validate() error {
	if err := validateAddresses(p.ForceUnlockAllowedAddresses); err != nil {
		return err
	}
	return nil
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyForceUnlockAllowedAddresses, &p.ForceUnlockAllowedAddresses, validateAddresses),
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
