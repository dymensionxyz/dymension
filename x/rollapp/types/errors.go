package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/rollapp module sentinel errors
var (
	ErrRollappExists = sdkerrors.Register(ModuleName, 1000, "rollapp already exist for this rollapp-id; must use new rollapp-id")
)
