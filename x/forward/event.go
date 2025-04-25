package forward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

func (k Forward) executeWithErrEvent(ctx sdk.Context, f func() error) {
	err := f()
	evt := &types.EventForward{
		Ok: err == nil,
	}
	if err != nil {
		evt.Err = err.Error()
	}
	emitErr := uevent.EmitTypedEvent(ctx, evt)
	if emitErr != nil {
		k.Logger(ctx).Error("Emit forward event", "error", emitErr)
	}
}
