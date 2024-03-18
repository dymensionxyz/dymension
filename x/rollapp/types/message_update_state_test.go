package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

var hash32 = []byte("12345678901234567890123456789012")

func TestMsgUpdateState_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateState
		err  error
	}{
		{
			name: "valid initial state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   1,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
		}, {
			name: "valid initial state with 3 blocks",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 2, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 3, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
		}, {
			name: "valid state from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   1,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
		}, {
			name: "valid state from known state with 3 blocks",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 5, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 6, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
		}, {
			name: "invalid address",
			msg: MsgUpdateState{
				Creator:     "invalid_address",
				StartHeight: 1,
				NumBlocks:   1,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid zero blocks initial state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   0,
				BDs:         BlockDescriptors{},
			},
			err: ErrInvalidNumBlocks,
		}, {
			name: "invalid zero blocks from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 8,
				NumBlocks:   0,
				BDs:         BlockDescriptors{},
			},
			err: ErrInvalidNumBlocks,
		}, {
			name: "BlockDescriptors length mismatch",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 5, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidNumBlocks,
		}, {
			name: "wrong block height 0",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   2,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 0, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrWrongBlockHeight,
		}, {
			name: "last block height mismatch",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 2,
				NumBlocks:   2,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 5, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "initial state error",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   2,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 5, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "initial state error one block",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   1,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence initial state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 2, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence skip one from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 5, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 7, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence not sorted from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 4, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 6, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 5, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal invalid state root empty",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 2, IntermediateStatesRoot: hash32},
					{Height: 3, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidStateRoot,
		}, {
			name: "illegal invalid state root small",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 2, StateRoot: []byte("1"), IntermediateStatesRoot: hash32},
					{Height: 3, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidStateRoot,
		}, {
			name: "illegal invalid state root big",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 2, StateRoot: []byte("112345678901234567890123456789012"), IntermediateStatesRoot: hash32},
					{Height: 3, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidStateRoot,
		}, {
			name: "illegal invalid intermediate state root",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				BDs: BlockDescriptors{BD: []BlockDescriptor{
					{Height: 1, StateRoot: hash32, IntermediateStatesRoot: hash32},
					{Height: 2, StateRoot: hash32, IntermediateStatesRoot: []byte("112345678901234567890123456789012")},
					{Height: 3, StateRoot: hash32, IntermediateStatesRoot: hash32},
				}},
			},
			err: ErrInvalidIntermediateStatesRoot,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
