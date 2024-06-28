package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	types "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	v2 "github.com/dymensionxyz/dymension/v3/x/rollapp/types/v2"
)

var hash32 = []byte("12345678901234567890123456789012")

func TestMsgUpdateState_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  v2.MsgUpdateState
		err  error
	}{
		{
			name: "valid initial state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   1,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
				}},
			},
		}, {
			name: "valid initial state with 3 blocks",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
					{Height: 2, StateRoot: hash32},
					{Height: 3, StateRoot: hash32},
				}},
			},
		}, {
			name: "valid state from known state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   1,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
				}},
			},
		}, {
			name: "valid state from known state with 3 blocks",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
					{Height: 5, StateRoot: hash32},
					{Height: 6, StateRoot: hash32},
				}},
			},
		}, {
			name: "invalid address",
			msg: v2.MsgUpdateState{
				Creator:     "invalid_address",
				StartHeight: 1,
				NumBlocks:   1,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidAddress,
		}, {
			name: "invalid zero blocks initial state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   0,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{},
			},
			err: types.ErrInvalidNumBlocks,
		}, {
			name: "invalid zero blocks from known state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 8,
				NumBlocks:   0,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{},
			},
			err: types.ErrInvalidNumBlocks,
		}, {
			name: "BlockDescriptors length mismatch",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
					{Height: 5, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidNumBlocks,
		}, {
			name: "wrong block height 0",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 0,
				NumBlocks:   2,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 0, StateRoot: hash32},
					{Height: 1, StateRoot: hash32},
				}},
			},
			err: types.ErrWrongBlockHeight,
		}, {
			name: "last block height mismatch",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 2,
				NumBlocks:   2,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
					{Height: 5, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidBlockSequence,
		}, {
			name: "num blocks overflow",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				NumBlocks: ^uint64(0),
			},
			err: types.ErrInvalidNumBlocks,
		}, {
			name: "initial state error",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   2,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
					{Height: 5, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidBlockSequence,
		}, {
			name: "initial state error one block",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   1,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence initial state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
					{Height: 2, StateRoot: hash32},
					{Height: 4, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence skip one from known state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
					{Height: 5, StateRoot: hash32},
					{Height: 7, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidBlockSequence,
		}, {
			name: "illegal sequence not sorted from known state",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 4,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 4, StateRoot: hash32},
					{Height: 6, StateRoot: hash32},
					{Height: 5, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidBlockSequence,
		}, {
			name: "illegal invalid state root empty",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
					{Height: 2},
					{Height: 3, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidStateRoot,
		}, {
			name: "illegal invalid state root small",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
					{Height: 2, StateRoot: []byte("1")},
					{Height: 3, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidStateRoot,
		}, {
			name: "illegal invalid state root big",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   3,
				DAPath: &types.DAPath{
					DaType: "interchainda",
				},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
					{Height: 2, StateRoot: []byte("112345678901234567890123456789012")},
					{Height: 3, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidStateRoot,
		}, {
			name: "empty dapath type",
			msg: v2.MsgUpdateState{
				Creator:     sample.AccAddress(),
				StartHeight: 1,
				NumBlocks:   1,
				DAPath:      &types.DAPath{},
				BDs: types.BlockDescriptors{BD: []types.BlockDescriptor{
					{Height: 1, StateRoot: hash32},
				}},
			},
			err: types.ErrInvalidDaClientType,
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
