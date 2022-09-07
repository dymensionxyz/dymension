package types_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

var (
	seqAddr1 = sample.AccAddress()
	seqAddr2 = sample.AccAddress()
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state with empty DeployerWhitelist",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks,
					DeployerWhitelist:     []string{},
				},
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
						FinalizationHeight: 0,
					},
					{
						FinalizationHeight: 1,
					},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "valid genesis state with DeployerWhitelist",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks,
					DeployerWhitelist:     []string{seqAddr1, seqAddr2},
				},
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
						FinalizationHeight: 0,
					},
					{
						FinalizationHeight: 1,
					},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated rollapp",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks,
					DeployerWhitelist:     []string{},
				},
				RollappList:                        []types.Rollapp{{RollappId: "0"}, {RollappId: "0"}},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "invalid DisputePeriodInBlocks",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.MinDisputePeriodInBlocks - 1,
					DeployerWhitelist:     []string{},
				},
				RollappList:                        []types.Rollapp{{RollappId: "0"}},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "invalid DeployerWhitelist",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.MinDisputePeriodInBlocks,
					DeployerWhitelist:     []string{"asdad"},
				},
				RollappList:                        []types.Rollapp{{RollappId: "0"}},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "duplicated stateInfo",
			genState: &types.GenesisState{
				Params:                             types.Params{},
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{{StateInfoIndex: types.StateInfoIndex{RollappId: "0", Index: 0}}, {StateInfoIndex: types.StateInfoIndex{RollappId: "0", Index: 0}}},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "duplicated latestStateInfoIndex",
			genState: &types.GenesisState{
				Params:                             types.Params{},
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{{RollappId: "0"}, {RollappId: "0"}},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "duplicated blockHeightToFinalizationQueue",
			genState: &types.GenesisState{
				Params:                             types.Params{},
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{{FinalizationHeight: 0}, {FinalizationHeight: 0}},
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
