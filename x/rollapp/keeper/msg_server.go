package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) TriggerGenesisEvent(context.Context, *types.MsgRollappGenesisEvent) (*types.MsgRollappGenesisEventResponse, error) {
	// TODO: return not implemented or deprecated error
	return &types.MsgRollappGenesisEventResponse{}, nil
}
