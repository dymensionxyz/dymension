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

	keeper.InitGenesis(ctx, types.GenesisState{
		CanonicalClients: clients,
	})

	ibc, found := keeper.GetCanonicalClient(ctx, "rollapp-1")
	require.True(t, found)
	require.Equal(t, "client-1", ibc)
	ibc, found = keeper.GetCanonicalClient(ctx, "rollapp-2")
	require.True(t, found)
	require.Equal(t, "client-2", ibc)
}

func TestExportGenesis(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)

	keeper.SetCanonicalClient(ctx, "rollapp-1", "client-1")
	keeper.SetCanonicalClient(ctx, "rollapp-2", "client-2")

	genesis := keeper.ExportGenesis(ctx)

	require.Len(t, genesis.CanonicalClients, 2)
	require.Equal(t, "client-1", genesis.CanonicalClients[0].IbcClientId)
	require.Equal(t, "client-2", genesis.CanonicalClients[1].IbcClientId)
	require.Equal(t, "rollapp-1", genesis.CanonicalClients[0].RollappId)
	require.Equal(t, "rollapp-2", genesis.CanonicalClients[1].RollappId)
}
