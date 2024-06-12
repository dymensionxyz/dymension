package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/denommetadata module sentinel errors
var (
	ErrDenomDoesNotExist = errorsmod.Register(ModuleName, 1001, "unable to find denom metadata registered")
	ErrUnknownRequest    = errorsmod.Register(ModuleName, 1002, "unknown request")
)
