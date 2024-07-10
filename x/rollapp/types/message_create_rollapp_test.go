package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

const bech32Prefix = "eth"

var genesisInfo = GenesisInfo{
	GenesisUrls:     []string{"https://example.com/genesis"},
	GenesisChecksum: "1234abcdef",
}

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
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				GenesisInfo:             genesisInfo,
			},
		},
		{
			name: "invalid rollappID",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               " ",
				GenesisInfo:             genesisInfo,
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
				GenesisInfo:             genesisInfo,
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
				GenesisInfo:             genesisInfo,
			},
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: "invalid_address",
				RollappId:               "dym_100-1",
				GenesisInfo:             genesisInfo,
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
				GenesisInfo:             genesisInfo,
			},
			err: ErrInvalidBech32Prefix,
		},
		{
			name: "empty genesis urls",
			msg: MsgCreateRollapp{
				Creator:                 sample.AccAddress(),
				Bech32Prefix:            bech32Prefix,
				InitialSequencerAddress: sample.AccAddress(),
				RollappId:               "dym_100-1",
				GenesisInfo: GenesisInfo{
					GenesisUrls:     nil,
					GenesisChecksum: "checksum",
				},
			},
			err: ErrEmptyGenesisURLs,
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
