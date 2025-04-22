package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) executeWithErrEvent(ctx sdk.Context, f func() error) {
	err := f()
	evt := &types.EventForward{
		Ok: err == nil,
	}
	if err != nil {
		evt.Err = err.Error()
	}
	emitErr := ctx.EventManager().EmitTypedEvent(evt)
	if emitErr != nil {
		k.Logger(ctx).Error("Emit forward event", "error", emitErr)
	}
}
