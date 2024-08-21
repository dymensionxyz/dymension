package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateRollapp{}, "rollapp/CreateRollapp", nil)
	cdc.RegisterConcrete(&MsgUpdateRollappInformation{}, "rollapp/UpdateRollappInformation", nil)
	cdc.RegisterConcrete(&MsgTransferOwnership{}, "rollapp/TransferDymNameOwnership", nil)
	cdc.RegisterConcrete(&MsgUpdateState{}, "rollapp/UpdateState", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateRollapp{},
		&MsgUpdateRollappInformation{},
		&MsgTransferOwnership{},
		&MsgUpdateState{},
	)
	registry.RegisterImplementations((*govtypes.Content)(nil), &SubmitFraudProposal{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
