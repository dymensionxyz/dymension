package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterAgent{}, "agent/RegisterAgent", nil)
	cdc.RegisterConcrete(&MsgDeactivateAgent{}, "agent/DeactivateAgent", nil)
	cdc.RegisterConcrete(&MsgUpdateAgentPolicy{}, "agent/UpdateAgentPolicy", nil)
	cdc.RegisterConcrete(&MsgSubmitAttestedAction{}, "agent/SubmitAttestedAction", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterAgent{},
		&MsgDeactivateAgent{},
		&MsgUpdateAgentPolicy{},
		&MsgSubmitAttestedAction{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
