package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
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
			packetId commontypes.PacketUID
		)
		switch msg := m.(type) {
		case *channeltypes.MsgRecvPacket:
			height = msg.ProofHeight
			packetId = commontypes.NewPacketUID(
				commontypes.RollappPacket_ON_RECV,
				msg.Packet.DestinationPort,
				msg.Packet.DestinationChannel,
				msg.Packet.Sequence,
			)

		case *channeltypes.MsgAcknowledgement:
			height = msg.ProofHeight
			packetId = commontypes.NewPacketUID(
				commontypes.RollappPacket_ON_ACK,
				msg.Packet.SourcePort,
				msg.Packet.SourceChannel,
				msg.Packet.Sequence,
			)

		case *channeltypes.MsgTimeout:
			height = msg.ProofHeight
			packetId = commontypes.NewPacketUID(
				commontypes.RollappPacket_ON_TIMEOUT,
				msg.Packet.SourcePort,
				msg.Packet.SourceChannel,
				msg.Packet.Sequence,
			)
		default:
			continue
		}

		ctx = delayedacktypes.NewIBCProofContext(ctx, packetId, height)
	}
	return next(ctx, tx, simulate)
}
