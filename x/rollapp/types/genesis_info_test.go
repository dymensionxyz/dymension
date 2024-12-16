package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func TestGenesisInfo_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  GenesisInfo
		err  error
	}{
		{
			name: "valid - full features",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
				InitialSupply:   sdk.NewInt(1000),
			},
			err: nil,
		},

		{
			name: "invalid genesis checksum: too long",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: strings.Repeat("a", maxGenesisChecksumLength+1),
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
				InitialSupply:   sdk.NewInt(1000),
			},
			err: ErrInvalidGenesisChecksum,
		},

		{
			name: "invalid native denom",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 2},
				InitialSupply:   sdk.NewInt(1000),
			},
			err: ErrInvalidMetadata,
		},

		{
			name: "invalid initial supply",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
				InitialSupply:   sdk.NewInt(-1),
			},
			err: ErrInvalidInitialSupply,
		},
		{
			name: "valid - no native denom, zero supply",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				InitialSupply:   sdk.NewInt(0),
			},
			err: nil,
		},
		{
			name: "valid - no native denom, supply not set",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
			},
			err: nil,
		},
		{
			name: "no native denom, supply greater than 0",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				InitialSupply:   sdk.NewInt(1),
			},
			err: ErrNoNativeTokenRollapp,
		},
		{
			name: "no native denom, genesis accounts",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				GenesisAccounts: createManyGenesisAccounts(1),
			},
			err: ErrNoNativeTokenRollapp,
		},
		{
			name: "not enough supply",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
				InitialSupply:   sdk.NewInt(1),
				GenesisAccounts: createManyGenesisAccounts(100),
			},
			err: ErrInvalidInitialSupply,
		},

		// genesis accounts - invalid address
		{
			name: "genesis accounts - invalid address",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
				InitialSupply:   sdk.NewInt(1000),
				GenesisAccounts: &GenesisAccounts{Accounts: []GenesisAccount{{Address: "invalid", Amount: sdk.NewInt(100)}}},
			},
			err: gerrc.ErrInvalidArgument,
		},
		{
			name: "genesis accounts - too many accounts",
			msg: GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "checksum",
				NativeDenom:     DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
				InitialSupply:   sdk.NewInt(1000),
				GenesisAccounts: createManyGenesisAccounts(maxAllowedGenesisAccounts + 1),
			},
			err: ErrTooManyGenesisAccounts,
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
