package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/params"

	sdk "github.com/cosmos/cosmos-sdk/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	cryptocodec "github.com/evmos/ethermint/crypto/codec"
	ethermint "github.com/evmos/ethermint/types"
)

// makeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func makeEncodingConfig() params.EncodingConfig {
	amino := codec.NewLegacyAmino()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(codec, tx.DefaultSignModes)

	return params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := makeEncodingConfig()

	RegisterLegacyAminoCodec(encodingConfig.Amino)
	RegisterInterfaces(encodingConfig.InterfaceRegistry)

	keepers.ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	keepers.ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

// RegisterLegacyAminoCodec registers Interfaces from types, crypto, and SDK std.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)
}

// RegisterInterfaces registers Interfaces from types, crypto, and SDK std.
func RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	std.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	ethermint.RegisterInterfaces(interfaceRegistry)
}
