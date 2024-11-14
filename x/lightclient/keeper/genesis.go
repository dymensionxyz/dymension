package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genesisState types.GenesisState) {
	if err := genesisState.Validate(); err != nil {
		panic(err)
	}
	for _, client := range genesisState.GetCanonicalClients() {
		k.SetCanonicalClient(ctx, client.RollappId, client.IbcClientId)
	}
	for _, signer := range genesisState.HeaderSigners {
		if err := k.SaveSigner(ctx, signer.SequencerAddress, signer.ClientId, signer.Height); err != nil {
			panic(err)
		}
	}
	for _, rollappID := range genesisState.HardForkKeys {
		k.SetHardForkInProgress(ctx, rollappID)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	clients := k.GetAllCanonicalClients(ctx)
	hardForkKeys := k.ListHardForkKeys(ctx)

	ret := types.GenesisState{
		CanonicalClients: clients,
		HardForkKeys:     hardForkKeys,
	}

	if err := k.headerSigners.Walk(ctx, nil,
		func(key collections.Triple[string, string, uint64]) (stop bool, err error) {
			ret.HeaderSigners = append(ret.HeaderSigners, types.HeaderSignerEntry{
				SequencerAddress: key.K1(),
				ClientId:         key.K2(),
				Height:           key.K3(),
			})
			return false, nil
		}); err != nil {
		panic(err)
	}
	return ret
}
