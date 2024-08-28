package keeper_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	clients := []types.CanonicalClient{
		{RollappId: "rollapp-1", IbcClientId: "client-1"},
		{RollappId: "rollapp-2", IbcClientId: "client-2"},
	}
	stateSigners := []types.ConsensusStateSigner{
		{IbcClientId: "client-1", Height: 1, Signer: "signer-1"},
		{IbcClientId: "client-1", Height: 2, Signer: "signer-1"},
	}

	keeper.InitGenesis(ctx, types.GenesisState{
		CanonicalClients:      clients,
		ConsensusStateSigners: stateSigners,
	})

	ibc, found := keeper.GetCanonicalClient(ctx, "rollapp-1")
	require.True(t, found)
	require.Equal(t, "client-1", ibc)
	ibc, found = keeper.GetCanonicalClient(ctx, "rollapp-2")
	require.True(t, found)
	require.Equal(t, "client-2", ibc)

	signer, found := keeper.GetConsensusStateSigner(ctx, "client-1", 1)
	require.True(t, found)
	require.Equal(t, "signer-1", string(signer))
	signer, found = keeper.GetConsensusStateSigner(ctx, "client-1", 2)
	require.True(t, found)
	require.Equal(t, "signer-1", string(signer))
}

func TestExportGenesis(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)

	keeper.SetCanonicalClient(ctx, "rollapp-1", "client-1")
	keeper.SetCanonicalClient(ctx, "rollapp-2", "client-2")
	keeper.SetConsensusStateSigner(ctx, "client-1", 1, []byte("signer-1"))
	keeper.SetConsensusStateSigner(ctx, "client-1", 2, []byte("signer-1"))

	genesis := keeper.ExportGenesis(ctx)

	require.Len(t, genesis.CanonicalClients, 2)
	require.Equal(t, "client-1", genesis.CanonicalClients[0].IbcClientId)
	require.Equal(t, "client-2", genesis.CanonicalClients[1].IbcClientId)
	require.Equal(t, "rollapp-1", genesis.CanonicalClients[0].RollappId)
	require.Equal(t, "rollapp-2", genesis.CanonicalClients[1].RollappId)
	require.Len(t, genesis.ConsensusStateSigners, 2)
	require.Equal(t, "client-1", genesis.ConsensusStateSigners[0].IbcClientId)
	require.Equal(t, "client-1", genesis.ConsensusStateSigners[1].IbcClientId)
	require.Equal(t, uint64(1), genesis.ConsensusStateSigners[0].Height)
	require.Equal(t, uint64(2), genesis.ConsensusStateSigners[1].Height)
	require.Equal(t, "signer-1", genesis.ConsensusStateSigners[0].Signer)
	require.Equal(t, "signer-1", genesis.ConsensusStateSigners[1].Signer)
}
