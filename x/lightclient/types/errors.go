package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrTimestampNotFound          = errorsmod.Wrap(gerrc.ErrNotFound, "block descriptors do not contain block timestamp")
	ErrNextBlockDescriptorMissing = errorsmod.Wrap(gerrc.ErrNotFound, "next block descriptor is missing")
)
