package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

var (
	_ sdk.Msg = &MsgFinalizePacket{}
	_ sdk.Msg = &MsgFinalizePacketByPacketKey{}
	_ sdk.Msg = &MsgUpdateParams{}
)

func (m MsgFinalizePacket) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "sender must be a valid bech32 address: %s", m.Sender),
		)
	}
	if len(m.RollappId) == 0 {
		return gerrc.ErrInvalidArgument.Wrap("rollappId must be non-empty")
	}
	if len(m.PacketSrcChannel) == 0 {
		return gerrc.ErrInvalidArgument.Wrap("packet src channel must be non-empty")
	}
	return nil
}

func (m MsgFinalizePacket) PendingPacketKey() []byte {
	return commontypes.RollappPacketKey(
		commontypes.Status_PENDING,
		m.RollappId,
		m.PacketProofHeight,
		m.PacketType,
		m.PacketSrcChannel,
		m.PacketSequence,
	)
}

func (m MsgFinalizePacketByPacketKey) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "sender must be a valid bech32 address: %s", m.Sender),
		)
	}
	if len(m.PacketKey) == 0 {
		return gerrc.ErrInvalidArgument.Wrap("rollappId must be non-empty")
	}

	if _, err := commontypes.DecodePacketKey(m.PacketKey); err != nil {
		return gerrc.ErrInvalidArgument.Wrap("packet key must be a valid base64 encoded string")
	}
	return nil
}

func (m MsgFinalizePacketByPacketKey) MustDecodePacketKey() []byte {
	packetKey, err := commontypes.DecodePacketKey(m.PacketKey)
	if err != nil {
		panic(fmt.Errorf("failed to decode base64 packet key: %w", err))
	}
	return packetKey
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	err = m.Params.ValidateBasic()
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidRequest,
			errorsmod.Wrapf(err, "failed to validate params"),
		)
	}

	return nil
}
