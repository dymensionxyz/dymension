package rollapp_test

import (
	"bytes"
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

var (
	blockDescriptors = types.BlockDescriptors{BD: []types.BlockDescriptor{
		{
			Height:                 1,
			StateRoot:              bytes.Repeat([]byte{byte(1)}, 32),
			IntermediateStatesRoot: bytes.Repeat([]byte{byte(1)}, 32),
		},
	}}
	genesisState = types.GenesisState{
		Params: types.DefaultParams(),

		RollappList: []types.Rollapp{
			{
				RollappId: "0",
				Creator:   sample.AccAddress(),
			},
			{
				RollappId: "1",
				Creator:   sample.AccAddress(),
			},
		},
		StateInfoList: []types.StateInfo{
			{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: "0",
					Index:     1,
				},
				Sequencer:   sample.AccAddress(),
				NumBlocks:   1,
				StartHeight: 1,
				BDs:         blockDescriptors,
			},
			{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: "1",
					Index:     1,
				},
				Sequencer:   sample.AccAddress(),
				NumBlocks:   1,
				StartHeight: 1,
				BDs:         blockDescriptors,
			},
		},
		LatestStateInfoIndexList: []types.StateInfoIndex{
			{
				RollappId: "0",
				Index:     1,
			},
			{
				RollappId: "1",
				Index:     1,
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
)

func TestValidGenesis(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	require.ElementsMatch(t, genesisState.RollappList, got.RollappList)
	require.ElementsMatch(t, genesisState.StateInfoList, got.StateInfoList)
	require.ElementsMatch(t, genesisState.LatestStateInfoIndexList, got.LatestStateInfoIndexList)
	require.ElementsMatch(t, genesisState.BlockHeightToFinalizationQueueList, got.BlockHeightToFinalizationQueueList)
	// this line is used by starport scaffolding # genesis/test/assert
}

func TestMissingRollappInfoGenesis(t *testing.T) {
	genesisState.RollappList[0].Creator = ""
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.RollappList, got.RollappList)
}

func TestDuplicatedRollappInitGenesis(t *testing.T) {
	genesisState.RollappList[1] = types.Rollapp{
		RollappId: "0",
		Creator:   sample.AccAddress(),
	}

	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.RollappList, got.RollappList)
}

func TestInvalidRollappIdInitGenesis(t *testing.T) {
	genesisState.RollappList[0].RollappId = ""
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.RollappList, got.RollappList)
}
