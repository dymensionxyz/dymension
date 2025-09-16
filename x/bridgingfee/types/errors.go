package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/bridgingfee module sentinel errors
var (
	ErrInvalidTokenId   = errorsmod.Register(ModuleName, 1100, "invalid token id")
	ErrInvalidFee       = errorsmod.Register(ModuleName, 1101, "invalid fee")
	ErrDuplicateHookId  = errorsmod.Register(ModuleName, 1102, "duplicate hook id")
	ErrEmptyHookIds     = errorsmod.Register(ModuleName, 1103, "empty hook ids")
	ErrHookNotFound     = errorsmod.Register(ModuleName, 1104, "hook not found")
	ErrNotOwner         = errorsmod.Register(ModuleName, 1105, "not the owner")
	ErrInvalidOwner     = errorsmod.Register(ModuleName, 1106, "invalid owner address")
)