package types

// DONTCOVER

import (
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/eibc module sentinel errors
var (
	ErrInvalidOrderID               = gerrc.ErrInvalidArgument.Wrap("order id")
	ErrDemandOrderAlreadyExist      = gerrc.ErrAlreadyExists.Wrap("demand order")
	ErrDemandOrderDoesNotExist      = gerrc.ErrNotFound.Wrap("demand order")
	ErrDemandOrderInactive          = gerrc.ErrFailedPrecondition.Wrap("demand order inactive")
	ErrFulfillerAddressDoesNotExist = gerrc.ErrNotFound.Wrap("fulfiller address")
	ErrInvalidRecipientAddress      = gerrc.ErrInvalidArgument.Wrap("recipient address")
	ErrBlockedAddress               = gerrc.ErrFailedPrecondition.Wrap("can't purchase demand order for recipient with blocked address")
	ErrDemandAlreadyFulfilled       = gerrc.ErrFailedPrecondition.Wrap("demand order already fulfilled")
	ErrFeeTooHigh                   = gerrc.ErrOutOfRange.Wrap("fee must be less than or equal to the total amount")
	ErrExpectedFeeNotMet            = gerrc.ErrOutOfRange.Wrap("expected fee not met")
	ErrNegativeFee                  = gerrc.ErrOutOfRange.Wrap("fee must be greater than or equal to 0")
	ErrMultipleDenoms               = gerrc.ErrInvalidArgument.Wrap("multiple denoms")
	ErrEmptyPrice                   = gerrc.ErrOutOfRange.Wrap("price must be greater than 0")
)
