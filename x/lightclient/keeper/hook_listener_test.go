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
		// TODO: tests need expanding
		//  At least the following need to be added
		//  - Client with all cons states after the state update will not be canonical
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
			name: "both states are not compatible - slash the sequencer who signed",
			prepare: func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
				err := k.SaveSigner(ctx, keepertest.Alice.Address, keepertest.CanonClientID, 2)
				require.NoError(t, err)
				return testInput{
					rollappId: keepertest.DefaultRollapp,
					stateInfo: &rollapptypes.StateInfo{
						Sequencer:   keepertest.Alice.Address,
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
			expectErr: true,
		},
		{
			name: "state is compatible",
			prepare: func(ctx sdk.Context, k lightClientKeeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
				err := k.SaveSigner(ctx, keepertest.Alice.Address, keepertest.CanonClientID, 2)
				require.NoError(t, err)

				return testInput{
					rollappId: keepertest.DefaultRollapp,
					stateInfo: &rollapptypes.StateInfo{
						Sequencer:   keepertest.Alice.Address,
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
