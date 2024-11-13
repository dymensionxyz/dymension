package keeper_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

func TestInitGenesis(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	clients := []types.CanonicalClient{
		{RollappId: "rollapp-1", IbcClientId: "client-1"},
		{RollappId: "rollapp-2", IbcClientId: "client-2"},
	}

	keeper.InitGenesis(ctx, types.GenesisState{
		CanonicalClients: clients,
		HardForkKeys:     []string{"rollapp-1", "rollapp-2"},
	})

	ibc, found := keeper.GetCanonicalClient(ctx, "rollapp-1")
	require.True(t, found)
	require.Equal(t, "client-1", ibc)
	ibc, found = keeper.GetCanonicalClient(ctx, "rollapp-2")
	require.True(t, found)
	require.Equal(t, "client-2", ibc)
	hfks := keeper.ListHardForkKeys(ctx)
	require.Len(t, hfks, 2)
	require.Equal(t, "rollapp-1", hfks[0])
	require.Equal(t, "rollapp-2", hfks[1])
}

func TestExportGenesis(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)

	keeper.SetCanonicalClient(ctx, "rollapp-1", "client-1")
	keeper.SetCanonicalClient(ctx, "rollapp-2", "client-2")
	keeper.SetHardForkInProgress(ctx, "rollapp-1")
	keeper.SetHardForkInProgress(ctx, "rollapp-2")

	genesis := keeper.ExportGenesis(ctx)

	require.Len(t, genesis.CanonicalClients, 2)
	require.Equal(t, "client-1", genesis.CanonicalClients[0].IbcClientId)
	require.Equal(t, "client-2", genesis.CanonicalClients[1].IbcClientId)
	require.Equal(t, "rollapp-1", genesis.CanonicalClients[0].RollappId)
	require.Equal(t, "rollapp-2", genesis.CanonicalClients[1].RollappId)
	require.Len(t, genesis.HardForkKeys, 2)
	require.Equal(t, "rollapp-1", genesis.HardForkKeys[0])
	require.Equal(t, "rollapp-2", genesis.HardForkKeys[1])
}

func TestImportExportGenesis(t *testing.T) {
	k, ctx := keepertest.LightClientKeeper(t)

	g := types.GenesisState{
		CanonicalClients: []types.CanonicalClient{
			{
				RollappId:   "rollapp-1",
				IbcClientId: "client-1",
			},
			{
				RollappId:   "rollapp-2",
				IbcClientId: "client-2",
			},
		},
		HeaderSigners: []types.HeaderSignerEntry{
			{
				SequencerAddress: "signer-1",
				ClientId:         "client-1",
				Height:           42,
			},
			{
				SequencerAddress: "signer-2",
				ClientId:         "client-2",
				Height:           43,
			},
		},
		HardForkKeys: []string{"rollapp-1", "rollapp-2"},
	}

	k.InitGenesis(ctx, g)
	compare := k.ExportGenesis(ctx)
	require.True(t, reflect.DeepEqual(g, compare), "expected %v but got %v", g, compare)
}
