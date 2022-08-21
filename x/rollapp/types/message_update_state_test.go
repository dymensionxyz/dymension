package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateState_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateState
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateState{
				Creator:     "invalid_address",
				StartHeight: 0,
				NumBlocks:   1,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 0}}},
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid initial state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   1,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 0}}},
			},
		}, {
			name: "valid initial state with 3 blocks",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   3,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 0}, {Height: 1}, {Height: 2}}},
			},
		}, {
			name: "valid state from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   1,
				LastBD:      BlockDescriptor{Height: 3},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}}},
			},
		}, {
			name: "valid state from known state with 3 blocks",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				LastBD:      BlockDescriptor{Height: 3},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}, {Height: 5}, {Height: 6}}},
			},
		}, {
			name: "invalid zero blocks initial state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   0,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{},
			},
			err: ErrInvalidNumBlocks,
		}, {
			name: "invalid zero blocks from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 8,
				NumBlocks:   0,
				LastBD:      BlockDescriptor{Height: 7},
				BDs:         BlockDescriptors{},
			},
			err: ErrInvalidNumBlocks,
		}, {
			name: "BlockDescriptors length mismatch",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				LastBD:      BlockDescriptor{Height: 3},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}, {Height: 5}}},
			},
			err: ErrInvalidNumBlocks,
		}, {
			name: "last block height mismatch",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 2,
				NumBlocks:   2,
				LastBD:      BlockDescriptor{Height: 3},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}, {Height: 5}}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "initial state error",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   2,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}, {Height: 5}}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "initial state error one block",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   1,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence initial state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   3,
				LastBD:      BlockDescriptor{Height: 0},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 0}, {Height: 1}, {Height: 3}}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence skip one from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				LastBD:      BlockDescriptor{Height: 3},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}, {Height: 5}, {Height: 7}}},
			},
			err: ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence not sorted from known state",
			msg: MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				LastBD:      BlockDescriptor{Height: 3},
				BDs:         BlockDescriptors{BD: []BlockDescriptor{{Height: 4}, {Height: 6}, {Height: 5}}},
			},
			err: ErrInvalidBlockSequence,
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
