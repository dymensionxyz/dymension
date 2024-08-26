package types

import (
	"fmt"
	"strings"
	"testing"

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
				Bech32Prefix:     bech32Prefix,
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					LogoUrl:     "https://dymension.xyz/logo.png",
					Telegram:    "https://t.me/rolly",
					X:           "https://x.dymension.xyz",
					GenesisUrl:  "https://genesis.dymension.xyz/file.json",
					DisplayName: "Rollapp",
					Tagline:     "Tagline",
					TokenSymbol: "ROLL",
					ExplorerUrl: "https://explorer.dymension.xyz",
				},
			},
		},
		{
			name: "invalid rollappID",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: sample.AccAddress(),
				RollappId:        " ",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
			},
			err: ErrInvalidRollappID,
		},
		{
			name: "invalid creator address",
			msg: MsgCreateRollapp{
				Creator:          "invalid_address",
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
			},
			err: ErrInvalidCreatorAddress,
		},
		{
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: "invalid_address",
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "multiple initial sequencer addresses",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: fmt.Sprintf("%s,%s,%s", sample.AccAddress(), sample.AccAddress(), sample.AccAddress()),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_WASM,
			},
			err: nil,
		},
		{
			name: "all initial sequencers allowed",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: "*",
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_WASM,
			},
			err: nil,
		},
		{
			name: "invalid initial sequencer - duplicate address",
			msg: MsgCreateRollapp{
				Creator:      sample.AccAddress(),
				Bech32Prefix: bech32Prefix,
				InitialSequencer: fmt.Sprintf("%s,%s",
					sample.AccAddressFromSecret("same"),
					sample.AccAddressFromSecret("same")),
				RollappId:       "dym_100-1",
				GenesisChecksum: "checksum",
				Alias:           "Rollapp",
				VmType:          Rollapp_EVM,
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "invalid bech32 prefix",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     "DYM",
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp",
				VmType:           Rollapp_EVM,
			},
			err: gerrc.ErrInvalidArgument,
		},
		{
			name: "invalid metadata: invalid logo url",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "alias",
				VmType:           Rollapp_EVM,
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
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  strings.Repeat("a", maxGenesisChecksumLength+1),
				Alias:            "alias",
				VmType:           Rollapp_EVM,
			},
			err: ErrInvalidGenesisChecksum,
		},
		{
			name: "invalid explorer url",
			msg: MsgCreateRollapp{
				Creator:          sample.AccAddress(),
				Bech32Prefix:     bech32Prefix,
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisChecksum:  "checksum",
				Alias:            "alias",
				VmType:           Rollapp_EVM,
				Metadata: &RollappMetadata{
					ExplorerUrl: string(rune(0x7f)),
				},
			},
			err: ErrInvalidURL,
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
