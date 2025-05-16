package types

// DONTCOVER

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/eibc module sentinel errors
var (
	ErrInvalidRecipientAddress = gerrc.ErrInvalidArgument.Wrap("recipient address")
	ErrInvalidCreationHeight   = gerrc.ErrInvalidArgument.Wrap("creation height")
	ErrMultipleDenoms          = gerrc.ErrInvalidArgument.Wrap("multiple denoms not allowed")
	ErrEmptyPrice              = gerrc.ErrInvalidArgument.Wrap("price must be greater than 0")
	ErrDemandAlreadyFulfilled  = gerrc.ErrFailedPrecondition.Wrap("demand order already fulfilled")
	ErrDemandOrderInactive     = gerrc.ErrInvalidArgument.Wrap("demand order inactive")
)
