package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type IBCProofHeightDecorator struct{}

func NewIBCProofHeightDecorator() IBCProofHeightDecorator {
	return IBCProofHeightDecorator{}
}

func (rrd IBCProofHeightDecorator) InnerCallback(ctx sdk.Context, m sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
	var (
		height   clienttypes.Height
		packetId common.PacketUID
	)

	switch msg := m.(type) {
	case *channeltypes.MsgRecvPacket:
		height = msg.ProofHeight
		packetId = common.NewPacketUID(
			common.RollappPacket_ON_RECV,
			msg.Packet.DestinationPort,
			msg.Packet.DestinationChannel,
			msg.Packet.Sequence,
		)

	case *channeltypes.MsgAcknowledgement:
		height = msg.ProofHeight
		packetId = common.NewPacketUID(
			common.RollappPacket_ON_ACK,
			msg.Packet.SourcePort,
			msg.Packet.SourceChannel,
			msg.Packet.Sequence,
		)

	case *channeltypes.MsgTimeout:
		height = msg.ProofHeight
		packetId = common.NewPacketUID(
			common.RollappPacket_ON_TIMEOUT,
			msg.Packet.SourcePort,
			msg.Packet.SourceChannel,
			msg.Packet.Sequence,
		)
	default:
		return ctx, nil
	}

	ctx = CtxWithPacketProofHeight(ctx, packetId, height)
	return ctx, nil
}

func UnpackPacketProofHeight(
	ctx sdk.Context,
	packet channeltypes.Packet,
	packetType common.RollappPacket_Type,
) (uint64, error) {
	port, channel := common.PacketHubPortChan(packetType, packet)

	packetID := common.NewPacketUID(packetType, port, channel, packet.Sequence)
	height, ok := PacketProofHeightFromCtx(ctx, packetID)
	if !ok {
		return 0, errorsmod.Wrapf(gerrc.ErrInternal, "get proof height from context: packetID: %s", packetID)
	}
	return height.RevisionHeight, nil
}

const (
	// proofHeightCtxKey is a context key to pass the proof height from the msg to the IBC middleware
	proofHeightCtxKey = "ibc_proof_height"
)

func CtxWithPacketProofHeight(ctx sdk.Context, packetId common.PacketUID, height clienttypes.Height) sdk.Context {
	key := fmt.Sprintf("%s_%s", proofHeightCtxKey, packetId.String())
	return ctx.WithValue(key, height)
}

func PacketProofHeightFromCtx(ctx sdk.Context, packetId common.PacketUID) (clienttypes.Height, bool) {
	key := fmt.Sprintf("%s_%s", proofHeightCtxKey, packetId.String())
	u, ok := ctx.Value(key).(clienttypes.Height)
	return u, ok
}
