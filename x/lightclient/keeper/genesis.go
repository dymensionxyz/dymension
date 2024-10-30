package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

// TODO: add signer bookkeeping to genesis and do import/export

func (k Keeper) InitGenesis(ctx sdk.Context, genesisState types.GenesisState) {
	if err := genesisState.Validate(); err != nil {
		panic(err)
	}
	for _, client := range genesisState.GetCanonicalClients() {
		k.SetCanonicalClient(ctx, client.RollappId, client.IbcClientId)
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	clients := k.GetAllCanonicalClients(ctx)
	return types.GenesisState{
		CanonicalClients: clients,
	}
}
