package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// DelayedAckIBCModule represents an IBC module for x/delayedack. We need to trigger _next_ IBC middlewares
// after x/delayedack in order to process packet finalization requests.
type DelayedAckIBCModule interface {
	NextIBCMiddleware() porttypes.IBCModule
}

var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	k   Keeper
	ibc DelayedAckIBCModule // x/delayedack IBC module
}

func NewMsgServer(k Keeper, ibc DelayedAckIBCModule) MsgServer {
	return MsgServer{k: k, ibc: ibc}
}

func (m MsgServer) FinalizePacket(goCtx context.Context, msg *types.MsgFinalizePacket) (*types.MsgFinalizePacketResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err = m.k.FinalizeRollappPacket(ctx, m.ibc.NextIBCMiddleware(), string(msg.PendingPacketKey()))
	if err != nil {
		return nil, err
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventFinalizePacket{
		Sender:            msg.Sender,
		RollappId:         msg.RollappId,
		PacketProofHeight: msg.PacketProofHeight,
		PacketType:        msg.PacketType,
		PacketSrcChannel:  msg.PacketSrcChannel,
		PacketSequence:    msg.PacketSequence,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFinalizePacketResponse{}, nil
}

func (m MsgServer) FinalizePacketByPacketKey(goCtx context.Context, msg *types.MsgFinalizePacketByPacketKey) (*types.MsgFinalizePacketByPacketKeyResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	packetKey := string(msg.MustDecodePacketKey())
	packet, err := m.k.FinalizeRollappPacket(ctx, m.ibc.NextIBCMiddleware(), packetKey)
	if err != nil {
		return nil, err
	}

	var (
		sourceChannel string
		sequence      uint64
	)

	if packet.Packet != nil {
		sourceChannel = packet.Packet.SourceChannel
		sequence = packet.Packet.Sequence
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventFinalizePacket{
		Sender:            msg.Sender,
		RollappId:         packet.RollappId,
		PacketProofHeight: packet.ProofHeight,
		PacketType:        packet.Type,
		PacketSrcChannel:  sourceChannel,
		PacketSequence:    sequence,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFinalizePacketByPacketKeyResponse{}, nil
}

func (m MsgServer) FinalizeRollappPacketsByReceiver(goCtx context.Context, msg *types.MsgFinalizeRollappPacketsByReceiver) (*types.MsgFinalizeRollappPacketsByReceiverResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	result, err := m.k.FinalizeRollappPacketsByReceiver(ctx, m.ibc.NextIBCMiddleware(), msg.RollappId, msg.Receiver)
	if err != nil {
		return nil, err
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventFinalizeRollappPacketsByReceiver{
		Sender:       msg.Sender,
		RollappId:    msg.RollappId,
		Receiver:     msg.Receiver,
		Height:       result.LatestFinalizedHeight,
		FinalizedNum: result.FinalizedNum,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFinalizeRollappPacketsByReceiverResponse{}, nil
}
