package genesisbridge

import (
	fmt "fmt"

	"cosmossdk.io/errors"
	types "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// ValidateBasic performs basic validation checks on the GenesisBridgeData.
func (data GenesisBridgeData) ValidateBasic() error {
	// Validate genesis info
	if err := data.GenesisInfo.ValidateBasic(); err != nil {
		return errors.Wrap(err, "invalid genesis info")
	}

	// Validate metadata
	if err := data.NativeDenom.Validate(); err != nil {
		return errors.Wrap(err, "invalid metadata")
	}

	// validate metadata corresponding to the genesis info denom
	// check the base denom, display unit and decimals
	if data.NativeDenom.Base != data.GenesisInfo.NativeDenom.Base {
		return fmt.Errorf("metadata denom does not match genesis info denom")
	}
	// validate the decimals of the display denom
	valid := false
	for _, unit := range data.NativeDenom.DenomUnits {
		if unit.Denom == data.GenesisInfo.NativeDenom.Display {
			if unit.Exponent == data.GenesisInfo.NativeDenom.Exponent {
				valid = true
				break
			}
		}
	}
	if !valid {
		return fmt.Errorf("denom metadata does not contain display unit with the correct exponent")
	}

	// Validate genesis transfers
	if data.GenesisTransfer != nil {
		if err := data.GenesisTransfer.ValidateBasic(); err != nil {
			return errors.Wrap(err, "invalid genesis transfer")
		}
	}

	return nil
}

// ValidateBasic performs basic validation checks on the GenesisInfo.
func (info GenesisBridgeInfo) ValidateBasic() error {
	// wrap the genesis info in a GenesisInfo struct, to reuse the validation logic
	raGenesisInfo := types.GenesisInfo{
		GenesisChecksum: info.GenesisChecksum,
		Bech32Prefix:    info.Bech32Prefix,
		NativeDenom:     info.NativeDenom,
		InitialSupply:   info.InitialSupply,
		GenesisAccounts: info.GenesisAccounts,
	}

	if !raGenesisInfo.AllSet() {
		return fmt.Errorf("missing fields in genesis info")
	}

	return raGenesisInfo.Validate()
}
