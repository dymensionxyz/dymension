package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateSequencer{}, "sequencer/CreateSequencer", nil)
	cdc.RegisterConcrete(&MsgUnbond{}, "sequencer/Unbond", nil)
	cdc.RegisterConcrete(&MsgIncreaseBond{}, "sequencer/IncreaseBond", nil)
	cdc.RegisterConcrete(&MsgDecreaseBond{}, "sequencer/DecreaseBond", nil)
	cdc.RegisterConcrete(&MsgUpdateRewardAddress{}, "sequencer/UpdateRewardAddress", nil)
	cdc.RegisterConcrete(&MsgUpdateWhitelistedRelayers{}, "sequencer/UpdateWhitelistedRelayers", nil)
	cdc.RegisterConcrete(&MsgKickProposer{}, "sequencer/KickProposer", nil)
	cdc.RegisterConcrete(&MsgUpdateOptInStatus{}, "sequencer/UpdateOtpInStatus", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateSequencer{},
		&MsgDecreaseBond{},
		&MsgUnbond{},
		&MsgIncreaseBond{},
		&MsgKickProposer{},
		&MsgUpdateOptInStatus{},
		&MsgUpdateRewardAddress{},
		&MsgUpdateWhitelistedRelayers{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino        = codec.NewLegacyAmino()
	InterfaceReg = cdctypes.NewInterfaceRegistry()
	ModuleCdc    = codec.NewProtoCodec(InterfaceReg)
)
