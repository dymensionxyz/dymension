package ante_test

import (
	"testing"
	"time"

	"github.com/cometbft/cometbft/libs/math"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	ibcsolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var testClientState = ibctm.NewClientState("chain-id",
	ibctm.DefaultTrustLevel, time.Hour*24*7*2, time.Hour*24*7*2, time.Second*10,
	ibcclienttypes.MustParseHeight("1-1"), commitmenttypes.GetSDKSpecs(), []string{},
)

func TestHandleMsgCreateClient(t *testing.T) {
	type testInput struct {
		msg        *ibcclienttypes.MsgCreateClient
		rollapps   map[string]rollapptypes.Rollapp
		stateInfos map[string]map[uint64]rollapptypes.StateInfo
	}

	testCases := []struct {
		name    string
		prepare func(ctx sdk.Context, k keeper.Keeper) testInput
		assert  func(ctx sdk.Context, k keeper.Keeper)
	}{
		{
			name: "Could not unpack light client state to tendermint state",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {},
		},
		{
			name: "Client is not a tendermint client",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				solomachineClientState := ibcsolomachine.NewClientState(0, nil)
				cs, err := ibcclienttypes.PackClientState(solomachineClientState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {},
		},
		{
			name: "Rollapp with given chainID does not exist",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "not-a-rollapp"
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				_, found := k.GetCanonicalLightClientRegistration(ctx, "not-a-rollapp")
				require.False(t, found)
			},
		},
		{
			name: "Canonical client for the rollapp already exists",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "rollapp-has-canon-client"
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-has-canon-client": {
							RollappId: "rollapp-has-canon-client",
							ChannelId: "channel-on-canon-client",
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				clientID, found := k.GetCanonicalClient(ctx, "rollapp-has-canon-client")
				require.True(t, found)
				require.Equal(t, "canon-client-id", clientID)
				_, registrationPending := k.GetCanonicalLightClientRegistration(ctx, "rollapp-has-canon-client")
				require.False(t, registrationPending)
			},
		},
		{
			name: "Could not find block desc for given height - continue without setting as canonical client",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "rollapp-wants-canon-client"
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-wants-canon-client": {
							RollappId: "rollapp-wants-canon-client",
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				_, found := k.GetCanonicalLightClientRegistration(ctx, "rollapp-wants-canon-client")
				require.False(t, found)
			},
		},
		{
			name: "Could not find block descriptor for the next height (h+1) - continue without setting as canonical client",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "rollapp-wants-canon-client"
				cs, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState: cs,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-wants-canon-client": {
							RollappId: "rollapp-wants-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-wants-canon-client": {
							1: {
								StartHeight: 1,
								NumBlocks:   1,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 1,
								},
								Sequencer: keepertest.Alice,
								BDs: rollapptypes.BlockDescriptors{
									BD: []rollapptypes.BlockDescriptor{
										{
											Height:    1,
											StateRoot: []byte{},
											Timestamp: time.Unix(1724392989, 0),
										},
									},
								},
							},
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				_, found := k.GetCanonicalLightClientRegistration(ctx, "rollapp-wants-canon-client")
				require.False(t, found)
			},
		},
		{
			name: "State incompatible - continue without setting as canonical client",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				testClientState.ChainId = "rollapp-wants-canon-client"
				clientState, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				testConsensusState := ibctm.NewConsensusState(
					time.Unix(1724392989, 0),
					commitmenttypes.NewMerkleRoot([]byte{}),
					ctx.BlockHeader().ValidatorsHash,
				)
				consState, err := ibcclienttypes.PackConsensusState(testConsensusState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState:    clientState,
						ConsensusState: consState,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-wants-canon-client": {
							RollappId: "rollapp-wants-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-wants-canon-client": {
							1: {
								StartHeight: 1,
								NumBlocks:   2,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 1,
								},
								Sequencer: keepertest.Alice,
								BDs: rollapptypes.BlockDescriptors{
									BD: []rollapptypes.BlockDescriptor{
										{
											Height:    1,
											StateRoot: []byte{},
											Timestamp: time.Unix(1724392989, 0),
										},
										{
											Height:    2,
											StateRoot: []byte{},
											Timestamp: time.Unix(1724392989, 0),
										},
									},
								},
							},
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				_, found := k.GetCanonicalLightClientRegistration(ctx, "rollapp-wants-canon-client")
				require.False(t, found)
			},
		},
		{
			name: "State compatible but client params not conforming to expected params - continue without setting as canonical client",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				blocktimestamp := time.Unix(1724392989, 0)
				testClientState.ChainId = "rollapp-wants-canon-client"
				clientState, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				var nextVals cmttypes.ValidatorSet
				tmPk, err := k.GetTmPubkey(ctx, keepertest.Alice)
				require.NoError(t, err)
				updates, err := cmttypes.PB2TM.ValidatorUpdates([]abci.ValidatorUpdate{{Power: 1, PubKey: tmPk}})
				require.NoError(t, err)
				err = nextVals.UpdateWithChangeSet(updates)
				require.NoError(t, err)
				testConsensusState := ibctm.NewConsensusState(
					blocktimestamp,
					commitmenttypes.NewMerkleRoot([]byte("appHash")),
					nextVals.Hash(),
				)
				consState, err := ibcclienttypes.PackConsensusState(testConsensusState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState:    clientState,
						ConsensusState: consState,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-wants-canon-client": {
							RollappId: "rollapp-wants-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-wants-canon-client": {
							1: {
								StartHeight: 1,
								NumBlocks:   2,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 1,
								},
								Sequencer: keepertest.Alice,
								BDs: rollapptypes.BlockDescriptors{
									BD: []rollapptypes.BlockDescriptor{
										{
											Height:    1,
											StateRoot: []byte("appHash"),
											Timestamp: blocktimestamp,
										},
										{
											Height:    2,
											StateRoot: []byte("appHash2"),
											Timestamp: blocktimestamp.Add(1),
										},
									},
								},
							},
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				_, found := k.GetCanonicalLightClientRegistration(ctx, "rollapp-wants-canon-client")
				require.False(t, found)
			},
		},
		{
			name: "State compatible + expected client params - Candidate canonical client set",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				blocktimestamp := time.Unix(1724392989, 0)
				testClientState.ChainId = "rollapp-wants-canon-client"
				testClientState.TrustLevel = ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 1})
				testClientState.TrustingPeriod = time.Hour * 24 * 7 * 2
				testClientState.UnbondingPeriod = time.Hour * 24 * 7 * 3
				testClientState.MaxClockDrift = time.Minute * 10
				testClientState.AllowUpdateAfterExpiry = false
				testClientState.AllowUpdateAfterMisbehaviour = false
				clientState, err := ibcclienttypes.PackClientState(testClientState)
				require.NoError(t, err)
				var nextVals cmttypes.ValidatorSet
				tmPk, err := k.GetTmPubkey(ctx, keepertest.Alice)
				require.NoError(t, err)
				updates, err := cmttypes.PB2TM.ValidatorUpdates([]abci.ValidatorUpdate{{Power: 1, PubKey: tmPk}})
				require.NoError(t, err)
				err = nextVals.UpdateWithChangeSet(updates)
				require.NoError(t, err)
				testConsensusState := ibctm.NewConsensusState(
					blocktimestamp,
					commitmenttypes.NewMerkleRoot([]byte("appHash")),
					nextVals.Hash(),
				)
				consState, err := ibcclienttypes.PackConsensusState(testConsensusState)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgCreateClient{
						ClientState:    clientState,
						ConsensusState: consState,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-wants-canon-client": {
							RollappId: "rollapp-wants-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-wants-canon-client": {
							1: {
								StartHeight: 1,
								NumBlocks:   2,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 1,
								},
								Sequencer: keepertest.Alice,
								BDs: rollapptypes.BlockDescriptors{
									BD: []rollapptypes.BlockDescriptor{
										{
											Height:    1,
											StateRoot: []byte("appHash"),
											Timestamp: blocktimestamp,
										},
										{
											Height:    2,
											StateRoot: []byte("appHash2"),
											Timestamp: blocktimestamp.Add(1),
										},
									},
								},
							},
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper) {
				clientID, found := k.GetCanonicalLightClientRegistration(ctx, "rollapp-wants-canon-client")
				require.True(t, found)
				require.Equal(t, "new-canon-client-1", clientID)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.LightClientKeeper(t)
			ibcclientKeeper := NewMockIBCClientKeeper(nil)
			ibcchannelKeeper := NewMockIBCChannelKeeper(nil)
			input := tc.prepare(ctx, *keeper)
			rollappKeeper := NewMockRollappKeeper(input.rollapps, input.stateInfos)
			ibcMsgDecorator := ante.NewIBCMessagesDecorator(*keeper, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)

			ibcMsgDecorator.HandleMsgCreateClient(ctx, input.msg)
			tc.assert(ctx, *keeper)
		})
	}
}
