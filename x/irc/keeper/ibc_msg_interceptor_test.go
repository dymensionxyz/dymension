package keeper_test

import (
	"testing"
	"time"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibcdmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/01-dymint/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	keeper "github.com/dymensionxyz/dymension/x/irc/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

func TestIRCMessageInterceptorDymCHain(t *testing.T) {
	var (
		rollappKeeper *rollappkeeper.Keeper
		keeper        *keeper.Keeper

		ctx sdk.Context

		clientState    exported.ClientState
		consensusState exported.ConsensusState
		header         exported.Header
		misbehaviour   exported.Misbehaviour

		height  uint64
		appHash []byte
	)
	rollappId := "rollapp1"

	tests := []struct {
		name     string
		malleate func()
		err      error
	}{
		{
			"valid state", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    height,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: height, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, nil,
		},
		{
			"valid multiple BDs, last BD", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    1,
					NumBlocks:      3,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 1, StateRoot: []byte{}, IntermediateStatesRoot: nil},
						{Height: 2, StateRoot: []byte{}, IntermediateStatesRoot: nil},
						{Height: 3, StateRoot: appHash, IntermediateStatesRoot: nil},
					}},
				})
			}, nil,
		},
		{
			"valid multiple BDs, middle BD", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    2,
					NumBlocks:      3,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 2, StateRoot: []byte{}, IntermediateStatesRoot: nil},
						{Height: 3, StateRoot: appHash, IntermediateStatesRoot: nil},
						{Height: 4, StateRoot: []byte{}, IntermediateStatesRoot: nil},
					}},
				})
			}, nil,
		},
		{
			"valid multiple StateInfo", func() {
				height = 12
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex1 := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				stateInfoIndex2 := types.StateInfoIndex{RollappId: rollappId, Index: 2}
				stateInfoIndex3 := types.StateInfoIndex{RollappId: rollappId, Index: 3}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex3)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex1,
					StartHeight:    1,
					NumBlocks:      10},
				)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex2,
					StartHeight:    11,
					NumBlocks:      2,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 11, StateRoot: nil, IntermediateStatesRoot: nil},
						{Height: 12, StateRoot: appHash, IntermediateStatesRoot: nil},
						{Height: 13, StateRoot: nil, IntermediateStatesRoot: nil},
					}},
				})
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex3,
					StartHeight:    14,
					NumBlocks:      7},
				)
			}, nil,
		},
		{
			"unknown rollappId", func() {
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: "unknown"})
			}, types.ErrUnknownRollappID,
		},
		{
			"invalid height=0", func() {
				height = 0
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
			}, types.ErrInvalidHeight,
		},
		{
			"state not fainalized", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    height,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_RECEIVED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: height, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, types.ErrHeightStateNotFainalized,
		},
		{
			"invalid app hash", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    height,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: height, StateRoot: []byte{1, 2, 3, 4, 57}, IntermediateStatesRoot: nil}}},
				})
			}, types.ErrInvalidAppHash,
		},
		{
			"LatestStateInfoIndex wasn't found", func() {
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
			}, sdkerrors.ErrLogic,
		},
		{
			"StateInfo wasn't found", func() {
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
			}, sdkerrors.ErrLogic,
		},
		{
			"No such state in height", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    2,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 2, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, types.ErrStateNotExists,
		},
		{
			"No such state in lower height", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 2}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    4,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 4, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, sdkerrors.ErrLogic,
		},
		{
			"No such state at all - got to stateIndex=0 ", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    4,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 4, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, sdkerrors.ErrLogic,
		},
		{
			"BDs array is wrong", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    1,
					NumBlocks:      4,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: 3, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, sdkerrors.ErrLogic,
		},
		{
			"high mismatch in BD", func() {
				height = 3
				appHash = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 255}
				stateInfoIndex := types.StateInfoIndex{RollappId: rollappId, Index: 1}
				rollappKeeper.SetRollapp(ctx, types.Rollapp{RollappId: rollappId})
				rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfoIndex)
				rollappKeeper.SetStateInfo(ctx, types.StateInfo{
					StateInfoIndex: stateInfoIndex,
					StartHeight:    height,
					NumBlocks:      1,
					Status:         types.STATE_STATUS_FINALIZED,
					BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
						{Height: height + 1, StateRoot: appHash, IntermediateStatesRoot: nil}}},
				})
			}, sdkerrors.ErrLogic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			keeper, rollappKeeper, ctx = keepertest.IRCKeeper(t)

			tt.malleate()

			// build client state
			clientState = ibcdmtypes.NewClientState(
				rollappId,
				time.Duration(0),
				time.Duration(0),
				clienttypes.Height{
					RevisionNumber: 0,
					RevisionHeight: height,
				},
				nil, nil,
			)

			// build consensus state
			root := commitmenttypes.MerkleRoot{Hash: appHash}
			consensusState = ibcdmtypes.NewConsensusState(time.Time{}, root, nil)

			// build header
			h := tmtypes.Header{
				ChainID: rollappId,
				Height:  int64(height),
				AppHash: appHash,
			}
			signedHeader := &tmproto.SignedHeader{
				Header: h.ToProto(),
				Commit: nil,
			}
			header = &ibcdmtypes.Header{
				SignedHeader: signedHeader,
				TrustedHeight: clienttypes.Height{
					RevisionNumber: 0,
					RevisionHeight: height,
				},
			}

			// build misbehaviour
			misbehaviour = ibcdmtypes.NewMisbehaviour("clientID", header.(*ibcdmtypes.Header), header.(*ibcdmtypes.Header))

			// check CreateClientValidate
			if err := keeper.CreateClientValidate(ctx, clientState, consensusState); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			// check UpdateClientValidate
			if err := keeper.UpdateClientValidate(ctx, "clientID", header); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			// check UpgradeClientValidate
			if err := keeper.UpgradeClientValidate(ctx, "clientID", clientState, consensusState, nil, nil); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			// check SubmitMisbehaviourValidate
			if err := keeper.SubmitMisbehaviourValidate(ctx, misbehaviour); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIRCMessageInterceptorNotDymChain(t *testing.T) {
	var (
		keeper *keeper.Keeper

		ctx sdk.Context

		clientState    exported.ClientState
		consensusState exported.ConsensusState
		header         exported.Header
		misbehaviour   exported.Misbehaviour
	)

	tests := []struct {
		name     string
		malleate func()
		err      error
	}{
		{
			"valid state", func() {
			}, nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			keeper, _, ctx = keepertest.IRCKeeper(t)

			tt.malleate()

			// build client state
			clientState = &ibctmtypes.ClientState{ChainId: "chain1"}

			// build consensus state
			consensusState = &ibctmtypes.ConsensusState{}

			// build header
			header = &ibctmtypes.Header{
				SignedHeader: &tmproto.SignedHeader{
					Header: (&tmtypes.Header{ChainID: "chain1"}).ToProto(),
					Commit: nil,
				},
				TrustedHeight: clienttypes.Height{},
			}

			// check CreateClientValidate
			if err := keeper.CreateClientValidate(ctx, clientState, consensusState); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			// check UpdateClientValidate
			if err := keeper.UpdateClientValidate(ctx, "clientID", header); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			// check UpgradeClientValidate
			if err := keeper.UpgradeClientValidate(ctx, "clientID", clientState, consensusState, nil, nil); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}

			// build misbehaviour
			misbehaviour = ibctmtypes.NewMisbehaviour("clientID", header.(*ibctmtypes.Header), header.(*ibctmtypes.Header))
			// check SubmitMisbehaviourValidate
			if err := keeper.SubmitMisbehaviourValidate(ctx, misbehaviour); tt.err != nil {
				require.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
