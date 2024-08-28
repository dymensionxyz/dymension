package keeper

import (
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
	for _, stateSigner := range genesisState.GetConsensusStateSigners() {
		k.SetConsensusStateSigner(ctx, stateSigner.IbcClientId, stateSigner.Height, []byte(stateSigner.Signer))
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	clients := k.GetAllCanonicalClients(ctx)
	stateSigners := k.GetAllConsensusStateSigners(ctx)
	return types.GenesisState{
		CanonicalClients:      clients,
		ConsensusStateSigners: stateSigners,
	}
}
