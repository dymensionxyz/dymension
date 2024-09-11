package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// delayedAckIBCModule represents an IBC module for x/delayedack. We need to trigger _next_ IBC middlewares
// after x/delayedack in order to process packet finalization requests.
type delayedAckIBCModule interface {
	NextIBCMiddleware() porttypes.IBCModule
}

var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	k   Keeper
	ibc delayedAckIBCModule // x/delayedack IBC module
}

func NewMsgServer(k Keeper, ibc delayedAckIBCModule) MsgServer {
	return MsgServer{k: k, ibc: ibc}
}

func (m MsgServer) FinalizePacket(goCtx context.Context, msg *types.MsgFinalizePacket) (*types.MsgFinalizePacketResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	err = m.k.FinalizeRollappPacket(ctx, m.ibc.NextIBCMiddleware(), msg.RollappId, string(msg.PendingPacketKey()))
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

func (m MsgServer) FinalizePacketsUntilHeight(goCtx context.Context, msg *types.MsgFinalizePacketsUntilHeight) (*types.MsgFinalizePacketsUntilHeightResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	finalized, err := m.k.FinalizeRollappPackets(ctx, m.ibc.NextIBCMiddleware(), msg.RollappId, msg.Height)
	if err != nil {
		return nil, err
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventFinalizePacketsUntilHeight{
		Sender:       msg.Sender,
		RollappId:    msg.RollappId,
		Height:       msg.Height,
		FinalizedNum: uint64(finalized),
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFinalizePacketsUntilHeightResponse{}, nil
}
