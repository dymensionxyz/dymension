package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the necessary x/sponsorship interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgVote{}, "sponsorship/Vote", nil)
	cdc.RegisterConcrete(&MsgRevokeVote{}, "sponsorship/RevokeVote", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "sponsorship/UpdateParams", nil)
}

// RegisterInterfaces registers interfaces types with the interface registry.
func RegisterInterfaces(reg types.InterfaceRegistry) {
	reg.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgVote{},
		&MsgRevokeVote{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(reg, &_Msg_serviceDesc)
}
