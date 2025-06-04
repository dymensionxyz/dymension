package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgIndicateProgress{}, "kas/IndicateProgress", nil)
	cdc.RegisterConcrete(&MsgBootstrap{}, "kas/Bootstrap", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgIndicateProgress{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgBootstrap{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
