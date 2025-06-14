package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the necessary x/delayedack interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFinalizePacket{}, "delayedack/FinalizePacket", nil)
	cdc.RegisterConcrete(&MsgFinalizePacketByPacketKey{}, "delayedack/FinalizeByPacketKey", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "delayedack/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "delayedack/Params", nil)
}

// RegisterInterfaces registers interfaces types with the interface registry.
func RegisterInterfaces(reg types.InterfaceRegistry) {
	reg.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFinalizePacket{},
		&MsgFinalizePacketByPacketKey{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(reg, &_Msg_serviceDesc)
}
