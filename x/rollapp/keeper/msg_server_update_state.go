package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

func (k msgServer) UpdateState(goCtx context.Context, msg *types.MsgUpdateState) (*types.MsgUpdateStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// call the before-update-state hook
	err := k.BeforeUpdateStateRecoverable(ctx, msg.Creator, msg.RollappId)

	return &types.MsgUpdateStateResponse{}, err
}
