package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrTimestampNotFound = sdkerrors.Register(ModuleName, 2, "block descriptors do not contain block timestamp")
)
