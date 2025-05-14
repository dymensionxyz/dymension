package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the necessary x/streamer interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateStream{}, "dymension/CreateStream", nil)
	cdc.RegisterConcrete(&MsgTerminateStream{}, "dymension/TerminateStream", nil)
	cdc.RegisterConcrete(&MsgReplaceStream{}, "dymension/ReplaceStream", nil)
	cdc.RegisterConcrete(&MsgUpdateStream{}, "dymension/UpdateStream", nil)
}

// RegisterInterfaces registers interfaces and implementations of the streamer module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*types.Msg)(nil),
		&MsgCreateStream{},
		&MsgTerminateStream{},
		&MsgReplaceStream{},
		&MsgUpdateStream{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
