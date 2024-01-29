package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
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
				Creator:       "invalid_address",
				MaxSequencers: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
			},
		}, {
			name: "invalid max sequencers",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 0,
			},
			err: ErrInvalidMaxSequencers,
		}, {
			name: "valid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				PermissionedAddresses: []string{sample.AccAddress(), sample.AccAddress()},
			},
		}, {
			name: "duplicate permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				PermissionedAddresses: []string{seqDupAddr, seqDupAddr},
			},
			err: ErrPermissionedAddressesDuplicate,
		}, {
			name: "invalid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				PermissionedAddresses: []string{seqDupAddr, "invalid permissioned address"},
			},
			err: ErrInvalidPermissionedAddress,
		}, {
			name: "valid token metadata",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				Metadatas: []TokenMetadata{{
					Description: "valid",
					DenomUnits: []*DenomUnit{
						{Denom: "uvalid", Exponent: 0},
						{Denom: "valid", Exponent: 18},
					},
					Base:    "uvalid",
					Display: "valid",
					Name:    "valid",
					Symbol:  "VALID",
				}},
			},
			err: nil,
		}, {
			name: "invalid token metadata", // just trigger one case to see if validation is done or not
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				Metadatas: []TokenMetadata{{
					Description: "valid",
					DenomUnits: []*DenomUnit{
						{Denom: "uvalid", Exponent: 0},
						{Denom: "valid", Exponent: 18},
					},
					Base:    "uvalid",
					Display: "valid",
					Name:    "", // empty
					Symbol:  "VALID",
				}},
			},
			err: ErrInvalidTokenMetadata,
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
