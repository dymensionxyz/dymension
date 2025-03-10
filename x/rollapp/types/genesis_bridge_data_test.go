package types_test

import (
	"testing"

	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestGenesisBridgeData_ValidateBasic(t *testing.T) {
	validGenInfo := types.GenesisBridgeInfo{
		GenesisChecksum: "checksum",
		Bech32Prefix:    "prefix",
		NativeDenom: types.DenomMetadata{
			Base:     "base",
			Display:  "display",
			Exponent: 18,
		},
		InitialSupply:   math.NewInt(1000),
		GenesisAccounts: []types.GenesisAccount{},
	}

	validMetadata := banktypes.Metadata{
		Name:    "name",
		Symbol:  "symbol",
		Base:    "base",
		Display: "display",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "base",
				Exponent: 0,
			},
			{
				Denom:    "display",
				Exponent: 18,
			},
		},
	}

	validGenTransfer := transfertypes.NewFungibleTokenPacketData(
		"base",
		"1000",
		"sender",
		"receiver",
		"",
	)

	tests := []struct {
		name    string
		data    types.GenesisBridgeData
		wantErr bool
	}{
		{
			name: "valid data",
			data: types.GenesisBridgeData{
				GenesisInfo: validGenInfo,
				NativeDenom: validMetadata,
			},
			wantErr: false,
		},
		{
			name: "valid data with genesis transfer",
			data: types.GenesisBridgeData{
				GenesisInfo:     validGenInfo,
				NativeDenom:     validMetadata,
				GenesisTransfer: &validGenTransfer,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis info",
			data: types.GenesisBridgeData{
				GenesisInfo: types.GenesisBridgeInfo{
					GenesisChecksum: "",
					Bech32Prefix:    "prefix",
					NativeDenom: types.DenomMetadata{
						Base:     "base",
						Display:  "display",
						Exponent: 18,
					},
					InitialSupply:   math.NewInt(1000),
					GenesisAccounts: []types.GenesisAccount{},
				},
				NativeDenom: validMetadata,
			},
			wantErr: true,
		},
		{
			name: "invalid metadata",
			data: types.GenesisBridgeData{
				GenesisInfo: validGenInfo,
				NativeDenom: banktypes.Metadata{
					Base:    "base",
					Display: "", // missing display is not valid
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "base",
							Exponent: 0,
						},
						{
							Denom:    "display",
							Exponent: 18,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid genesis transfer",
			data: types.GenesisBridgeData{
				GenesisInfo: validGenInfo,
				NativeDenom: validMetadata,
				GenesisTransfer: &transfertypes.FungibleTokenPacketData{
					Denom:    "",
					Amount:   "",
					Sender:   "",
					Receiver: "",
					Memo:     "",
				},
			},
			wantErr: true,
		},
		{
			name: "metadata not matching genesis info denom",
			data: types.GenesisBridgeData{
				GenesisInfo: validGenInfo,
				NativeDenom: banktypes.Metadata{
					Base:    "base",
					Display: "NONdisplay",
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "base",
							Exponent: 0,
						},
						{
							Denom:    "NONdisplay",
							Exponent: 18,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "genesis transfer denom is wrong",
			data: types.GenesisBridgeData{
				GenesisInfo: validGenInfo,
				NativeDenom: validMetadata,
				GenesisTransfer: &transfertypes.FungibleTokenPacketData{
					Denom:    "NONbase",
					Amount:   "1000",
					Sender:   "sender",
					Receiver: "receiver",
					Memo:     "",
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate genesis accounts",
			data: types.GenesisBridgeData{
				GenesisInfo: types.GenesisBridgeInfo{
					GenesisChecksum: "checksum",
					Bech32Prefix:    "prefix",
					NativeDenom: types.DenomMetadata{
						Base:     "base",
						Display:  "display",
						Exponent: 18,
					},
					InitialSupply: math.NewInt(1000),
					GenesisAccounts: []types.GenesisAccount{
						{Address: "dym17g9cn4ss0h0dz5qhg2cg4zfnee6z3ftg3q6v58"},
						{Address: "dym17g9cn4ss0h0dz5qhg2cg4zfnee6z3ftg3q6v58"}, // duplicate account
					},
				},
				NativeDenom: validMetadata,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenesisBridgeInfo_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		info    types.GenesisBridgeInfo
		wantErr bool
	}{
		{
			name: "valid info",
			info: types.GenesisBridgeInfo{
				GenesisChecksum: "checksum",
				Bech32Prefix:    "prefix",
				NativeDenom: types.DenomMetadata{
					Base:     "base",
					Display:  "display",
					Exponent: 18,
				},
				InitialSupply:   math.NewInt(1000),
				GenesisAccounts: []types.GenesisAccount{},
			},
			wantErr: false,
		},
		{
			name: "missing fields",
			info: types.GenesisBridgeInfo{
				GenesisChecksum: "",
				Bech32Prefix:    "prefix",
				NativeDenom: types.DenomMetadata{
					Base:     "base",
					Display:  "display",
					Exponent: 18,
				},
				InitialSupply:   math.NewInt(1000),
				GenesisAccounts: []types.GenesisAccount{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.info.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
