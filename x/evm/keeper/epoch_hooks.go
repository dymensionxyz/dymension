package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	"strings"
)

// EvmEpochHooks is the wrapper struct for the epoch-hook evm keeper.
type EvmEpochHooks struct {
	ek evmkeeper.Keeper
	bk bankkeeper.Keeper
}

var _ epochstypes.EpochHooks = EvmEpochHooks{}

// NewEvmEpochHooks returns the epoch-hook wrapper struct.
func NewEvmEpochHooks(ek evmkeeper.Keeper, bk bankkeeper.Keeper) EvmEpochHooks {
	return EvmEpochHooks{ek, bk}
}

/* -------------------------------------------------------------------------- */
/*                                 epoch hooks                                */
/* -------------------------------------------------------------------------- */

const triggerVirtualFrontierBankContractRegistrationAtEpochIdentifier = "day"

// BeforeEpochStart is the epoch start hook.
func (h EvmEpochHooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	var err error

	err = h.deployVirtualFrontierBankContractsForNewNetwork(ctx)
	if err != nil {
		return err
	}

	if epochIdentifier == triggerVirtualFrontierBankContractRegistrationAtEpochIdentifier {
		err = h.ek.DeployVirtualFrontierBankContractForAllBankDenomMetadataRecords(ctx, func(metadata banktypes.Metadata) bool {
			return strings.HasPrefix(metadata.Base, "ibc/") // only deploy for IBC denoms
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// AfterEpochEnd is the epoch end hook.
func (h EvmEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	// no-op
	return nil
}

// deployVirtualFrontierBankContractsForNewNetwork attempts to deploy virtual frontier bank contracts
// for ALL denom metadata added to the bank module at the genesis.
// If the denom metadata for the native coin is not available, it will be added automatically before deployment.
// The execution only happens once at the first block of the chain.
func (h EvmEpochHooks) deployVirtualFrontierBankContractsForNewNetwork(ctx sdk.Context) error {
	// TODO: consider comment entire this method, as it was added for local-net testing

	if ctx.BlockHeight() != 1 {
		// trigger for new network only
		return nil
	}

	base := h.ek.GetParams(ctx).EvmDenom
	_, foundMetadata := h.bk.GetDenomMetaData(ctx, base)
	if !foundMetadata {
		display := strings.ToUpper(base[1:])
		metadata := banktypes.Metadata{
			Description: display,
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    base,
					Exponent: 0,
				},
				{
					Denom:    display,
					Exponent: 18,
				},
			},
			Base:    base,
			Display: display,
			Name:    display,
			Symbol:  display,
		}

		h.bk.SetDenomMetaData(ctx, metadata)
	}

	// 0xbd8eff67ca469df5cd89f7a9b2890f043188d695 is the contract address of the first virtual frontier bank contract
	err := h.ek.DeployVirtualFrontierBankContractForAllBankDenomMetadataRecords(ctx, func(banktypes.Metadata) bool {
		// deploy all
		return true
	})
	if err != nil {
		return err
	}

	// the contract for native denom is disabled by default, so we need to enable it
	vfbcContractAddr, found := h.ek.GetVirtualFrontierBankContractAddressByDenom(ctx, base)
	if !found {
		return fmt.Errorf("is the metadata for the native coin added and valid? %s", base)
	}
	vfbcContract := h.ek.GetVirtualFrontierContract(ctx, vfbcContractAddr)
	vfbcContract.Active = true
	err = h.ek.SetVirtualFrontierContract(ctx, vfbcContractAddr, vfbcContract)
	if err != nil {
		return err
	}

	return nil
}
