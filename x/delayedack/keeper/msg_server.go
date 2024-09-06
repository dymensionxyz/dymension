package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/delayedack"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

type MsgServer struct {
	k   Keeper
	ibc delayedack.IBCMiddleware // x/delayedack IBC module
}

func NewMsgServer(k Keeper, ibc delayedack.IBCMiddleware) MsgServer {
	return MsgServer{
		k:   k,
		ibc: ibc,
	}
}

func (m MsgServer) FinalizePacket(goCtx context.Context, msg *types.MsgFinalizePacket) (*types.MsgFinalizePacketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	latestHeight, found := m.k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, msg.RollappId)
	if !found {
		return nil, fmt.Errorf("latest finalized state index for rollapp %s is not found", msg.RollappId)
	}

	m.k.GetRollappPacket()

	// get packet key for the storage get
	packetKey := msg.PacketKey()
	packet := m.keeper.GetPacket(packetKey)

	// check finalization height of the rollapp is higher than the packet proof.
	if packet.ProofHeight.Later(latestHeight) {
		return fmt.Errorf("packet height did not finalize yet")
	}
	// apply the same logic as before
	m.k.finalizeRollappPacket(ctx, m.ibc.NextIBCMiddleware(), packet)
}
