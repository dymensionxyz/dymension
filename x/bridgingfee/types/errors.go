package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidFee   = errorsmod.Register(ModuleName, 1, "invalid fee")
	ErrInvalidOwner = errorsmod.Register(ModuleName, 2, "invalid owner address")
)
