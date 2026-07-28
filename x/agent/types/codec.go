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
	cdc.RegisterConcrete(&MsgFundAgentEscrow{}, "agent/FundAgentEscrow", nil)
	cdc.RegisterConcrete(&MsgWithdrawAgentEscrow{}, "agent/WithdrawAgentEscrow", nil)
	cdc.RegisterConcrete(&MsgUpdateAgentSpendPolicy{}, "agent/UpdateAgentSpendPolicy", nil)
	cdc.RegisterConcrete(&MsgSubmitAttestedTransfer{}, "agent/SubmitAttestedTransfer", nil)
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
		&MsgFundAgentEscrow{},
		&MsgWithdrawAgentEscrow{},
		&MsgUpdateAgentSpendPolicy{},
		&MsgSubmitAttestedTransfer{},
		&MsgRevokePolicy{},
		&MsgUnrevokePolicy{},
		&MsgSubmitFeedback{},
		&MsgRevokeFeedback{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
