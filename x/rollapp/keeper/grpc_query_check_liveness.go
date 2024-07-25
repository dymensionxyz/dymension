package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) Liveness(goCtx context.Context, req *types.QueryGetLivenessRequest) (*types.QueryGetLivenessResponse, error) {
}
