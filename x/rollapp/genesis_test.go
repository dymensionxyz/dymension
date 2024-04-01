package rollapp_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name     string
		params   types.Params
		rollapps []types.Rollapp
		expPanic bool
	}{
		{
			name: "only params - success",
			params: types.Params{
				DisputePeriodInBlocks: 11,
				DeployerWhitelist: []types.DeployerParams{{
					Address: "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft",
				}},
				RollappsEnabled: false,
			},
			rollapps: []types.Rollapp{},
			expPanic: false,
		},
		{
			name: "params and rollapps - panic",
			params: types.Params{
				DisputePeriodInBlocks: 11,
				DeployerWhitelist: []types.DeployerParams{{
					Address: "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft",
				}},
				RollappsEnabled: false,
			},
			rollapps: []types.Rollapp{{RollappId: "0"}},
			expPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.GenesisState{Params: tt.params, RollappList: tt.rollapps}
			k, ctx := keepertest.RollappKeeper(t)
			if tt.expPanic {
				require.Panics(t, func() {
					rollapp.InitGenesis(ctx, *k, genesisState)
				})
			} else {
				rollapp.InitGenesis(ctx, *k, genesisState)
				params := k.GetParams(ctx)
				require.Equal(t, genesisState.Params, params)
			}
		})
	}
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
