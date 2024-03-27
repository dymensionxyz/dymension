package hooks

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	denommetadatamoduletypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
)

var _ denommetadatamoduletypes.DenomMetadataHooks = VirtualFrontierBankContractRegistrationHook{}

const ibcDenomPrefix = "ibc/"

type VirtualFrontierBankContractRegistrationHook struct {
	evmKeeper evmkeeper.Keeper
}

// NewVirtualFrontierBankContractRegistrationHook returns the DenomMetadataHooks for VFBC registration
func NewVirtualFrontierBankContractRegistrationHook(evmKeeper evmkeeper.Keeper) VirtualFrontierBankContractRegistrationHook {
	return VirtualFrontierBankContractRegistrationHook{
		evmKeeper: evmKeeper,
	}
}

func (v VirtualFrontierBankContractRegistrationHook) AfterDenomMetadataCreation(ctx sdk.Context, newDenomMetadata banktypes.Metadata) error {
	if strings.HasPrefix(strings.ToLower(newDenomMetadata.Base), ibcDenomPrefix) { // only deploy for IBC denom.
		// Deploy the virtual frontier bank contract for the new IBC denom.
		// Error, if any, no state transition will be made.
		if err := v.evmKeeper.DeployVirtualFrontierBankContractForBankDenomMetadataRecord(ctx, newDenomMetadata.Base); err != nil {
			return fmt.Errorf("deploy virtual frontier bank contract for IBC denom %s: %w", newDenomMetadata.Base, err)
		}
	}

	return nil
}

func (v VirtualFrontierBankContractRegistrationHook) AfterDenomMetadataUpdate(sdk.Context, banktypes.Metadata) error {
	// do nothing

	return nil
}
