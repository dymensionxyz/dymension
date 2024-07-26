package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RegisterRollapp(ctx, msg.GetRollapp()); err != nil {
		return nil, err
	}

	return &types.MsgCreateRollappResponse{}, nil
}
