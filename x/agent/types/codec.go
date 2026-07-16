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
	cdc.RegisterConcrete(&MsgRevokePolicy{}, "agent/RevokePolicy", nil)
	cdc.RegisterConcrete(&MsgUnrevokePolicy{}, "agent/UnrevokePolicy", nil)
	cdc.RegisterConcrete(&MsgSubmitFeedback{}, "agent/SubmitFeedback", nil)
	cdc.RegisterConcrete(&MsgRevokeFeedback{}, "agent/RevokeFeedback", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterAgent{},
		&MsgDeactivateAgent{},
		&MsgUpdateAgentPolicy{},
		&MsgSubmitAttestedAction{},
		&MsgRevokePolicy{},
		&MsgUnrevokePolicy{},
		&MsgSubmitFeedback{},
		&MsgRevokeFeedback{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
