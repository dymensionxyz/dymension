package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the necessary x/auctionhouse interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgBuy{}, "auctionhouse/MsgBuy", nil)
	cdc.RegisterConcrete(&MsgBuyExactSpend{}, "auctionhouse/MsgBuyExactSpend", nil)
	cdc.RegisterConcrete(&MsgClaimTokens{}, "auctionhouse/MsgClaimTokens", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "auctionhouse/MsgUpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "auctionhouse/Params", nil)
}

// RegisterInterfaces registers interfaces and implementations of the auctionhouse module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgBuy{},
		&MsgBuyExactSpend{},
		&MsgClaimTokens{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
