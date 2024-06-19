package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/denommetadata module sentinel errors
var (
	ErrDenomAlreadyExists = errorsmod.Register(ModuleName, 1000, "denom metadata is already registered")
	ErrDenomDoesNotExist  = errorsmod.Register(ModuleName, 1001, "unable to find denom metadata registered")
	ErrUnknownRequest     = errorsmod.Register(ModuleName, 1002, "unknown request")
	ErrRollappNotFound    = errorsmod.Register(ModuleName, 1003, "rollapp not found")
)
