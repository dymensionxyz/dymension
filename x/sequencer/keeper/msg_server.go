package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) sequencertypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ sequencertypes.MsgServer = msgServer{}

// PunishSequencer implements the Msg/PunishSequencer RPC method
func (k msgServer) PunishSequencer(goCtx context.Context, msg *sequencertypes.MsgPunishSequencer) (*sequencertypes.MsgPunishSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rewardee, _ := sdk.AccAddressFromBech32(msg.Rewardee)

	if err := k.Keeper.PunishSequencer(ctx, msg.PunishSequencerAddress, &rewardee); err != nil {
		return nil, err
	}

	return &sequencertypes.MsgPunishSequencerResponse{}, nil
}
