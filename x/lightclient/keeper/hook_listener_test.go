package keeper_test

import (
	"testing"
	"time"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

type testInput struct {
	rollappId                 string
	stateInfo                 *rollapptypes.StateInfo
	isFirstStateUpdate        bool
	previousStateHasTimestamp bool
}

func TestAfterUpdateState(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	keeper.SetCanonicalClient(ctx, "rollapp-has-canon-client-but-no-state", "canon-client-id-no-state")
	keeper.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")

	testCases := []struct {
		name  string
		input testInput
	}{
		{
			name: "canonical client does not exist for rollapp",
			input: testInput{
				rollappId:          "rollapp-no-canon-client",
				stateInfo:          &rollapptypes.StateInfo{},
				isFirstStateUpdate: true,
			},
		},
		{
			name: "canonical client exists but the BDs are empty",
			input: testInput{
				rollappId:          "rollapp-has-canon-client",
				stateInfo:          &rollapptypes.StateInfo{},
				isFirstStateUpdate: true,
			},
		},
		{
			name: "canonical client exists but consensus state is not found",
			input: testInput{
				rollappId: "rollapp-has-canon-client-but-no-state",
				stateInfo: &rollapptypes.StateInfo{
					BDs: rollapptypes.BlockDescriptors{
						BD: []rollapptypes.BlockDescriptor{
							{
								Height:    1,
								StateRoot: []byte("test"),
								Timestamp: time.Now().UTC(),
							},
						},
					},
				},
				isFirstStateUpdate: true,
			},
		},
		{
			name: "BD does not include next block in state info",
			input: testInput{
				rollappId: "rollapp-has-canon-client",
				stateInfo: &rollapptypes.StateInfo{
					BDs: rollapptypes.BlockDescriptors{
						BD: []rollapptypes.BlockDescriptor{
							{
								Height:    1,
								StateRoot: []byte("test"),
								Timestamp: time.Now().UTC(),
							},
							{
								Height:    2,
								StateRoot: []byte("test2"),
								Timestamp: time.Now().Add(1).UTC(),
							},
						},
					},
				},
				isFirstStateUpdate: true,
			},
		},
		{
			name: "both states are not compatible - slash the sequencer who signed",
			input: testInput{
				rollappId: "rollapp-has-canon-client",
				stateInfo: &rollapptypes.StateInfo{
					Sequencer: keepertest.Alice,
					BDs: rollapptypes.BlockDescriptors{
						BD: []rollapptypes.BlockDescriptor{
							{
								Height:    1,
								StateRoot: []byte("test"),
								Timestamp: time.Now().UTC(),
							},
							{
								Height:    2,
								StateRoot: []byte("this is not compatible"),
								Timestamp: time.Now().Add(1).UTC(),
							},
							{
								Height:    3,
								StateRoot: []byte("test3"),
								Timestamp: time.Now().Add(2).UTC(),
							},
						},
					},
				},
				isFirstStateUpdate: true,
			},
		},
		{
			name: "timestamp is missing and its first state update",
			input: testInput{
				rollappId: "rollapp-has-canon-client",
				stateInfo: &rollapptypes.StateInfo{
					Sequencer: keepertest.Alice,
					BDs: rollapptypes.BlockDescriptors{
						BD: []rollapptypes.BlockDescriptor{
							{
								Height:    1,
								StateRoot: []byte("test"),
								Timestamp: time.Now().UTC(),
							},
							{
								Height:    2,
								StateRoot: []byte("test2"),
								Timestamp: time.Now().Add(1).UTC(),
							},
							{
								Height:    3,
								StateRoot: []byte("test3"),
								Timestamp: time.Now().Add(2).UTC(),
							},
						},
					},
				},
				isFirstStateUpdate: true,
			},
		},
		{
			name: "state is compatible",
			input: testInput{
				rollappId: "rollapp-has-canon-client",
				stateInfo: &rollapptypes.StateInfo{
					Sequencer: keepertest.Alice,
					BDs: rollapptypes.BlockDescriptors{
						BD: []rollapptypes.BlockDescriptor{
							{
								Height:    1,
								StateRoot: []byte("test"),
								Timestamp: time.Now().UTC(),
							},
							{
								Height:    2,
								StateRoot: []byte("test2"),
								Timestamp: time.Now().Add(1).UTC(),
							},
							{
								Height:    3,
								StateRoot: []byte("test3"),
								Timestamp: time.Now().Add(2).UTC(),
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := keeper.RollappHooks().AfterUpdateState(ctx, tc.input.rollappId, tc.input.stateInfo, tc.input.isFirstStateUpdate, tc.input.previousStateHasTimestamp)
			require.NoError(t, err)
		})
	}
}
