package uevent

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// NewErrorAcknowledgement wraps the standard error acknowledgement with an error event.
func NewErrorAcknowledgement(ctx sdk.Context, err error) types.Acknowledgement {
	EmitErrorEvent(ctx, err)
	return types.NewErrorAcknowledgement(err)
}

const (
	EventTypeError = "error"

	AttributeKeyError  = "message"
	AttributeKeyHeight = "height"
)

// EmitErrorEvent emits an error event.
// Example of an error event:
/*
- attributes:
  - key: ibccallbackerror-height
	value: "439"
  - key: ibccallbackerror-message
	value: 'transfer genesis: get valid transfer: get rollapp id: rollapp canonical
	  channel mapping has not been set: rollappeve_1235-1: unmet precondition'
  type: ibccallbackerror-error
*/
// Example query to get the error event:
// dymd q txs --events ibccallbackerror-error.ibccallbackerror-height=439
func EmitErrorEvent(ctx sdk.Context, err error) {
	if err == nil {
		return
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeError,
			sdk.NewAttribute(AttributeKeyHeight, fmt.Sprint(ctx.BlockHeight())),
			sdk.NewAttribute(AttributeKeyError, err.Error()),
		),
	)
}
