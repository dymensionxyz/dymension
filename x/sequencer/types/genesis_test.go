package types_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/x/sequencer/types"
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

				SequencerList: []types.Sequencer{
					{
						SequencerAddress: "0",
					},
					{
						SequencerAddress: "1",
					},
				},
				SequencersByRollappList: []types.SequencersByRollapp{
					{
						RollappId: "0",
					},
					{
						RollappId: "1",
					},
				},
				SchedulerList: []types.Scheduler{
					{
						SequencerAddress: "0",
					},
					{
						SequencerAddress: "1",
					},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated sequencer",
			genState: &types.GenesisState{
				SequencerList: []types.Sequencer{
					{
						SequencerAddress: "0",
					},
					{
						SequencerAddress: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated sequencersByRollapp",
			genState: &types.GenesisState{
				SequencersByRollappList: []types.SequencersByRollapp{
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
			desc: "duplicated scheduler",
			genState: &types.GenesisState{
				SchedulerList: []types.Scheduler{
					{
						SequencerAddress: "0",
					},
					{
						SequencerAddress: "0",
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
