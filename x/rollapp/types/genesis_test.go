package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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
			desc: "valid genesis state",
			genState: &types.GenesisState{
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
				ObsoleteDrsVersions: []uint32{1, 2},
			},
			valid: true,
		},
		{
			desc: "duplicated rollapp",
			genState: &types.GenesisState{
				Params:                             types.DefaultParams(),
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
				Params:                             types.DefaultParams().WithDisputePeriodInBlocks(types.MinDisputePeriodInBlocks - 1),
				RollappList:                        []types.Rollapp{{RollappId: "0"}},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "invalid LivenessSlashBlocks",
			genState: &types.GenesisState{
				Params:                             types.DefaultParams().WithLivenessSlashBlocks(0),
				RollappList:                        []types.Rollapp{{RollappId: "0"}},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
			},
			valid: false,
		},
		{
			desc: "invalid LivenessSlashInterval",
			genState: &types.GenesisState{
				Params:                             types.DefaultParams().WithLivenessSlashInterval(0),
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
				Params:                             types.DefaultParams(),
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
				Params:                             types.DefaultParams(),
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
				Params:                             types.DefaultParams(),
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{{CreationHeight: 0}, {CreationHeight: 0}},
			},
			valid: false,
		},
		{
			desc: "duplicated ObsoleteDrsVersions",
			genState: &types.GenesisState{
				Params:                             types.DefaultParams(),
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
				ObsoleteDrsVersions:                []uint32{1, 1},
			},
			valid: false,
		},
		{
			desc: "duplicated livenessEvents",
			genState: &types.GenesisState{
				Params:                             types.DefaultParams(),
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
				ObsoleteDrsVersions:                []uint32{},
				LivenessEvents: []types.LivenessEvent{
					{RollappId: "rollapp1"},
					{RollappId: "rollapp1"},
				},
			},
			valid: false,
		},
		{
			desc: "empty RollappId in RegisteredDenoms",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				RegisteredDenoms: []types.RollappRegisteredDenoms{
					{RollappId: "", Denoms: []string{"denom1"}},
				},
			},
			valid: false,
		},
		{
			desc: "duplicate RollappId in RegisteredDenoms",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				RegisteredDenoms: []types.RollappRegisteredDenoms{
					{RollappId: "rollapp1", Denoms: []string{"denom1"}},
					{RollappId: "rollapp1", Denoms: []string{"denom2"}},
				},
			},
			valid: false,
		},
		{
			desc: "empty Denoms list in RegisteredDenoms",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				RegisteredDenoms: []types.RollappRegisteredDenoms{
					{RollappId: "rollapp1", Denoms: []string{}},
				},
			},
			valid: false,
		},
		{
			desc: "empty Denom value in RegisteredDenoms",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				RegisteredDenoms: []types.RollappRegisteredDenoms{
					{RollappId: "rollapp1", Denoms: []string{"", "denom2"}},
				},
			},
			valid: false,
		},
		{
			desc: "duplicate Denoms in RegisteredDenoms for the same RollappId",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				RegisteredDenoms: []types.RollappRegisteredDenoms{
					{RollappId: "rollapp1", Denoms: []string{"denom1", "denom1"}},
				},
			},
			valid: false,
		},
		{
			desc: "valid RegisteredDenoms entry",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				RegisteredDenoms: []types.RollappRegisteredDenoms{
					{RollappId: "rollapp1", Denoms: []string{"denom1", "denom2"}},
					{RollappId: "rollapp2", Denoms: []string{"denom3"}},
				},
			},
			valid: true,
		},
		{
			desc: "empty Sequencer field",
			genState: &types.GenesisState{
				SequencerHeightPairs: []types.SequencerHeightPair{
					{Sequencer: "", Height: 10},
				},
			},
			valid: false,
		},
		{
			desc: "zero Height field",
			genState: &types.GenesisState{
				SequencerHeightPairs: []types.SequencerHeightPair{
					{Sequencer: "sequencer1", Height: 0},
				},
			},
			valid: false,
		},
		{
			desc: "duplicate Sequencer-Height pair",
			genState: &types.GenesisState{
				SequencerHeightPairs: []types.SequencerHeightPair{
					{Sequencer: "sequencer1", Height: 10},
					{Sequencer: "sequencer1", Height: 10},
				},
			},
			valid: false,
		},
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
