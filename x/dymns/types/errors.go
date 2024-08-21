package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var ErrBadDymNameAddress = errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name address is invalid")
