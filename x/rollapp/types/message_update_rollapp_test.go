package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestMsgUpdateRollappInformation_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateRollappInformation
		err  error
	}{
		{
			name: "valid - full features",
			msg: MsgUpdateRollappInformation{
				Creator:                 sample.AccAddress(),
				RollappId:               "dym_100-1",
				InitialSequencerAddress: sample.AccAddress(),
				GenesisChecksum:         "checksum",
				Metadata: &RollappMetadata{
					Website:          "https://dymension.xyz",
					Description:      "Sample description",
					LogoDataUri:      "data:image/png;base64,c2lzZQ==",
					TokenLogoDataUri: "data:image/png;base64,ZHVwZQ==",
					Telegram:         "rolly",
					X:                "rolly",
				},
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgUpdateRollappInformation{
				Creator:                 sample.AccAddress(),
				InitialSequencerAddress: "invalid_address",
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
			},
			err: ErrInvalidInitialSequencerAddress,
		},
		{
			name: "invalid metadata: invalid logo data uri",
			msg: MsgUpdateRollappInformation{
				Creator:                 sample.AccAddress(),
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         "checksum",
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
			msg: MsgUpdateRollappInformation{
				Creator:                 sample.AccAddress(),
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisChecksum:         strings.Repeat("a", maxGenesisChecksumLength+1),
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
