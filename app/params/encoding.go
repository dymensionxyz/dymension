package params

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/std"

	sdk "github.com/cosmos/cosmos-sdk/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	ethcryptocodec "github.com/evmos/ethermint/crypto/codec"

	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	eip712 "github.com/evmos/ethermint/ethereum/eip712"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/address"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/gogoproto/proto"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeEncodingConfig creates a new EncodingConfig and returns it
func MakeEncodingConfig() sdktestutil.TestEncodingConfig {
	cdc := codec.NewLegacyAmino()

	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
		CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
			evmtypes.MsgEthereumTxCustomGetSigner.MsgType: evmtypes.MsgEthereumTxCustomGetSigner.Fn,
		},
	}

	interfaceRegistry, _ := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	codec := codec.NewProtoCodec(interfaceRegistry)
	RegisterLegacyAminoCodec(cdc)
	RegisterInterfaces(interfaceRegistry)

	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = cdc
	eip712.SetEncodingConfig(cdc, interfaceRegistry)

	RegisterMissingIBCAminoTypes(cdc)

	return sdktestutil.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             cdc,
	}
}

// RegisterLegacyAminoCodec registers Interfaces from types, crypto, and SDK std.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sdk.RegisterLegacyAminoCodec(cdc)
	codec.RegisterEvidences(cdc)
	ethcryptocodec.RegisterCrypto(cdc)
}

// Workaround: add manually amino register of ibc msgs
func RegisterMissingIBCAminoTypes(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&ibcclienttypes.MsgUpdateClient{}, "ibc/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&ibcchanneltypes.MsgRecvPacket{}, "ibc/MsgRecvPacket", nil)
}

// RegisterInterfaces registers Interfaces from types, crypto, and SDK std.
func RegisterInterfaces(interfaceRegistry codectypes.InterfaceRegistry) {
	std.RegisterInterfaces(interfaceRegistry)
	ethcryptocodec.RegisterInterfaces(interfaceRegistry)
	ethermint.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
}
