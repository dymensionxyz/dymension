package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	denommetadatamoduletypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	"strings"
)

var _ denommetadatamoduletypes.DenomMetadataHooks = VirtualFrontierBankContractRegistrationHook{}

type VirtualFrontierBankContractRegistrationHook struct {
	evmKeeper evmkeeper.Keeper
}

// NewVirtualFrontierBankContractRegistrationHook returns the DenomMetadataHooks for VFBC registration
func NewVirtualFrontierBankContractRegistrationHook(evmKeeper evmkeeper.Keeper) VirtualFrontierBankContractRegistrationHook {
	return VirtualFrontierBankContractRegistrationHook{
		evmKeeper: evmKeeper,
	}
}

func (v VirtualFrontierBankContractRegistrationHook) AfterDenomMetadataCreation(ctx sdk.Context, _ banktypes.Metadata) error {
	// Deploy the virtual frontier bank contract for all IBC denoms.
	// Error, if any, no state transition will be made and logged as error.
	v.evmKeeper.DeployVirtualFrontierBankContractForAllBankDenomMetadataRecords(ctx, func(metadata banktypes.Metadata) bool {
		return strings.HasPrefix(metadata.Base, "ibc/") // only deploy for IBC denoms
	})

	return nil
}

func (v VirtualFrontierBankContractRegistrationHook) AfterDenomMetadataUpdate(sdk.Context, banktypes.Metadata) error {
	// do nothing

	return nil
}
