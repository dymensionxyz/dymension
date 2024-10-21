package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFulfillOrder{}, "eibc/MsgFulfillOrder", nil)
	cdc.RegisterConcrete(&MsgFulfillOrderAuthorized{}, "eibc/MsgFulfillOrderAuthorized", nil)
	cdc.RegisterConcrete(&FulfillOrderAuthorization{}, "/dymensionxyz.dymension.eibc.FulfillOrderAuthorization", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFulfillOrder{},
		&MsgFulfillOrderAuthorized{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&FulfillOrderAuthorization{},
	)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
