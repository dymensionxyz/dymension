package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// This decorator enforces that once we have canonical channel set for a rollapp, no more channels can be opened.
func (i IBCMessagesDecorator) HandleMsgChannelOpenAck(ctx sdk.Context, msg *ibcchanneltypes.MsgChannelOpenAck) error {
	if msg.PortId != ibctransfertypes.PortID { // We only care about transfer channels to mark them as canonical
		return nil
	}
	// Check if this channel is being opened on a known canonical client
	_, connection, err := i.ibcChannelKeeper.GetChannelConnection(ctx, msg.PortId, msg.ChannelId)
	if err != nil {
		return err
	}
	rollappID, found := i.k.GetRollappForClientID(ctx, connection.GetClientID())
	if !found {
		// channel is for non rollapp
		return nil
	}
	// Check if canon channel already exists for rollapp, if yes, return err
	rollapp, found := i.raK.GetRollapp(ctx, rollappID)
	if !found {
		return errorsmod.Wrap(gerrc.ErrInternal, "rollapp not found")
	}
	if rollapp.ChannelId != "" {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "canonical channel already exists for the rollapp")
	}

	return nil
}
