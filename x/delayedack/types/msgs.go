package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
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

func (m MsgFinalizePacket) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{signer}
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

func (m MsgFinalizePacketsUntilHeight) ValidateBasic() error {
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
	return nil
}

func (m MsgFinalizePacketsUntilHeight) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{signer}
}

func (m MsgFinalizeRollappPacketsByReceiver) ValidateBasic() error {
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
	if len(m.Receiver) == 0 {
		return gerrc.ErrInvalidArgument.Wrap("receiver must be non-empty")
	}
	return nil
}

func (m MsgFinalizeRollappPacketsByReceiver) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{signer}
}
