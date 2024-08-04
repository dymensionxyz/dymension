package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
				Params: types.Params{
					DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks,
					RegistrationFee:       sdk.NewCoin("adym", sdk.NewInt(1000)),
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
						CreationHeight: 0,
					},
					{
						CreationHeight: 1,
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicated rollapp",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks,
					RegistrationFee:       sdk.NewCoin("adym", sdk.NewInt(1000)),
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
					RegistrationFee:       sdk.NewCoin("adym", sdk.NewInt(1000)),
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
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{{CreationHeight: 0}, {CreationHeight: 0}},
			},
			valid: false,
		},
		{
			desc: "invalid registration fee",
			genState: &types.GenesisState{
				Params: types.Params{
					DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks,
					RegistrationFee:       sdk.NewCoin("cosmos", sdk.NewInt(0)),
				},
				RollappList:                        []types.Rollapp{},
				StateInfoList:                      []types.StateInfo{},
				LatestStateInfoIndexList:           []types.StateInfoIndex{},
				BlockHeightToFinalizationQueueList: []types.BlockHeightToFinalizationQueue{},
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
