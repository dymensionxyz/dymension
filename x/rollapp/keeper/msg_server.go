package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	v2types "github.com/dymensionxyz/dymension/v3/x/rollapp/types/v2"
)

type msgServer struct {
	Keeper
}

type msgServerV2 struct {
	Keeper
}

var _ types.MsgServer = msgServer{}
var _ v2types.MsgServer = msgServerV2{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// NewMsgServerV2Impl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerV2Impl(keeper Keeper) v2types.MsgServer {
	return &msgServerV2{Keeper: keeper}
}
