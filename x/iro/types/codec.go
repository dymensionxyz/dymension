package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgBuy{}, "iro/Buy", nil)
	cdc.RegisterConcrete(&MsgSell{}, "iro/Sell", nil)
	cdc.RegisterConcrete(&MsgClaim{}, "iro/Claim", nil)
	cdc.RegisterConcrete(&MsgCreatePlan{}, "iro/CreatePlan", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "iro/UpdateParams", nil)
	cdc.RegisterConcrete(&MsgBuyExactSpend{}, "iro/BuyExactSpend", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgBuy{},
		&MsgSell{},
		&MsgClaim{},
		&MsgCreatePlan{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
