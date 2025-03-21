package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) Foo(ctx context.Context, msg *types.MsgFoo) (*types.MsgFooResponse, error) {
	return &types.MsgFooResponse{}, nil
}
