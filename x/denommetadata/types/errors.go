package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// The following regiisters various lockdrop errors.
var (
	ErrDenomBadFormat = sdkerrors.Register(ModuleName, 1, "input denom format is not correct")
)
