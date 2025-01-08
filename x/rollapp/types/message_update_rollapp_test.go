package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func TestMsgUpdateRollappInformation_ValidateBasic(t *testing.T) {
	// valid basic genesis info
	gInfo := GenesisInfo{
		Bech32Prefix:    bech32Prefix,
		GenesisChecksum: "checksum",
		NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
		InitialSupply:   sdk.NewInt(1000),
	}

	// valid modified genesis info
	gInfo2 := gInfo
	gInfo2.InitialSupply = sdk.NewInt(2000)
	gInfo2.GenesisAccounts = createManyGenesisAccounts(10)

	// invalid modified genesis info
	gInfo3 := gInfo
	gInfo3.InitialSupply = sdk.NewInt(20000000000000)
	gInfo3.GenesisAccounts = createManyGenesisAccounts(200)

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
				GenesisInfo:      &gInfo,
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					LogoUrl:     "https://dymension.xyz/logo.png",
					Telegram:    "https://t.me/rolly",
					X:           "https://x.dymension.xyz",
				},
			},
			err: nil,
		},
		{
			name: "valid - set initial sequencer to *",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				RollappId:        "dym_100-1",
				InitialSequencer: "*",
			},
			err: nil,
		},
		{
			name: "invalid owner address",
			msg: MsgUpdateRollappInformation{
				Owner:            "invalid_address",
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
			},
			err: ErrInvalidCreatorAddress,
		},
		{
			name: "invalid initial sequencer address",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: "invalid_address",
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo,
			},
			err: ErrInvalidInitialSequencer,
		},
		{
			name: "invalid metadata: invalid logo url",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo,
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					LogoUrl:     string(rune(0x7f)),
				},
			},
			err: ErrInvalidURL,
		},
		{
			name: "invalid metadata: too many tags",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo,
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					Tags:        []string{"tag1", "tag2", "tag3", "tag4"},
				},
			},
			err: ErrTooManyTags,
		},
		{
			name: "invalid metadata: invalid tag",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo,
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					Tags:        []string{"invalid"},
				},
			},
			err: ErrInvalidTag,
		},
		{
			name: "invalid metadata: duplicate tag",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo,
				Metadata: &RollappMetadata{
					Website:     "https://dymension.xyz",
					Description: "Sample description",
					Tags:        []string{"AI", "DeFi", "AI"},
				},
			},
			err: ErrDuplicateTag,
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
		{
			name: "valid: updating genesis info",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo2,
			},
			err: nil,
		},
		{
			name: "invalid: invalid genesis info",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo:      &gInfo3,
			},
			err: gerrc.ErrInvalidArgument,
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
