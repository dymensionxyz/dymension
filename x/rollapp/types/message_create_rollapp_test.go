package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/shared/types"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateRollapp_ValidateBasic(t *testing.T) {
	seqDupAddr := sample.AccAddress()
	tests := []struct {
		name string
		msg  MsgCreateRollapp
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateRollapp{
				Creator:              "invalid_address",
				MaxSequencers:        1,
				MaxWithholdingBlocks: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:              sample.AccAddress(),
				MaxSequencers:        1,
				MaxWithholdingBlocks: 1,
			},
		}, {
			name: "invalid max sequencers",
			msg: MsgCreateRollapp{
				Creator:              sample.AccAddress(),
				MaxSequencers:        0,
				MaxWithholdingBlocks: 1,
			},
			err: ErrInvalidMaxSequencers,
		}, {
			name: "invalid max withholding blocks",
			msg: MsgCreateRollapp{
				Creator:              sample.AccAddress(),
				MaxSequencers:        1,
				MaxWithholdingBlocks: 0,
			},
			err: ErrInvalidMaxWithholding,
		}, {
			name: "valid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				MaxWithholdingBlocks:  1,
				PermissionedAddresses: types.Sequencers{Addresses: []string{sample.AccAddress(), sample.AccAddress()}},
			},
		}, {
			name: "duplicate permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				MaxWithholdingBlocks:  1,
				PermissionedAddresses: types.Sequencers{Addresses: []string{seqDupAddr, seqDupAddr}},
			},
			err: ErrPermissionedAddressesDuplicate,
		}, {
			name: "invalid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				MaxWithholdingBlocks:  1,
				PermissionedAddresses: types.Sequencers{Addresses: []string{seqDupAddr, "invalid permissioned address"}},
			},
			err: ErrInvalidPermissionedAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err, "test %s failed", tt.name)
				return
			}
			require.NoError(t, err)
		})
	}
}
