package forward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// only apply F in state machine if it succeeds
// emit an event (which has the error if there is one)
func (k Forward) executeAtomicWithErrEvent(ctx sdk.Context, f func(sdk.Context) (bool, error)) {
	forwardWasIntended, err := osmosF(ctx, f)
	evt := &types.EventForward{
		Ok:           err == nil,
		WasForwarded: forwardWasIntended,
	}
	if err != nil {
		evt.Err = err.Error()
	}
	emitErr := uevent.EmitTypedEvent(ctx, evt)
	if emitErr != nil {
		k.Logger(ctx).Error("Emit forward event", "error", emitErr)
	}
}

func osmosF(ctx sdk.Context, f func(sdk.Context) (bool, error)) (bool, error) {
	var forwardWasIntended bool // did the user intend to forward?
	err := osmoutils.ApplyFuncIfNoError(ctx, func(c sdk.Context) error {
		var err error
		forwardWasIntended, err = f(c)
		return err
	})
	return forwardWasIntended, err
}
