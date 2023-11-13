package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/lockdrop keeper providing gRPC
// method handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// DistrInfo returns gauges receiving pool rewards and their respective weights.
func (q Querier) DistrInfo(ctx context.Context, _ *types.QueryDistrInfoRequest) (*types.QueryDistrInfoResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryDistrInfoResponse{DistrInfo: q.Keeper.GetDistrInfo(sdkCtx)}, nil
}

// Params return lockdrop module params.
func (q Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(sdkCtx)}, nil
}
