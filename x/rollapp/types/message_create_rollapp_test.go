package types

import (
	"strings"
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
				Alias:                   "Rollapp",
				Metadata: &RollappMetadata{
					Website:      "https://dymension.xyz",
					Description:  "Sample description",
					LogoDataUri:  "data:image/png;base64,c2lzZQ==",
					TokenLogoUri: "data:image/png;base64,ZHVwZQ==",
					Telegram:     "rolly",
					X:            "rolly",
				},
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
				Alias:                   "Rollapp",
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
				Alias:                   "Rollapp",
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
				Alias:                   "Rollapp",
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
				Alias:                   "Rollapp",
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
				Alias:                   "Rollapp",
			},
			err: ErrInvalidBech32Prefix,
		},
		{
			name: "invalid alias: too long",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
				Alias:                   strings.Repeat("a", maxAliasLength+1),
			},
			err: ErrInvalidAlias,
		},
		{
			name: "invalid metadata: invalid logo data uri",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
				Alias:                   "alias",
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					LogoDataUri: "invalid_uri",
				},
			},
			err: ErrInvalidLogoURI,
		},
		{
			name: "invalid genesis checksum: too long",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         strings.Repeat("a", maxGenesisChecksumLength+1),
				Alias:                   "alias",
			},
			err: ErrInvalidGenesisChecksum,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error(), "test %s failed", tt.name)
				return
			}
			require.NoError(t, err)
		})
	}
}
