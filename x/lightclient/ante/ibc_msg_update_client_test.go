package ante_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibcsolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgUpdateClient(t *testing.T) {
	type testInput struct {
		msg        *ibcclienttypes.MsgUpdateClient
		rollapps   map[string]rollapptypes.Rollapp
		stateInfos map[string]map[uint64]rollapptypes.StateInfo
	}
	testCases := []struct {
		name    string
		prepare func(ctx sdk.Context, k keeper.Keeper) testInput
		assert  func(ctx sdk.Context, k keeper.Keeper, err error)
	}{
		{
			name: "Could not find a client with given client id",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId: "non-existent-client",
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Could not unpack as tendermint client state",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId: "non-tm-client-id",
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Client is not a known canonical client of a rollapp",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId: "canon-client-id",
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "Could not find state info for height - ensure optimistically accepted and signer stored in state",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
				var valSet, trustedVals *cmtproto.ValidatorSet
				signedHeader := &cmtproto.SignedHeader{
					Header: &cmtproto.Header{
						ProposerAddress: []byte("sequencerAddr"),
					},
					Commit: &cmtproto.Commit{},
				}
				header := ibctm.Header{
					SignedHeader:      signedHeader,
					ValidatorSet:      valSet,
					TrustedHeight:     ibcclienttypes.MustParseHeight("1-1"),
					TrustedValidators: trustedVals,
				}
				clientMsg, err := ibcclienttypes.PackClientMessage(&header)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId:      "canon-client-id",
						ClientMessage: clientMsg,
						Signer:        "relayerAddr",
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-has-canon-client": {
							RollappId: "rollapp-has-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-has-canon-client": {
							3: {
								Sequencer: keepertest.Alice,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 3,
								},
								StartHeight: 3,
								NumBlocks:   1,
								BDs: rollapptypes.BlockDescriptors{
									BD: []rollapptypes.BlockDescriptor{
										{
											Height:    3,
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
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.NoError(t, err)
				signer, found := k.GetConsensusStateSigner(ctx, "canon-client-id", 1)
				require.True(t, found)
				require.Equal(t, sdk.AccAddress([]byte("sequencerAddr")).String(), signer)
			},
		},
		{
			name: "State is incompatible - do not accept",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
				var (
					valSet      *cmtproto.ValidatorSet
					trustedVals *cmtproto.ValidatorSet
				)
				signedHeader := &cmtproto.SignedHeader{
					Header: &cmtproto.Header{
						AppHash:            []byte("appHash"),
						ProposerAddress:    []byte("sequencerAddr"),
						Time:               time.Unix(1724392989, 0),
						NextValidatorsHash: []byte("nextValHash"),
					},
					Commit: &cmtproto.Commit{},
				}
				header := ibctm.Header{
					SignedHeader:      signedHeader,
					ValidatorSet:      valSet,
					TrustedHeight:     ibcclienttypes.MustParseHeight("1-1"),
					TrustedValidators: trustedVals,
				}
				clientMsg, err := ibcclienttypes.PackClientMessage(&header)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId:      "canon-client-id",
						ClientMessage: clientMsg,
						Signer:        "sequencerAddr",
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-has-canon-client": {
							RollappId: "rollapp-has-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-has-canon-client": {
							1: {
								Sequencer: keepertest.Alice,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 1,
								},
								StartHeight: 1,
								NumBlocks:   1,
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
							2: {
								Sequencer: keepertest.Alice,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 2,
								},
								StartHeight: 2,
								NumBlocks:   1,
								BDs: rollapptypes.BlockDescriptors{
									BD: []rollapptypes.BlockDescriptor{
										{
											Height:    2,
											StateRoot: []byte("appHash2"),
											Timestamp: time.Unix(1724392989, 0),
										},
									},
								},
							},
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.ErrorIs(t, err, types.ErrStateRootsMismatch)
			},
		},
		{
			name: "Ensure state is compatible - happy path",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				sequencer := keepertest.Alice
				proposerAddr, err := k.GetSequencerPubKey(ctx, sequencer)
				require.NoError(t, err)
				proposerAddrBytes, err := proposerAddr.Marshal()
				require.NoError(t, err)
				blocktimestamp := time.Unix(1724392989, 0)
				k.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
				var (
					valSet      *cmtproto.ValidatorSet
					trustedVals *cmtproto.ValidatorSet
				)
				nextValsHash, err := k.GetSequencerHash(ctx, sequencer)
				require.NoError(t, err)
				signedHeader := &cmtproto.SignedHeader{
					Header: &cmtproto.Header{
						AppHash:            []byte("appHash"),
						ProposerAddress:    proposerAddrBytes,
						Time:               blocktimestamp,
						ValidatorsHash:     nextValsHash,
						NextValidatorsHash: nextValsHash,
					},
					Commit: &cmtproto.Commit{},
				}
				header := ibctm.Header{
					SignedHeader:      signedHeader,
					ValidatorSet:      valSet,
					TrustedHeight:     ibcclienttypes.MustParseHeight("1-1"),
					TrustedValidators: trustedVals,
				}
				clientMsg, err := ibcclienttypes.PackClientMessage(&header)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId:      "canon-client-id",
						ClientMessage: clientMsg,
						Signer:        "relayerAddr",
					},
					rollapps: map[string]rollapptypes.Rollapp{
						"rollapp-has-canon-client": {
							RollappId: "rollapp-has-canon-client",
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						"rollapp-has-canon-client": {
							1: {
								Sequencer: keepertest.Alice,
								StateInfoIndex: rollapptypes.StateInfoIndex{
									Index: 1,
								},
								StartHeight: 1,
								NumBlocks:   2,
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
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.LightClientKeeper(t)
			testClientStates := map[string]exported.ClientState{
				"non-tm-client-id": &ibcsolomachine.ClientState{},
				"canon-client-id": &ibctm.ClientState{
					ChainId: "rollapp-has-canon-client",
				},
			}
			ibcclientKeeper := NewMockIBCClientKeeper(testClientStates)
			ibcchannelKeeper := NewMockIBCChannelKeeper(nil)
			input := tc.prepare(ctx, *keeper)
			rollappKeeper := NewMockRollappKeeper(input.rollapps, input.stateInfos)
			ibcMsgDecorator := ante.NewIBCMessagesDecorator(*keeper, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)

			err := ibcMsgDecorator.HandleMsgUpdateClient(ctx, input.msg)
			tc.assert(ctx, *keeper, err)
		})
	}
}
