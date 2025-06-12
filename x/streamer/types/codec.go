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
	cdc.RegisterConcrete(&MsgCreateStream{}, "streamer/CreateStream", nil)
	cdc.RegisterConcrete(&DistrRecord{}, "streamer/DistrRecord", nil)
	cdc.RegisterConcrete(&MsgTerminateStream{}, "streamer/TerminateStream", nil)
	cdc.RegisterConcrete(&MsgReplaceStream{}, "streamer/ReplaceStream", nil)
	cdc.RegisterConcrete(&MsgUpdateStream{}, "streamer/UpdateStream", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "streamer/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "streamer/Params", nil)
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
