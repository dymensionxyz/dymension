package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	lightClientKeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

type testInput struct {
	rollappId string
	stateInfo *rollapptypes.StateInfo
}

func TestAfterUpdateState(t *testing.T) {
	testCases := []struct {
		name      string
		prepare   func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput
		expectErr bool
	}{
		{
			name: "canonical client does not exist for rollapp",
			prepare: func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput {
				return testInput{
					rollappId: "rollapp-no-canon-client",
					stateInfo: &rollapptypes.StateInfo{},
				}
			},
			expectErr: false,
		},
		{
			name: "canonical client exists but consensus state is not found for given height",
			prepare: func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client-but-no-state", "canon-client-id-no-state")
				return testInput{
					rollappId: "rollapp-has-canon-client-but-no-state",
					stateInfo: &rollapptypes.StateInfo{
						Sequencer:   keepertest.Alice,
						StartHeight: 1,
						NumBlocks:   1,
						BDs: rollapptypes.BlockDescriptors{
							BD: []rollapptypes.BlockDescriptor{
								{
									Height:    1,
									StateRoot: []byte("test"),
									Timestamp: time.Unix(1724392989, 0),
								},
							},
						},
					},
				}
			},
			expectErr: false,
		},
		{
			name: "both states are not compatible - slash the sequencer who signed",
			prepare: func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
				blockSignerTmPubKey, err := k.GetSequencerPubKey(ctx, keepertest.Alice)
				require.NoError(t, err)
				blockSignerTmPubKeyBytes, err := blockSignerTmPubKey.Marshal()
				require.NoError(t, err)
				k.SetConsensusStateSigner(ctx, "canon-client-id", 2, blockSignerTmPubKeyBytes)
				return testInput{
					rollappId: "rollapp-has-canon-client",
					stateInfo: &rollapptypes.StateInfo{
						Sequencer:   keepertest.Alice,
						StartHeight: 1,
						NumBlocks:   3,
						BDs: rollapptypes.BlockDescriptors{
							BD: []rollapptypes.BlockDescriptor{
								{
									Height:    1,
									StateRoot: []byte("test"),
									Timestamp: time.Unix(1724392989, 0),
								},
								{
									Height:    2,
									StateRoot: []byte("this is not compatible"),
									Timestamp: time.Unix(1724392989, 0).Add(1),
								},
								{
									Height:    3,
									StateRoot: []byte("test3"),
									Timestamp: time.Unix(1724392989, 0).Add(2),
								},
							},
						},
					},
				}
			},
			expectErr: false,
		},
		{
			name: "state is compatible",
			prepare: func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
				blockSignerTmPubKey, err := k.GetSequencerPubKey(ctx, keepertest.Alice)
				require.NoError(t, err)
				blockSignerTmPubKeyBytes, err := blockSignerTmPubKey.Marshal()
				require.NoError(t, err)
				k.SetConsensusStateSigner(ctx, "canon-client-id", 2, blockSignerTmPubKeyBytes)
				return testInput{
					rollappId: "rollapp-has-canon-client",
					stateInfo: &rollapptypes.StateInfo{
						Sequencer:   keepertest.Alice,
						StartHeight: 1,
						NumBlocks:   3,
						BDs: rollapptypes.BlockDescriptors{
							BD: []rollapptypes.BlockDescriptor{
								{
									Height:    1,
									StateRoot: []byte("test"),
									Timestamp: time.Unix(1724392989, 0),
								},
								{
									Height:    2,
									StateRoot: []byte("test2"),
									Timestamp: time.Unix(1724392989, 0),
								},
								{
									Height:    3,
									StateRoot: []byte("test3"),
									Timestamp: time.Unix(1724392989, 0).Add(1),
								},
							},
						},
					},
				}
			},
			expectErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.LightClientKeeper(t)
			input := tc.prepare(ctx, *keeper)
			err := keeper.RollappHooks().AfterUpdateState(ctx, input.rollappId, input.stateInfo)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAfterUpdateState_SetCanonicalClient(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	rollappId := "rollapp-wants-canon-client"
	stateInfo := &rollapptypes.StateInfo{
		Sequencer:   keepertest.Alice,
		StartHeight: 1,
		NumBlocks:   3,
		BDs: rollapptypes.BlockDescriptors{
			BD: []rollapptypes.BlockDescriptor{
				{
					Height:    1,
					StateRoot: []byte("test"),
					Timestamp: time.Unix(1724392989, 0),
				},
				{
					Height:    2,
					StateRoot: []byte("test2"),
					Timestamp: time.Unix(1724392989, 0),
				},
				{
					Height:    3,
					StateRoot: []byte("test3"),
					Timestamp: time.Unix(1724392989, 0).UTC(),
				},
			},
		},
	}
	err := keeper.RollappHooks().AfterUpdateState(ctx, rollappId, stateInfo)
	require.NoError(t, err)

	clientID, found := keeper.GetCanonicalClient(ctx, rollappId)
	require.True(t, found)
	require.Equal(t, "canon-client-id", clientID)
}
