package types

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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
				Creator:          sample.AccAddress(),
				RollappId:        "dym_100-1",
				InitialSequencer: sample.AccAddress(),
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
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
					GenesisUrl:  "https://genesis.dymension.xyz/file.json",
					DisplayName: "Rollapp",
					Tagline:     "Tagline",
					ExplorerUrl: "https://explorer.dymension.xyz",
				},
			},
		},
		{
			name: "invalid rollappID",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        " ",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidRollappID,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateRollapp{
				Creator:          "invalid_address",
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidCreatorAddress,
		},
		{
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: "invalid_address",
				RollappId:        "dym_100-1",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "multiple initial sequencer addresses",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: fmt.Sprintf("%s,%s,%s", sample.AccAddress(), sample.AccAddress(), sample.AccAddress()),
				RollappId:        "dym_100-1",
				Alias:            "Rollapp",
				VmType:           Rollapp_WASM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: nil,
		},
		{
			name: "all initial sequencers allowed",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: "*",
				RollappId:        "dym_100-1",
				Alias:            "Rollapp",
				VmType:           Rollapp_WASM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: nil,
		},
		{
			name: "invalid initial sequencer - duplicate address",
			msg: MsgCreateRollapp{
				Creator: sample.AccAddress(),
				InitialSequencer: fmt.Sprintf("%s,%s",
					sample.AccAddressFromSecret("same"),
					sample.AccAddressFromSecret("same")),
				RollappId: "dym_100-1",
				Alias:     "Rollapp",
				VmType:    Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "invalid bech32 prefix",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    "DYM",
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: gerrc.ErrInvalidArgument,
		},
		{
			name: "invalid metadata: invalid logo url",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "alias",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
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
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "alias",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: strings.Repeat("a", maxGenesisChecksumLength+1),
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
			},
			err: ErrInvalidGenesisChecksum,
		},
		{
			name: "invalid explorer url",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "alias",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.NewInt(1000),
				},
				Metadata: &RollappMetadata{
					ExplorerUrl: string(rune(0x7f)),
				},
			},
			err: ErrInvalidURL,
		},
		{
			name: "invalid initial supply",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				Alias:            "alias",
				VmType:           Rollapp_EVM,
				GenesisInfo: GenesisInfo{
					Bech32Prefix:    bech32Prefix,
					GenesisChecksum: "checksum",
					NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
					InitialSupply:   sdk.Int{},
				},
			},
			err: ErrInvalidInitialSupply,
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
