package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

type msgServer struct {
	*Keeper
}

func (m msgServer) CreateSequencer(context.Context, *types.MsgFoo) (*types.MsgFooResponse, error) {
	panic("unimplemented")
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
