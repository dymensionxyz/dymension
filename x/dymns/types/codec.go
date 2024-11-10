package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterCodec registers the necessary types and interfaces for the module
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterName{}, "dymns/RegisterName", nil)
	cdc.RegisterConcrete(&MsgRegisterAlias{}, "dymns/RegisterAlias", nil)
	cdc.RegisterConcrete(&MsgTransferDymNameOwnership{}, "dymns/TransferDymNameOwnership", nil)
	cdc.RegisterConcrete(&MsgSetController{}, "dymns/SetController", nil)
	cdc.RegisterConcrete(&MsgUpdateResolveAddress{}, "dymns/UpdateResolveAddress", nil)
	cdc.RegisterConcrete(&MsgUpdateDetails{}, "dymns/UpdateDetails", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "dymns/UpdateParams", nil)
	cdc.RegisterConcrete(&MsgPlaceSellOrder{}, "dymns/PlaceSellOrder", nil)
	cdc.RegisterConcrete(&MsgCompleteSellOrder{}, "dymns/CompleteSellOrder", nil)
	cdc.RegisterConcrete(&MsgCancelSellOrder{}, "dymns/CancelSellOrder", nil)
	cdc.RegisterConcrete(&MsgCancelBuyOrder{}, "dymns/CancelBuyOrder", nil)
	cdc.RegisterConcrete(&MsgAcceptBuyOrder{}, "dymns/AcceptBuyOrder", nil)
	cdc.RegisterConcrete(&MsgPurchaseOrder{}, "dymns/PurchaseName", nil)
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
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
	sdk.RegisterLegacyAminoCodec(Amino)
	RegisterCodec(authzcodec.Amino)

	Amino.Seal()
}
