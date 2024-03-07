package denommetadata

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	abci "github.com/tendermint/tendermint/abci/types"
	"strings"
)

// DeployVFBCForNativeCoin performs deployment of VFBC for native coin.
func DeployVFBCForNativeCoin(
	ctx sdk.Context,
	evmKeeper evmkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) (ret []abci.ValidatorUpdate) {
	ret = nil

	if ctx.BlockHeight() != 0 {
		// only deploy at genesis
		return
	}

	if ctx.ChainID() != "dymension_100-1" {
		// only deploy for local-net
		return
	}

	base := evmKeeper.GetParams(ctx).EvmDenom
	_, foundMetadata := bankKeeper.GetDenomMetaData(ctx, base)

	if !foundMetadata {
		display := strings.ToUpper(base[1:])
		metadata := banktypes.Metadata{
			Description: fmt.Sprintf("Native coin %s", display),
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

		bankKeeper.SetDenomMetaData(ctx, metadata)
	}

	// 0xbd8eff67ca469df5cd89f7a9b2890f043188d695 is the contract address of the first virtual frontier bank contract
	err := evmKeeper.DeployVirtualFrontierBankContractForBankDenomMetadataRecord(ctx, base)
	if err != nil {
		panic(err)
	}

	// the contract for native denom is disabled by default, so we need to enable it
	vfbcContractAddr, _ := evmKeeper.GetVirtualFrontierBankContractAddressByDenom(ctx, base)
	vfbcContract := evmKeeper.GetVirtualFrontierContract(ctx, vfbcContractAddr)
	vfbcContract.Active = true
	err = evmKeeper.SetVirtualFrontierContract(ctx, vfbcContractAddr, vfbcContract)
	if err != nil {
		panic(err)
	}

	return
}
