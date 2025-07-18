package forward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// f returns <is a forward operation, error>. Thus enabling wrapping non-forward operations (parsing and so on).
func (k Forward) executeWithErrEvent(ctx sdk.Context, f func() (bool, error)) {
	isForward, err := f()
	evt := &types.EventForward{
		Ok:           err == nil,
		WasForwarded: isForward,
	}
	if err != nil {
		evt.Err = err.Error()
	}
	emitErr := uevent.EmitTypedEvent(ctx, evt)
	if emitErr != nil {
		k.Logger(ctx).Error("Emit forward event", "error", emitErr)
	}
}
