package types_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
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
				Params: types.Params{DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks},
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
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated rollapp",
			genState: &types.GenesisState{
				Params: types.Params{DisputePeriodInBlocks: types.DefaultGenesis().Params.DisputePeriodInBlocks},
				RollappList: []types.Rollapp{
					{
						RollappId: "0",
					},
					{
						RollappId: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid DisputePeriodInBlocks",
			genState: &types.GenesisState{
				Params: types.Params{DisputePeriodInBlocks: types.MinDisputePeriodInBlocks - 1},
				RollappList: []types.Rollapp{
					{
						RollappId: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated stateInfo",
			genState: &types.GenesisState{
				StateInfoList: []types.StateInfo{
					{
						StateInfoIndex: types.StateInfoIndex{
							RollappId: "0",
							Index:     0,
						},
					},
					{
						StateInfoIndex: types.StateInfoIndex{
							RollappId: "0",
							Index:     0,
						},
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated latestStateInfoIndex",
			genState: &types.GenesisState{
				LatestStateInfoIndexList: []types.StateInfoIndex{
					{
						RollappId: "0",
					},
					{
						RollappId: "0",
					},
				},
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
