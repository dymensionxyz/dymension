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
				Owner:            sample.AccAddress(),
				RollappId:        "dym_100-1",
				InitialSequencer: sample.AccAddress(),
				GenesisChecksum:  "checksum",
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					LogoUrl:     "https://dymension.xyz/logo.png",
					Telegram:    "https://t.me/rolly",
					X:           "https://x.dymension.xyz",
				},
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: "invalid_address",
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "invalid metadata: invalid logo url",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					LogoUrl:     string(rune(0x7f)),
				},
			},
			err: ErrInvalidURL,
		},
		{
			name: "invalid genesis checksum: too long",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  strings.Repeat("a", maxGenesisChecksumLength+1),
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
