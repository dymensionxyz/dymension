package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the necessary types and interfaces for the module
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterName{}, "dymns/RegisterName", nil)
	cdc.RegisterConcrete(&MsgRegisterAlias{}, "dymns/RegisterAlias", nil)
	cdc.RegisterConcrete(&MsgTransferDymNameOwnership{}, "dymns/TransferDymNameOwnership", nil)
	cdc.RegisterConcrete(&MsgSetController{}, "dymns/SetController", nil)
	cdc.RegisterConcrete(&MsgUpdateResolveAddress{}, "dymns/UpdateResolveAddress", nil)
	cdc.RegisterConcrete(&MsgUpdateDetails{}, "dymns/UpdateDetails", nil)
	cdc.RegisterConcrete(&MsgPlaceSellOrder{}, "dymns/PlaceSellOrder", nil)
	cdc.RegisterConcrete(&MsgCompleteSellOrder{}, "dymns/CompleteSellOrder", nil)
	cdc.RegisterConcrete(&MsgCancelSellOrder{}, "dymns/CancelSellOrder", nil)
	cdc.RegisterConcrete(&MsgCancelBuyOrder{}, "dymns/CancelBuyOrder", nil)
	cdc.RegisterConcrete(&MsgAcceptBuyOrder{}, "dymns/AcceptBuyOrder", nil)
	cdc.RegisterConcrete(&MsgPurchaseOrder{}, "dymns/PurchaseName", nil)
	cdc.RegisterConcrete(&MsgMigrateChainIds{}, "dymns/MigrateChainIds", nil)
	cdc.RegisterConcrete(&MsgUpdateAliases{}, "dymns/UpdateAliases", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "dymns/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "dymns/Params", nil)
}

// RegisterInterfaces registers implementations by its interface, for the module
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterName{},
		&MsgRegisterAlias{},
		&MsgTransferDymNameOwnership{},
		&MsgSetController{},
		&MsgUpdateResolveAddress{},
		&MsgUpdateDetails{},
		&MsgUpdateParams{},
		&MsgPlaceSellOrder{},
		&MsgCompleteSellOrder{},
		&MsgCancelSellOrder{},
		&MsgCancelBuyOrder{},
		&MsgAcceptBuyOrder{},
		&MsgPurchaseOrder{},
		&MsgMigrateChainIds{},
		&MsgUpdateAliases{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
