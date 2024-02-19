package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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

// Bond defines a method for adding coins to sequencer's bond
func (k msgServer) Bond(context.Context, *types.MsgBond) (*types.MsgBondResponse, error) {
	panic("implement me")
}
