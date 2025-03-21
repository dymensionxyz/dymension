package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Wiz(goCtx context.Context, req *types.WizRequest) (*types.WizResponse, error) {
	return &types.WizResponse{}, nil
}
