package types_test

import (
	"testing"
	"time"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestCheckCompatibility(t *testing.T) {
	type input struct {
		ibcState types.IBCState
		raState  types.RollappState
	}
	timestamp := time.Unix(1724392989, 0)
	testCases := []struct {
		name  string
		input input
		err   string
	}{
		{
			name: "roots are not equal",
			input: input{
				ibcState: types.IBCState{
					Root:               []byte("root"),
					Timestamp:          timestamp,
					NextValidatorsHash: []byte{156, 132, 96, 43, 190, 214, 140, 148, 216, 119, 98, 162, 97, 120, 115, 32, 39, 223, 114, 56, 224, 180, 80, 228, 190, 243, 9, 248, 190, 33, 188, 23},
					Validator:          []byte("sequencer"),
				},
				raState: types.RollappState{
					BlockSequencer: []byte("sequencer"),
					BlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("not same root"),
						Timestamp: timestamp,
					},
					NextBlockSequencer: []byte{10, 32, 86, 211, 180, 178, 104, 144, 159, 216, 7, 137, 173, 225, 55, 215, 228, 176, 29, 86, 98, 130, 25, 190, 214, 24, 198, 22, 111, 37, 100, 142, 154, 87},
					NextBlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root2"),
						Timestamp: timestamp,
					},
				},
			},
			err: "block descriptor state root does not match tendermint header app hash",
		},
		{
			name: "validator who signed the block header is not the sequencer who submitted the block",
			input: input{
				ibcState: types.IBCState{
					Root:               []byte("root"),
					Timestamp:          timestamp,
					NextValidatorsHash: []byte{156, 132, 96, 43, 190, 214, 140, 148, 216, 119, 98, 162, 97, 120, 115, 32, 39, 223, 114, 56, 224, 180, 80, 228, 190, 243, 9, 248, 190, 33, 188, 23},
					Validator:          []byte("validator"),
				},
				raState: types.RollappState{
					BlockSequencer: []byte("sequencer"),
					BlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root"),
						Timestamp: timestamp,
					},
					NextBlockSequencer: []byte{10, 32, 86, 211, 180, 178, 104, 144, 159, 216, 7, 137, 173, 225, 55, 215, 228, 176, 29, 86, 98, 130, 25, 190, 214, 24, 198, 22, 111, 37, 100, 142, 154, 87},
					NextBlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root2"),
						Timestamp: timestamp,
					},
				},
			},
			err: "validator does not match the sequencer",
		},
		{
			name: "nextValidatorHash does not match the sequencer who submitted the next block descriptor",
			input: input{
				ibcState: types.IBCState{
					Root:               []byte("root"),
					Timestamp:          timestamp,
					NextValidatorsHash: []byte("next validator hash"),
					Validator:          []byte("sequencer"),
				},
				raState: types.RollappState{
					BlockSequencer: []byte("sequencer"),
					BlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root"),
						Timestamp: timestamp,
					},
					NextBlockSequencer: []byte{10, 32, 6, 234, 170, 31, 60, 164, 237, 129, 237, 38, 152, 233, 81, 240, 243, 121, 79, 108, 152, 75, 27, 247, 76, 48, 15, 132, 61, 27, 161, 82, 197, 249},
					NextBlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root2"),
						Timestamp: timestamp,
					},
				},
			},
			err: "next validator hash does not match the sequencer for h+1",
		},
		{
			name: "timestamps is empty",
			input: input{
				ibcState: types.IBCState{
					Root:               []byte("root"),
					Timestamp:          timestamp,
					NextValidatorsHash: []byte{156, 132, 96, 43, 190, 214, 140, 148, 216, 119, 98, 162, 97, 120, 115, 32, 39, 223, 114, 56, 224, 180, 80, 228, 190, 243, 9, 248, 190, 33, 188, 23},
					Validator:          []byte("sequencer"),
				},
				raState: types.RollappState{
					BlockSequencer: []byte("sequencer"),
					BlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root"),
					},
					NextBlockSequencer: []byte{10, 32, 86, 211, 180, 178, 104, 144, 159, 216, 7, 137, 173, 225, 55, 215, 228, 176, 29, 86, 98, 130, 25, 190, 214, 24, 198, 22, 111, 37, 100, 142, 154, 87},
					NextBlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root2"),
					},
				},
			},
			err: "block descriptors do not contain block timestamp",
		},
		{
			name: "timestamps are not equal",
			input: input{
				ibcState: types.IBCState{
					Root:               []byte("root"),
					Timestamp:          timestamp,
					NextValidatorsHash: []byte{156, 132, 96, 43, 190, 214, 140, 148, 216, 119, 98, 162, 97, 120, 115, 32, 39, 223, 114, 56, 224, 180, 80, 228, 190, 243, 9, 248, 190, 33, 188, 23},
					Validator:          []byte("sequencer"),
				},
				raState: types.RollappState{
					BlockSequencer: []byte("sequencer"),
					BlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root"),
						Timestamp: timestamp.Add(1),
					},
					NextBlockSequencer: []byte{10, 32, 86, 211, 180, 178, 104, 144, 159, 216, 7, 137, 173, 225, 55, 215, 228, 176, 29, 86, 98, 130, 25, 190, 214, 24, 198, 22, 111, 37, 100, 142, 154, 87},
					NextBlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root2"),
						Timestamp: timestamp,
					},
				},
			},
			err: "block descriptor timestamp does not match tendermint header timestamp",
		},
		{
			name: "all fields are compatible",
			input: input{
				ibcState: types.IBCState{
					Root:               []byte("root"),
					Timestamp:          timestamp,
					NextValidatorsHash: []byte{156, 132, 96, 43, 190, 214, 140, 148, 216, 119, 98, 162, 97, 120, 115, 32, 39, 223, 114, 56, 224, 180, 80, 228, 190, 243, 9, 248, 190, 33, 188, 23},
					Validator:          []byte("sequencer"),
				},
				raState: types.RollappState{
					BlockSequencer: []byte("sequencer"),
					BlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root"),
						Timestamp: timestamp,
					},
					NextBlockSequencer: []byte{10, 32, 86, 211, 180, 178, 104, 144, 159, 216, 7, 137, 173, 225, 55, 215, 228, 176, 29, 86, 98, 130, 25, 190, 214, 24, 198, 22, 111, 37, 100, 142, 154, 87},
					NextBlockDescriptor: rollapptypes.BlockDescriptor{
						StateRoot: []byte("root2"),
						Timestamp: timestamp,
					},
				},
			},
			err: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.CheckCompatibility(tc.input.ibcState, tc.input.raState)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.err)
			}
		})
	}
}
