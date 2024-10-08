package types

import (
	fmt "fmt"
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
		// valid - updating genesis accounts
		{
			name: "valid: updating genesis accounts",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo: &GenesisInfo{
					GenesisAccounts: createManyGenesisAccounts(100),
				},
			},
			err: nil,
		},
		// invalid - updating genesis accounts: invalid address
		{
			name: "invalid: updating genesis accounts: invalid address",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo: &GenesisInfo{
					GenesisAccounts: &GenesisAccounts{
						Accounts: []GenesisAccount{
							{
								Address: "invalid_address",
								Amount:  sdk.NewInt(100),
							},
						},
					},
				},
			},
			err: fmt.Errorf("invalid"),
		},
		// invalid - too many genesis accounts
		{
			name: "invalid: too many genesis accounts",
			msg: MsgUpdateRollappInformation{
				Owner:            sample.AccAddress(),
				InitialSequencer: sample.AccAddress(),
				RollappId:        "dym_100-1",
				GenesisInfo: &GenesisInfo{
					GenesisAccounts: createManyGenesisAccounts(101),
				},
			},
			err: fmt.Errorf("too many genesis accounts"),
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

func createManyGenesisAccounts(n int) *GenesisAccounts {
	accounts := make([]GenesisAccount, n)
	for i := 0; i < n; i++ {
		accounts[i] = GenesisAccount{
			Address: sample.AccAddress(),
			Amount:  sdk.NewInt(100),
		}
	}
	return &GenesisAccounts{Accounts: accounts}
}
