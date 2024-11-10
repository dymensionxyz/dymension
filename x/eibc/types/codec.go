package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgFulfillOrder{}, "eibc/MsgFulfillOrder", nil)
	cdc.RegisterConcrete(&MsgFulfillOrderAuthorized{}, "eibc/MsgFulfillOrderAuthorized", nil)
	cdc.RegisterConcrete(&MsgUpdateDemandOrder{}, "eibc/MsgUpdateDemandOrder", nil)
	cdc.RegisterConcrete(&FulfillOrderAuthorization{}, "eibc/FulfillOrderAuthorization", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFulfillOrder{},
		&MsgFulfillOrderAuthorized{},
		&MsgUpdateDemandOrder{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&FulfillOrderAuthorization{},
	)
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
