package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFulfillOrder{}, "eibc/MsgFulfillOrder", nil)
	cdc.RegisterConcrete(&MsgFulfillOrderAuthorized{}, "eibc/MsgFulfillOrderAuthorized", nil)
	cdc.RegisterConcrete(&MsgUpdateDemandOrder{}, "eibc/MsgUpdateDemandOrder", nil)
	cdc.RegisterConcrete(&MsgCreateOnDemandLP{}, "eibc/CreateOnDemandLP", nil)
	cdc.RegisterConcrete(&MsgDeleteOnDemandLP{}, "eibc/DeleteOnDemandLP", nil)
	cdc.RegisterConcrete(&MsgTryFulfillOnDemand{}, "eibc/TryFulfillOnDemand", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "eibc/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "eibc/Params", nil)
	cdc.RegisterConcrete(&FulfillOrderAuthorization{}, "eibc/FulfillOrderAuthorization", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFulfillOrder{},
		&MsgFulfillOrderAuthorized{},
		&MsgUpdateDemandOrder{},
		&MsgCreateOnDemandLP{},
		&MsgDeleteOnDemandLP{},
		&MsgTryFulfillOnDemand{},
		&MsgUpdateParams{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&FulfillOrderAuthorization{},
	)
}
