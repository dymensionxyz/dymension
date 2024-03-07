package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethermintutils "github.com/evmos/ethermint/utils"
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
	if err := h.deployAVirtualFrontierBankSmartContractOnLocalNet(ctx); err != nil {
		return err
	}

	if epochIdentifier == triggerVirtualFrontierBankContractRegistrationAtEpochIdentifier {
		return h.ek.DeployVirtualFrontierBankContractForAllBankDenomMetadataRecords(ctx, nil)
	}

	return nil
}

// AfterEpochEnd is the epoch end hook.
func (h EvmEpochHooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func (h EvmEpochHooks) deployAVirtualFrontierBankSmartContractOnLocalNet(ctx sdk.Context) error {
	if ctx.BlockHeight() != 1 {
		return nil
	}

	isLocalNet := ctx.ChainID() == "dymension_100-1" || !ethermintutils.IsOneOfDymensionChains(ctx)
	if !isLocalNet {
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

	_, foundContract := h.ek.GetVirtualFrontierBankContractAddressByDenom(ctx, base)
	if foundContract {
		return nil
	}

	// 0xbd8eff67ca469df5cd89f7a9b2890f043188d695 is the contract address of the first virtual frontier bank contract
	err := h.ek.DeployVirtualFrontierBankContractForAllBankDenomMetadataRecords(ctx, func(banktypes.Metadata) bool {
		// deploy all
		return true
	})
	if err != nil {
		return err
	}

	// the contract for native denom is disabled by default so we need to enable it
	contractAddress, _ := h.ek.GetVirtualFrontierBankContractAddressByDenom(ctx, base)
	vfContract := h.ek.GetVirtualFrontierContract(ctx, contractAddress)
	vfContract.Active = true
	err = h.ek.SetVirtualFrontierContract(ctx, contractAddress, vfContract)
	if err != nil {
		return err
	}

	return nil
}
