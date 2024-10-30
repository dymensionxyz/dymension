package ante_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cometprototypes "github.com/cometbft/cometbft/proto/tendermint/types"
	comettypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
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

func ConvertValidator(src comettypes.Validator) *cometprototypes.Validator {
	// TODO: surely this must already exist somewhere

	dst := &cometprototypes.Validator{}
	pk, err := cryptocodec.FromTmPubKeyInterface(src.PubKey)
	if err != nil {
		panic(err)
	}
	pkP, err := cryptocodec.ToTmProtoPublicKey(pk)
	if err != nil {
		panic(err)
	}
	dst = &cometprototypes.Validator{
		Address:          src.Address,
		VotingPower:      src.VotingPower,
		ProposerPriority: src.ProposerPriority,
		PubKey:           pkP,
	}
	return dst
}

func ConvertValidatorSet(src *comettypes.ValidatorSet) *cometprototypes.ValidatorSet {
	// TODO: surely this must already exist somewhere

	if src == nil {
		return nil
	}

	dst := &cometprototypes.ValidatorSet{
		Validators: make([]*cometprototypes.Validator, len(src.Validators)),
	}

	for i, validator := range src.Validators {
		dst.Validators[i] = ConvertValidator(*validator)
	}
	dst.TotalVotingPower = src.TotalVotingPower()
	dst.Proposer = ConvertValidator(*src.Proposer)

	return dst
}

func TestHandleMsgUpdateClientGood(t *testing.T) {
	k, ctx := keepertest.LightClientKeeper(t)
	testClientStates := map[string]exported.ClientState{
		"non-tm-client-id": &ibcsolomachine.ClientState{},
	}
	testClientStates[keepertest.CanonClientID] = &ibctm.ClientState{
		ChainId: keepertest.DefaultRollapp,
	}

	blocktimestamp := time.Unix(1724392989, 0)
	var (
		trustedVals *cmtproto.ValidatorSet
	)
	signedHeader := &cmtproto.SignedHeader{
		Header: &cmtproto.Header{
			AppHash:            []byte("appHash"),
			ProposerAddress:    keepertest.Alice.MustProposerAddr(),
			Time:               blocktimestamp,
			ValidatorsHash:     keepertest.Alice.MustValsetHash(),
			NextValidatorsHash: keepertest.Alice.MustValsetHash(),
			Height:             1,
		},
		Commit: &cmtproto.Commit{},
	}
	header := ibctm.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      ConvertValidatorSet(keepertest.Alice.MustValset()),
		TrustedHeight:     ibcclienttypes.MustParseHeight("1-1"),
		TrustedValidators: trustedVals,
	}

	rollapps := map[string]rollapptypes.Rollapp{
		keepertest.DefaultRollapp: {
			RollappId: keepertest.DefaultRollapp,
		},
	}
	stateInfos := map[string]map[uint64]rollapptypes.StateInfo{
		keepertest.DefaultRollapp: {
			1: {
				Sequencer: keepertest.Alice.Address,
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
							Timestamp: header.SignedHeader.Header.Time,
						},
						{
							Height:    2,
							StateRoot: []byte("appHash2"),
							Timestamp: header.SignedHeader.Header.Time.Add(1),
						},
					},
				},
			},
		},
	}

	ibcclientKeeper := NewMockIBCClientKeeper(testClientStates)
	ibcchannelKeeper := NewMockIBCChannelKeeper(nil)
	rollappKeeper := NewMockRollappKeeper(rollapps, stateInfos)
	ibcMsgDecorator := ante.NewIBCMessagesDecorator(*k, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)

	clientMsg, err := ibcclienttypes.PackClientMessage(&header)
	require.NoError(t, err)
	msg := &ibcclienttypes.MsgUpdateClient{
		ClientId:      keepertest.CanonClientID,
		ClientMessage: clientMsg,
		Signer:        "relayerAddr",
	}
	err = ibcMsgDecorator.HandleMsgUpdateClient(ctx, msg)
	require.NoError(t, err)
}

func basicheaderLeagacy() ibctm.Header {
	blocktimestamp := time.Unix(1724392989, 0)
	var (
		valSet      *cmtproto.ValidatorSet
		trustedVals *cmtproto.ValidatorSet
	)
	signedHeader := &cmtproto.SignedHeader{
		Header: &cmtproto.Header{
			AppHash:            []byte("appHash"),
			ProposerAddress:    keepertest.Alice.MustProposerAddr(),
			Time:               blocktimestamp,
			ValidatorsHash:     keepertest.Alice.MustValsetHash(),
			NextValidatorsHash: keepertest.Alice.MustValsetHash(),
			Height:             1,
		},
		Commit: &cmtproto.Commit{},
	}
	header := ibctm.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      valSet,
		TrustedHeight:     ibcclienttypes.MustParseHeight("1-1"),
		TrustedValidators: trustedVals,
	}
	return header
}

func TestHandleMsgUpdateClientLegacy(t *testing.T) {
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
			name: "Ensure state is compatible - happy path",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
				header := basicheaderLeagacy()
				clientMsg, err := ibcclienttypes.PackClientMessage(&header)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId:      keepertest.CanonClientID,
						ClientMessage: clientMsg,
						Signer:        "relayerAddr",
					},
					rollapps: map[string]rollapptypes.Rollapp{
						keepertest.DefaultRollapp: {
							RollappId: keepertest.DefaultRollapp,
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						keepertest.DefaultRollapp: {
							1: {
								Sequencer: keepertest.Alice.Address,
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
											Timestamp: header.SignedHeader.Header.Time,
										},
										{
											Height:    2,
											StateRoot: []byte("appHash2"),
											Timestamp: header.SignedHeader.Header.Time.Add(1),
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
		{
			name: "Could not find a client with given client id",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				header := basicheaderLeagacy()
				clientMsg, err := ibcclienttypes.PackClientMessage(&header)
				require.NoError(t, err)
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId:      "non-existent-client",
						ClientMessage: clientMsg,
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
						ClientId: keepertest.CanonClientID,
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
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
				var valSet, trustedVals *cmtproto.ValidatorSet
				signedHeader := &cmtproto.SignedHeader{
					Header: &cmtproto.Header{
						ValidatorsHash: keepertest.Alice.MustValsetHash(),
						Height:         1,
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
						ClientId:      keepertest.CanonClientID,
						ClientMessage: clientMsg,
						Signer:        "relayerAddr",
					},
					rollapps: map[string]rollapptypes.Rollapp{
						keepertest.DefaultRollapp: {
							RollappId: keepertest.DefaultRollapp,
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						keepertest.DefaultRollapp: {
							3: {
								Sequencer: keepertest.Alice.Address,
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
				seq, err := k.GetSigner(ctx, keepertest.CanonClientID, 1)
				require.NoError(t, err)
				require.Equal(t, keepertest.Alice.Address, seq)
			},
		},
		{
			name: "State is incompatible - do not accept",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
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
						Height:             1,
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
						ClientId:      keepertest.CanonClientID,
						ClientMessage: clientMsg,
						Signer:        "sequencerAddr",
					},
					rollapps: map[string]rollapptypes.Rollapp{
						keepertest.DefaultRollapp: {
							RollappId: keepertest.DefaultRollapp,
						},
					},
					stateInfos: map[string]map[uint64]rollapptypes.StateInfo{
						keepertest.DefaultRollapp: {
							1: {
								Sequencer: keepertest.Alice.Address,
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
								Sequencer: keepertest.Alice.Address,
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
			name: "Client is not a known canonical client of a rollapp",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId: keepertest.CanonClientID,
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "SubmitMisbehavior for a canonical chain",
			prepare: func(ctx sdk.Context, k keeper.Keeper) testInput {
				k.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
				m := &ibctm.Misbehaviour{}
				mAny, _ := ibcclienttypes.PackClientMessage(m)

				return testInput{
					msg: &ibcclienttypes.MsgUpdateClient{
						ClientId:      keepertest.CanonClientID,
						ClientMessage: mAny,
					},
					rollapps: map[string]rollapptypes.Rollapp{
						keepertest.DefaultRollapp: {
							RollappId: keepertest.DefaultRollapp,
						},
					},
				}
			},
			assert: func(ctx sdk.Context, k keeper.Keeper, err error) {
				require.Error(t, err)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.LightClientKeeper(t)
			testClientStates := map[string]exported.ClientState{
				"non-tm-client-id": &ibcsolomachine.ClientState{},
				keepertest.CanonClientID: &ibctm.ClientState{
					ChainId: keepertest.DefaultRollapp,
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
