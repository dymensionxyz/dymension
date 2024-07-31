package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterName{}, "dymns/RegisterName", nil)
	cdc.RegisterConcrete(&MsgTransferOwnership{}, "dymns/TransferOwnership", nil)
	cdc.RegisterConcrete(&MsgSetController{}, "dymns/SetController", nil)
	cdc.RegisterConcrete(&MsgUpdateResolveAddress{}, "dymns/UpdateResolveAddress", nil)
	cdc.RegisterConcrete(&MsgUpdateDetails{}, "dymns/UpdateDetails", nil)
	cdc.RegisterConcrete(&MsgPutAdsSellName{}, "dymns/PutAdsSellName", nil)
	cdc.RegisterConcrete(&MsgCancelAdsSellName{}, "dymns/CancelAdsSellName", nil)
	cdc.RegisterConcrete(&MsgPurchaseName{}, "dymns/PurchaseName", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterName{},
		&MsgTransferOwnership{},
		&MsgSetController{},
		&MsgUpdateResolveAddress{},
		&MsgUpdateDetails{},
		&MsgPutAdsSellName{},
		&MsgCancelAdsSellName{},
		&MsgPurchaseName{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&MigrateChainIdsProposal{},
		&UpdateAliasesProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
