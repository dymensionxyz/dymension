package rollapp_test

import (
	"bytes"
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
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
	blockDescriptors2 = types.BlockDescriptors{BD: []types.BlockDescriptor{
		{
			Height:                 2,
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
					Index:     2,
				},
				Sequencer:   sample.AccAddress(),
				NumBlocks:   1,
				StartHeight: 2,
				BDs:         blockDescriptors2,
				Status:      common.Status_PENDING,
			},
			{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: "1",
					Index:     2,
				},
				Sequencer:   sample.AccAddress(),
				NumBlocks:   1,
				StartHeight: 2,
				BDs:         blockDescriptors2,
				Status:      common.Status_PENDING,
			},
			{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: "0",
					Index:     1,
				},
				Sequencer:   sample.AccAddress(),
				NumBlocks:   1,
				StartHeight: 1,
				BDs:         blockDescriptors,
				Status:      common.Status_FINALIZED,
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
				Status:      common.Status_FINALIZED,
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

func TestFailedValidationStateInfoInitGenesis(t *testing.T) {
	genesisState.StateInfoList[0].Sequencer = ""
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
}

func TestMissingRollappForStateInfoInitGenesis(t *testing.T) {
	genesisState.RollappList[0] = types.Rollapp{}
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
}

func TestStateInfoVersionMismatchInitGenesis(t *testing.T) {
	genesisState.RollappList[0].Version = 1
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
}

func TestMissingStateInfoInitGenesis(t *testing.T) {
	genesisState.StateInfoList[3] = types.StateInfo{}
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
}

func TestStateInfoWithMissingBlocksInitGenesis(t *testing.T) {
	bd := types.BlockDescriptors{
		BD: []types.BlockDescriptor{{
			Height:                 3,
			StateRoot:              bytes.Repeat([]byte{byte(1)}, 32),
			IntermediateStatesRoot: bytes.Repeat([]byte{byte(1)}, 32),
		}},
	}
	genesisState.StateInfoList[0].StartHeight = 3
	genesisState.StateInfoList[0].BDs = bd
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
}

func TestStateInfoWithWrongFinalizationInitGenesis(t *testing.T) {
	bd := types.BlockDescriptors{
		BD: []types.BlockDescriptor{{
			Height:                 3,
			StateRoot:              bytes.Repeat([]byte{byte(1)}, 32),
			IntermediateStatesRoot: bytes.Repeat([]byte{byte(1)}, 32),
		}},
	}
	genesisState.StateInfoList = append(genesisState.StateInfoList, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{
			RollappId: "0",
			Index:     3,
		},
		Sequencer:   sample.AccAddress(),
		NumBlocks:   1,
		StartHeight: 3,
		BDs:         bd,
		Status:      common.Status_FINALIZED,
	})
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
}

/*func TestMissingLatestStateInfoInitGenesis(t *testing.T) {
	genesisState.LatestStateInfoIndexList[0].Index = 3
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.StateInfoList, got.StateInfoList)
	require.NotEqual(t, genesisState.LatestStateInfoIndexList, got.LatestStateInfoIndexList)
	require.NotEqual(t, genesisState.LatestFinalizedStateIndexList, got.LatestFinalizedStateIndexList)
	require.NotEqual(t, genesisState.BlockHeightToFinalizationQueueList, got.BlockHeightToFinalizationQueueList)
}

func TestWrongLatestStateInfoInitGenesis(t *testing.T) {
	genesisState.LatestStateInfoIndexList[0].Index = 1
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.LatestStateInfoIndexList, got.LatestStateInfoIndexList)
}

func TestMissingFinalizedStateInfoInitGenesis(t *testing.T) {
	genesisState.StateInfoList[0] = types.StateInfo{}
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.LatestFinalizedStateIndexList, got.LatestFinalizedStateIndexList)
}

func TestMissingLatestFinalizedStateInfoInitGenesis(t *testing.T) {
	genesisState.LatestStateInfoIndexList[0] = types.StateInfoIndex{}
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.LatestFinalizedStateIndexList, got.LatestFinalizedStateIndexList)
}

func TestWrongFinalizedIndexInitGenesis(t *testing.T) {
	genesisState.LatestFinalizedStateIndexList[0] = types.StateInfoIndex{
		RollappId: "0",
		Index:     3,
	}

	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.LatestFinalizedStateIndexList, got.LatestFinalizedStateIndexList)
}

func TestWrongFinalizedHeightIndexInitGenesis(t *testing.T) {
	genesisState.LatestFinalizedStateIndexList[0] = types.StateInfoIndex{
		RollappId: "0",
		Index:     4,
	}
	blockDescriptors3 := types.BlockDescriptors{BD: []types.BlockDescriptor{
		{
			Height:                 4,
			StateRoot:              bytes.Repeat([]byte{byte(1)}, 32),
			IntermediateStatesRoot: bytes.Repeat([]byte{byte(1)}, 32),
		},
	}}
	genesisState.StateInfoList = append(genesisState.StateInfoList, types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{
			RollappId: "0",
			Index:     4,
		},
		Sequencer:   sample.AccAddress(),
		NumBlocks:   1,
		StartHeight: 4,
		BDs:         blockDescriptors3,
		Status:      common.Status_PENDING,
	})
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.LatestFinalizedStateIndexList, got.LatestFinalizedStateIndexList)
}

func TestWrongFinalizationQueueInitGenesis(t *testing.T) {
	genesisState.BlockHeightToFinalizationQueueList[0] = types.BlockHeightToFinalizationQueue{
		CreationHeight: 1,
		FinalizationQueue: []types.StateInfoIndex{
			{
				RollappId: "0",
				Index:     1,
			},
		},
	}
	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.NotEqual(t, genesisState.BlockHeightToFinalizationQueueList, got.BlockHeightToFinalizationQueueList)
}*/
