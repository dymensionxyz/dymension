package genesisbridge_test

import (
	"testing"

	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/genesisbridge"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisBridgeData_ValidateBasic(t *testing.T) {
	validGenInfo := genesisbridge.GenesisBridgeInfo{
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
		data    genesisbridge.GenesisBridgeData
		wantErr bool
	}{
		{
			name: "valid data",
			data: genesisbridge.GenesisBridgeData{
				GenesisInfo: validGenInfo,
				NativeDenom: validMetadata,
			},
			wantErr: false,
		},
		{
			name: "valid data with genesis transfer",
			data: genesisbridge.GenesisBridgeData{
				GenesisInfo:     validGenInfo,
				NativeDenom:     validMetadata,
				GenesisTransfer: &validGenTransfer,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis info",
			data: genesisbridge.GenesisBridgeData{
				GenesisInfo: genesisbridge.GenesisBridgeInfo{
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
			data: genesisbridge.GenesisBridgeData{
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
			data: genesisbridge.GenesisBridgeData{
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
			data: genesisbridge.GenesisBridgeData{
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
			data: genesisbridge.GenesisBridgeData{
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
		info    genesisbridge.GenesisBridgeInfo
		wantErr bool
	}{
		{
			name: "valid info",
			info: genesisbridge.GenesisBridgeInfo{
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
			info: genesisbridge.GenesisBridgeInfo{
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
