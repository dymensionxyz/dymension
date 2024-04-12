package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/denommetadata module sentinel errors
var (
	ErrDenomAlreadyExists = sdkerrors.Register(ModuleName, 1000, "denom metadata is already registered")
	ErrDenomDoesNotExist  = sdkerrors.Register(ModuleName, 1001, "unable to find denom metadata registered")
)
