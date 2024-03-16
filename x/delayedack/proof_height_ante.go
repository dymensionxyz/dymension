package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

type IBCProofHeightDecorator struct{}

func NewIBCProofHeightDecorator() IBCProofHeightDecorator {
	return IBCProofHeightDecorator{}
}

func (rrd IBCProofHeightDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, m := range tx.GetMsgs() {
		var (
			height   clienttypes.Height
			packetId channeltypes.PacketId
		)
		switch msg := m.(type) {
		case *channeltypes.MsgRecvPacket:
			height = msg.ProofHeight
			packetId = channeltypes.NewPacketID(msg.Packet.GetDestPort(), msg.Packet.GetDestChannel(), msg.Packet.GetSequence())

		case *channeltypes.MsgAcknowledgement:
			height = msg.ProofHeight
			packetId = channeltypes.NewPacketID(msg.Packet.GetDestPort(), msg.Packet.GetDestChannel(), msg.Packet.GetSequence())

		case *channeltypes.MsgTimeout:
			height = msg.ProofHeight
			packetId = channeltypes.NewPacketID(msg.Packet.GetDestPort(), msg.Packet.GetDestChannel(), msg.Packet.GetSequence())
		default:
			continue
		}

		ctx = delayedacktypes.NewIBCProofContext(ctx, packetId, height)
	}

	return next(ctx, tx, simulate)
}
