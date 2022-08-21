package keeper

import (
	"context"

    "github.com/dymensionxyz/dymension/x/rollapp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) UpdateState(goCtx context.Context,  msg *types.MsgUpdateState) (*types.MsgUpdateStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgUpdateStateResponse{}, nil
}
