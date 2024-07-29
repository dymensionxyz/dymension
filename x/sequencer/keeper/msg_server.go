package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type msgServer struct {
	Keeper
}

// DecreaseBond implements types.MsgServer.
func (k msgServer) DecreaseBond(context.Context, *types.MsgDecreaseBond) (*types.MsgDecreaseBondResponse, error) {
	panic("unimplemented")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
