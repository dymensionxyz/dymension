package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the necessary x/otcbuyback interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateAuction{}, "otcbuyback/MsgCreateAuction", nil)
	cdc.RegisterConcrete(&MsgTerminateAuction{}, "otcbuyback/MsgTerminateAuction", nil)
	cdc.RegisterConcrete(&MsgBuy{}, "otcbuyback/MsgBuy", nil)
	cdc.RegisterConcrete(&MsgBuyExactSpend{}, "otcbuyback/MsgBuyExactSpend", nil)
	cdc.RegisterConcrete(&MsgClaimTokens{}, "otcbuyback/MsgClaimTokens", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "otcbuyback/MsgUpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "otcbuyback/Params", nil)
}

// RegisterInterfaces registers interfaces and implementations of the otcbuyback module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateAuction{},
		&MsgTerminateAuction{},
		&MsgBuy{},
		&MsgBuyExactSpend{},
		&MsgClaimTokens{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
