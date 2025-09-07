package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
	cdc.RegisterConcrete(&MsgCreateAuction{}, "streamer/CreateAuction", nil)
	cdc.RegisterConcrete(&MsgTerminateAuction{}, "streamer/TerminateAuction", nil)
	cdc.RegisterConcrete(Params{}, "streamer/Params", nil)

	// Register legacy proposal types for backward compatibility with existing state
	cdc.RegisterConcrete(&CreateStreamProposal{}, "streamer/CreateStreamProposal", nil)
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
		&MsgCreateAuction{},
		&MsgTerminateAuction{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&CreateStreamProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
