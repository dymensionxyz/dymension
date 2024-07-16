package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

const bech32Prefix = "eth"

func TestMsgCreateRollapp_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateRollapp
		err  error
	}{
		{
			name: "valid - full features",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				RollappId:               "dym_100-1",
				InitialSequencerAddress: sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				GenesisChecksum:         "checksum",
				Website:                 "https://dymension.xyz",
				Description:             "Sample description",
				LogoDataUri:             "https://dymension.xyz/logo.png",
				Alias:                   "Rollapp",
			},
		},
		{
			name: "invalid rollappID",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               " ",
				GenesisChecksum:         "checksum",
			},
			err: ErrInvalidRollappID,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateRollapp{
				Creator:                 "invalid_address",
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
			},
			err: ErrInvalidCreatorAddress,
		},
		{
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: "invalid_address",
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
			},
			err: ErrInvalidInitialSequencerAddress,
		},
		{
			name: "invalid bech32 prefix",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            "DYM",
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
			},
			err: ErrInvalidBech32Prefix,
		},
		{
			name: "empty genesis checksum",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "",
			},
			err: ErrEmptyGenesisChecksum,
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
