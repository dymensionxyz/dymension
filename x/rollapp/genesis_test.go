package rollapp_test

import (
	"strings"
	"testing"

	"golang.org/x/exp/slices"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	"github.com/dymensionxyz/dymension/v3/x/rollapp"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestInitExportGenesis(t *testing.T) {
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
		GenesisTransfers: []types.GenesisTransfers{
			{
				RollappID:   "0",
				NumTotal:    3,
				NumReceived: 3,
				Received:    []uint64{0, 2, 1},
			},
			{
				RollappID:   "1",
				NumTotal:    3,
				NumReceived: 3,
				Received:    []uint64{0, 2, 1},
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(genesisState)
	nullify.Fill(*got)

	require.True(t, GenesisTransfersAreEquivalent(genesisState.GetGenesisTransfers(), got.GetGenesisTransfers()))
	require.ElementsMatch(t, genesisState.GenesisTransfers, got.GenesisTransfers)
	require.ElementsMatch(t, genesisState.RollappList, got.RollappList)
	require.ElementsMatch(t, genesisState.StateInfoList, got.StateInfoList)
	require.ElementsMatch(t, genesisState.LatestStateInfoIndexList, got.LatestStateInfoIndexList)
	require.ElementsMatch(t, genesisState.BlockHeightToFinalizationQueueList, got.BlockHeightToFinalizationQueueList)
	// this line is used by starport scaffolding # genesis/test/assert
}

// GenesisTransfersAreEquivalent returns if a,b are the same, in terms of containing
// the same semantic content. Intended for use in tests.
func GenesisTransfersAreEquivalent(x, y []types.GenesisTransfers) bool {
	if len(x) != len(y) {
		return false
	}
	sortOne := func(l []types.GenesisTransfers) {
		slices.SortStableFunc(l, func(a, b types.GenesisTransfers) bool {
			return strings.Compare(a.GetRollappID(), b.GetRollappID()) <= 0
		})
	}
	sortTwo := func(l []types.GenesisTransfers) {
		for _, transfer := range l {
			slices.SortStableFunc(transfer.Received, func(a, b uint64) bool {
				return a <= b
			})
		}
	}
	sortOne(x)
	sortOne(y)
	sortTwo(x)
	sortTwo(y)
	for i := range len(x) {
		a := x[i]
		b := y[i]
		if a.NumTotal != b.NumTotal {
			return false
		}
		if a.NumReceived != b.NumReceived {
			return false
		}
		if a.RollappID != b.RollappID {
			return false
		}
		if len(a.Received) != len(b.Received) {
			return false
		}
		for j := range len(a.Received) {
			if a.Received[j] != b.Received[j] {
				return false
			}
		}
	}

	return true
}
