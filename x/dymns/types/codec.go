package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterCodec registers the necessary types and interfaces for the module
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterName{}, "dymns/RegisterName", nil)
	cdc.RegisterConcrete(&MsgTransferDymNameOwnership{}, "dymns/TransferDymNameOwnership", nil)
	cdc.RegisterConcrete(&MsgSetController{}, "dymns/SetController", nil)
	cdc.RegisterConcrete(&MsgUpdateResolveAddress{}, "dymns/UpdateResolveAddress", nil)
	cdc.RegisterConcrete(&MsgUpdateDetails{}, "dymns/UpdateDetails", nil)
	cdc.RegisterConcrete(&MsgPlaceSellOrder{}, "dymns/PlaceSellOrder", nil)
	cdc.RegisterConcrete(&MsgCancelSellOrder{}, "dymns/CancelSellOrder", nil)
	cdc.RegisterConcrete(&MsgPurchaseOrder{}, "dymns/PurchaseName", nil)
}

// RegisterInterfaces registers implementations by its interface, for the module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterName{},
		&MsgTransferDymNameOwnership{},
		&MsgSetController{},
		&MsgUpdateResolveAddress{},
		&MsgUpdateDetails{},
		&MsgPlaceSellOrder{},
		&MsgCancelSellOrder{},
		&MsgPurchaseOrder{},
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
