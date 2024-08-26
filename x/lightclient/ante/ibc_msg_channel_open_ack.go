package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

func (i IBCMessagesDecorator) HandleMsgChannelOpenAck(ctx sdk.Context, msg *ibcchanneltypes.MsgChannelOpenAck) error {
	if msg.PortId != ibctransfertypes.PortID { // We only care about transfer channels to mark them as canonical
		return nil
	}
	// Check if this channel is being opened on a known canonical client
	_, connection, err := i.ibcChannelKeeper.GetChannelConnection(ctx, msg.PortId, msg.ChannelId)
	if err != nil {
		return err
	}
	rollappID, found := i.lightClientKeeper.GetRollappForClientID(ctx, connection.GetClientID())
	if !found {
		return nil
	}
	// Check if canon channel already exists for rollapp, if yes, return err
	rollapp, found := i.rollappKeeper.GetRollapp(ctx, rollappID)
	if !found {
		return nil
	}
	if rollapp.ChannelId != "" {
		return errorsmod.Wrap(ibcchanneltypes.ErrChannelExists, "cannot create a new channel when a canonical channel already exists for the rollapp")
	}
	// Set this channel as the canonical channel for the rollapp
	rollapp.ChannelId = msg.ChannelId
	i.rollappKeeper.SetRollapp(ctx, rollapp)

	return nil
}
