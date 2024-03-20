package types

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateRollapp_ValidateBasic(t *testing.T) {
	seqDupAddr := sample.AccAddress()

	var tooManyAddresses []string
	for i := 0; i < 200; i++ {
		tooManyAddresses = append(tooManyAddresses, sample.AccAddress())
	}
	var validNumberAddresses []string
	for i := 0; i < 100; i++ {
		validNumberAddresses = append(validNumberAddresses, sample.AccAddress())
	}
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
			err: ErrInvalidAddress,
		},
		{
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
			},
		},
		{
			name: "no max sequencers set",
			msg: MsgCreateRollapp{
				Creator: sample.AccAddress(),
			},
		},
		{
			name: "valid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				PermissionedAddresses: []string{sample.AccAddress(), sample.AccAddress()},
			},
		},
		{
			name: "duplicate permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				PermissionedAddresses: []string{seqDupAddr, seqDupAddr},
			},
			err: ErrPermissionedAddressesDuplicate,
		},
		{
			name: "invalid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				PermissionedAddresses: []string{seqDupAddr, "invalid permissioned address"},
			},
			err: ErrInvalidPermissionedAddress,
		},
		{
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
		},
		{
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
		{
			name: "more addresses than sequencers", // just trigger one case to see if validation is done or not
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         1,
				PermissionedAddresses: validNumberAddresses,
			},
			err: ErrTooManyPermissionedAddresses,
		},
		{
			name: "too many sequencers", // just trigger one case to see if validation is done or not
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         200,
				PermissionedAddresses: tooManyAddresses,
			},
			err: ErrInvalidMaxSequencers,
		},
		{
			name: "max sequencer not set",
			msg: MsgCreateRollapp{
				Creator: sample.AccAddress(),
			},
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
