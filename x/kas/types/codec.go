package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())

func RegisterCodec(cdc *codec.LegacyAmino) {
	// cdc.RegisterConcrete(&MsgCreateSequencer{}, "sequencer/CreateSequencer", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// registry.RegisterImplementations((*sdk.Msg)(nil),
	// &MsgCreateSequencer{},
	// )

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
