package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRollappGenesisEvent = "rollapp_genesis_event"

var _ sdk.Msg = &MsgRollappGenesisEvent{}

func NewMsgRollappGenesisEvent(address string, channel_id string, rollapp_id string) *MsgRollappGenesisEvent {
	return &MsgRollappGenesisEvent{
		Address:   address,
		ChannelId: channel_id,
		RollappId: rollapp_id,
	}
}

func (msg *MsgRollappGenesisEvent) Route() string {
	return RouterKey
}

func (msg *MsgRollappGenesisEvent) Type() string {
	return TypeMsgRollappGenesisEvent
}

func (msg *MsgRollappGenesisEvent) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRollappGenesisEvent) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRollappGenesisEvent) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address (%s)", err)
	}
	if msg.ChannelId == "" {
		return sdkerrors.Wrap(ErrInvalidGenesisChannelId, "channel id can not be empty")
	}
	if msg.RollappId == "" {
		return sdkerrors.Wrap(ErrInvalidRollappID, "rollapp id can not be empty")
	}
	return nil
}
