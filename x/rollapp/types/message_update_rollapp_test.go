package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
				GenesisInfo: &GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
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
			name: "valid - set initial sequencer to *",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				RollappId:        "dym_100-1",
				InitialSequencer: "*",
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: "invalid_address",
				RollappId:        "dym_100-1",
				GenesisInfo: &GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "invalid metadata: invalid logo url",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo: &GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
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
				GenesisInfo: &GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: strings.Repeat("a", maxGenesisChecksumLength+1),
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidGenesisChecksum,
		},
		{
			name: "valid: updating without genesis info",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      nil,
			},
			err: nil,
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
