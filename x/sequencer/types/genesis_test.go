package types_test

import (
	"testing"

	_ "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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
			desc: "valid genesis sequencers",
			genState: &types.GenesisState{
				Params:                  types.DefaultParams(),
				SequencerList:           []types.Sequencer{{SequencerAddress: "0"}, {SequencerAddress: "1"}},
				SequencersByRollappList: []types.SequencersByRollapp{{RollappId: "0"}, {RollappId: "1"}},
			},
			valid: true,
		},
		{
			desc: "duplicated sequencer",
			genState: &types.GenesisState{
				Params:                  types.DefaultParams(),
				SequencerList:           []types.Sequencer{{SequencerAddress: "0"}, {SequencerAddress: "0"}},
				SequencersByRollappList: []types.SequencersByRollapp{},
			},
			valid: false,
		},
		{
			desc: "duplicated sequencersByRollapp",
			genState: &types.GenesisState{
				Params:                  types.DefaultParams(),
				SequencerList:           []types.Sequencer{},
				SequencersByRollappList: []types.SequencersByRollapp{{RollappId: "0"}, {RollappId: "0"}},
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
