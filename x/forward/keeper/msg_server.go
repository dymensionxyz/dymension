package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) Foo(ctx context.Context, msg *types.MsgFoo) (*types.MsgFooResponse, error) {
	return &types.MsgFooResponse{}, nil
}

// assumes funds are already in module account
func (k Keeper) DemoHLToIBC(_ context.Context, _ *types.MsgDemoHLToIBC) (*types.MsgDemoHLToIBCResponse, error) {
	panic("not implemented") // TODO: Implement
}
