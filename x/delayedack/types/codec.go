package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

// RegisterCodec registers the necessary x/delayedack interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFinalizePacket{}, "delayedack/FinalizePacket", nil)
	cdc.RegisterConcrete(&MsgFinalizePacketByPacketKey{}, "delayedack/MsgFinalizePacketByPacketKey", nil)
}

// RegisterInterfaces registers interfaces types with the interface registry.
func RegisterInterfaces(reg types.InterfaceRegistry) {
	reg.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFinalizePacket{},
		&MsgFinalizePacketByPacketKey{},
	)
	msgservice.RegisterMsgServiceDesc(reg, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(Amino)
	RegisterCodec(authzcodec.Amino)

	Amino.Seal()
}
