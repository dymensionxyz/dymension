package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParamStore defines the interface the parameter store used by the BaseApp must
// fulfill.
type IRCRequestI interface {
	// GetReqId returns the request id
	GetReqId() uint64
	// GetMessageType returns the message type
	GetMessageType() MsgType
	// GetMsg returns the wrapped message
	GetMsg() sdk.Msg
	// Hnadler handling the request on dequeue
	Handler() error
}
