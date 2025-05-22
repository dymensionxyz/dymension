package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
)

var _ sdk.AnteDecorator = &InnerDecorator{}

// InnerCallback is a function that will be called for each leaf message in a transaction
// The interface is similar to sdk.AnteDecorator.AnteHandle, but for inner messages
type InnerCallback func(ctx sdk.Context, msg sdk.Msg, simulate bool, depth int) (sdk.Context, error)

// InnerDecorator applies a list of InnerDecoratorFuncs to all leaf messages in a transaction
// (recursively traversing wrappers like MsgExec, MsgSubmitProposal, etc.)
type InnerDecorator struct {
	callbacks []InnerCallback
}

// NewInnerDecorator constructs an InnerDecorator with the given list of decorator functions
func NewInnerDecorator(callbacks ...InnerCallback) *InnerDecorator {
	return &InnerDecorator{callbacks: callbacks}
}

// AnteHandle implements sdk.AnteDecorator
func (id InnerDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	msgs := tx.GetMsgs()
	ctx, err := id.handleMsgs(ctx, msgs, simulate, 0)
	if err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}

// handleMsgs recursively traverses messages, applying decorators to all leaf messages
func (id InnerDecorator) handleMsgs(ctx sdk.Context, msgs []sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
	var err error
	for _, msg := range msgs {
		ctx, err = id.handleMsg(ctx, msg, simulate, depth)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (id InnerDecorator) handleMsg(ctx sdk.Context, msg sdk.Msg, simulate bool, depth int) (sdk.Context, error) {
	if depth >= MaxInnerDepth {
		return ctx, fmt.Errorf("found more nested msgs than permitted. Limit is : %d", MaxInnerDepth)
	}

	var (
		err   error
		inner []sdk.Msg
	)

	switch m := msg.(type) {
	case *authz.MsgExec:
		inner, err = m.GetMessages()
	case *govtypesv1.MsgSubmitProposal:
		inner, err = m.GetMsgs()
	case *group.MsgSubmitProposal:
		inner, err = m.GetMsgs()
	default:
	}

	if err != nil {
		return ctx, err
	}

	if len(inner) > 0 {
		// Not a leaf, recurse
		ctx, err = id.handleMsgs(ctx, inner, simulate, depth+1)
		if err != nil {
			return ctx, err
		}

		return ctx, nil
	}

	// Leaf message: apply all callbacks
	for _, cb := range id.callbacks {
		ctx, err = cb(ctx, msg, simulate, depth)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
