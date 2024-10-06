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

	// FIXME: validate metadata corresponding to the genesis info denom

	// Validate genesis transfers
	if data.GenesisTransfer != nil {
		if err := data.GenesisTransfer.ValidateBasic(); err != nil {
			return errors.Wrap(err, "invalid genesis transfer")
		}
	}

	return nil
}

// helper function to convert GenesisBridgeInfo to GenesisInfo
func (info GenesisBridgeInfo) ToRAGenesisInfo() types.GenesisInfo {
	return types.GenesisInfo{
		GenesisChecksum: info.GenesisChecksum,
		Bech32Prefix:    info.Bech32Prefix,
		NativeDenom:     info.NativeDenom,
		InitialSupply:   info.InitialSupply,
		GenesisAccounts: info.GenesisAccounts,
	}
}

// ValidateBasic performs basic validation checks on the GenesisInfo.
func (info GenesisBridgeInfo) ValidateBasic() error {
	raGenesisInfo := info.ToRAGenesisInfo()

	if !raGenesisInfo.AllSet() {
		return fmt.Errorf("missing fields in genesis info")
	}

	return raGenesisInfo.Validate()
}
