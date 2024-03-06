package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

type IBCProofHeightDecorator struct {
}

func NewIBCProofHeightDecorator() IBCProofHeightDecorator {
	return IBCProofHeightDecorator{}
}

func (rrd IBCProofHeightDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// do not run check on CheckTx or simulate
	deliverTx := !(ctx.IsCheckTx() || ctx.IsReCheckTx())
	if simulate || !deliverTx {
		return next(ctx, tx, simulate)
	}

	for _, m := range tx.GetMsgs() {
		var (
			height   clienttypes.Height
			sequence uint64
		)
		switch msg := m.(type) {
		case *channeltypes.MsgRecvPacket:
			height = msg.ProofHeight
			sequence = msg.Packet.Sequence
		case *channeltypes.MsgAcknowledgement:
			height = msg.ProofHeight
			sequence = msg.Packet.Sequence
		case *channeltypes.MsgTimeout:
			height = msg.ProofHeight
			sequence = msg.Packet.Sequence
		default:
			continue
		}
		ctx = delayedacktypes.NewIBCProofContext(ctx, sequence, height)
	}

	return next(ctx, tx, simulate)
}
