package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) SetCanonicalClient(goCtx context.Context, msg *types.MsgSetCanonicalClient) (*types.MsgSetCanonicalClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.TrySetCanonicalClient(ctx, msg.ClientId); err != nil {
		return nil, err
	}
	return &types.MsgSetCanonicalClientResponse{}, nil
}
