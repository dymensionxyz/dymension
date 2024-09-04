package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/lockup module sentinel errors.
var (
	ErrNotLockOwner   = errorsmod.Wrap(gerrc.ErrInvalidArgument, "msg sender is not the owner of specified lock")
	ErrLockupNotFound = errorsmod.Wrap(gerrc.ErrNotFound, "lockup not found")
)
