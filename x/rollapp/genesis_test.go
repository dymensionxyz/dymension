package rollapp_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		RollappList: []types.Rollapp{
			{
				RollappId: "0",
			},
			{
				RollappId: "1",
			},
		},
		StateInfoList: []types.StateInfo{
			{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: "0",
					Index:     0,
				},
			},
			{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: "1",
					Index:     1,
				},
			},
		},
		LatestStateInfoIndexList: []types.StateInfoIndex{
			{
				RollappId: "0",
			},
			{
				RollappId: "1",
			},
		},
		BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{
			{
				CreationHeight: 0,
			},
			{
				CreationHeight: 1,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.RollappList, got.RollappList)
	require.ElementsMatch(t, genesisState.StateInfoList, got.StateInfoList)
	require.ElementsMatch(t, genesisState.LatestStateInfoIndexList, got.LatestStateInfoIndexList)
	require.ElementsMatch(t, genesisState.BlockHeightToFinalizationQueueList, got.BlockHeightToFinalizationQueueList)
	// this line is used by starport scaffolding # genesis/test/assert
}

func TestExportGenesis(t *testing.T) {
	params := types.Params{
		DisputePeriodInBlocks: 11,
		DeployerWhitelist:     []types.DeployerParams{{Address: "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"}},
		RollappsEnabled:       false,
	}
	rollappList := []types.Rollapp{{RollappId: "0"}, {RollappId: "1"}}
	stateInfoList := []types.StateInfo{
		{StateInfoIndex: types.StateInfoIndex{RollappId: "0", Index: 0}},
		{StateInfoIndex: types.StateInfoIndex{RollappId: "1", Index: 1}},
	}
	latestStateInfoIndexList := []types.StateInfoIndex{{RollappId: "0"}, {RollappId: "1"}}
	blockHeightToFinalizationQueueList := []types.BlockHeightToFinalizationQueue{{CreationHeight: 0}, {CreationHeight: 1}}
	// Set the items in the keeper
	k, ctx := keepertest.RollappKeeper(t)
	for _, rollapp := range rollappList {
		k.SetRollapp(ctx, rollapp)
	}
	for _, stateInfo := range stateInfoList {
		k.SetStateInfo(ctx, stateInfo)
	}
	for _, latestStateInfoIndex := range latestStateInfoIndexList {
		k.SetLatestStateInfoIndex(ctx, latestStateInfoIndex)
	}
	for _, blockHeightToFinalizationQueue := range blockHeightToFinalizationQueueList {
		k.SetBlockHeightToFinalizationQueue(ctx, blockHeightToFinalizationQueue)
	}
	k.SetParams(ctx, params)
	// Verify the exported genesis state
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	// Validate the exported genesis state
	require.Equal(t, params, got.Params)
	require.ElementsMatch(t, rollappList, got.RollappList)
	require.ElementsMatch(t, stateInfoList, got.StateInfoList)
	require.ElementsMatch(t, latestStateInfoIndexList, got.LatestStateInfoIndexList)
	require.ElementsMatch(t, blockHeightToFinalizationQueueList, got.BlockHeightToFinalizationQueueList)
}
