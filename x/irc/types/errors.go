package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/irc module sentinel errors
var (
	ErrInvalidMsgType = sdkerrors.Register(ModuleName, 1100, "invalid msg type")
)
