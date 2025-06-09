package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateSequencer{}, "sequencer/CreateSequencer", nil)
	cdc.RegisterConcrete(&MsgUnbond{}, "sequencer/Unbond", nil)
	cdc.RegisterConcrete(&MsgIncreaseBond{}, "sequencer/IncreaseBond", nil)
	cdc.RegisterConcrete(&MsgDecreaseBond{}, "sequencer/DecreaseBond", nil)
	cdc.RegisterConcrete(&MsgUpdateRewardAddress{}, "sequencer/UpdateRewardAddress", nil)
	cdc.RegisterConcrete(&MsgUpdateWhitelistedRelayers{}, "sequencer/UpdateWhitelistedRelayers", nil)
	cdc.RegisterConcrete(&MsgKickProposer{}, "sequencer/KickProposer", nil)
	cdc.RegisterConcrete(&MsgUpdateOptInStatus{}, "sequencer/UpdateOtpInStatus", nil)
	cdc.RegisterConcrete(&MsgPunishSequencer{}, "sequencer/PunishSequencer", nil)
	cdc.RegisterConcrete(&MsgUpdateSequencerInformation{}, "sequencer/UpdateSequencerInformation", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "sequencer/UpdateParams", nil)
	cdc.RegisterConcrete(Params{}, "sequencer/Params", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateSequencer{},
		&MsgUnbond{},
		&MsgIncreaseBond{},
		&MsgDecreaseBond{},
		&MsgUpdateRewardAddress{},
		&MsgUpdateWhitelistedRelayers{},
		&MsgKickProposer{},
		&MsgUpdateOptInStatus{},
		&MsgUpdateParams{},
		&MsgPunishSequencer{},
		&MsgUpdateSequencerInformation{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
