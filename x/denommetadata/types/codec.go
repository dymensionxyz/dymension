package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterCodec registers the necessary x/denommetadata interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterHLTokenDenomMetadata{}, "denommetadata/MsgRegisterHLTokenDenomMetadata", nil)
}

// RegisterInterfaces registers interfaces and implementations of the denommetadata module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&CreateDenomMetadataProposal{},
		&UpdateDenomMetadataProposal{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRegisterHLTokenDenomMetadata{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
