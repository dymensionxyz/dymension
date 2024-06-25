package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/denommetadata module sentinel errors
var (
	ErrUnknownRequest = errorsmod.Register(ModuleName, 1002, "unknown request")
)
