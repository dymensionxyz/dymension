package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateRollapp{}, "rollapp/CreateRollapp", nil)
	cdc.RegisterConcrete(&MsgUpdateRollappInformation{}, "rollapp/UpdateRollappInformation", nil)
	cdc.RegisterConcrete(&MsgTransferOwnership{}, "rollapp/TransferDymNameOwnership", nil)
	cdc.RegisterConcrete(&MsgUpdateState{}, "rollapp/UpdateState", nil)
	cdc.RegisterConcrete(&MsgAddApp{}, "rollapp/AddApp", nil)
	cdc.RegisterConcrete(&MsgUpdateApp{}, "rollapp/UpdateApp", nil)
	cdc.RegisterConcrete(&MsgRemoveApp{}, "rollapp/RemoveApp", nil)
	cdc.RegisterConcrete(&MsgRollappFraudProposal{}, "rollapp/RollappFraudProposal", nil)
	cdc.RegisterConcrete(&MsgMarkObsoleteRollapps{}, "rollapp/MarkObsoleteRollapps", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "rollapp/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "rollapp/Params", nil)
	cdc.RegisterConcrete(&MsgForceGenesisInfoChange{}, "rollapp/ForceGenesisInfoChange", nil)
	cdc.RegisterConcrete(&GenesisInfo{}, "rollapp/GenesisInfo", nil)
	cdc.RegisterConcrete(&MsgToggleTEE{}, "rollapp/ToggleTEE", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateRollapp{},
		&MsgUpdateRollappInformation{},
		&MsgTransferOwnership{},
		&MsgUpdateState{},
		&MsgAddApp{},
		&MsgUpdateApp{},
		&MsgRemoveApp{},
		&MsgRollappFraudProposal{},
		&MsgMarkObsoleteRollapps{},
		&MsgForceGenesisInfoChange{},
		&MsgUpdateParams{},
		&MsgToggleTEE{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
