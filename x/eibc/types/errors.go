package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/eibc module sentinel errors
var (
	ErrInvalidOrderID               = errorsmod.Register(ModuleName, 3, "Invalid order ID")
	ErrDemandOrderDoesNotExist      = errorsmod.Register(ModuleName, 5, "Demand order does not exist")
	ErrDemandOrderInactive          = errorsmod.Register(ModuleName, 6, "Demand order inactive")
	ErrFulfillerAddressDoesNotExist = errorsmod.Register(ModuleName, 7, "Fulfiller address does not exist")
	ErrInvalidRecipientAddress      = errorsmod.Register(ModuleName, 8, "Invalid recipient address")
	ErrBlockedAddress               = errorsmod.Register(ModuleName, 9, "Can't purchase demand order for recipient with blocked address")
	ErrDemandAlreadyFulfilled       = errorsmod.Register(ModuleName, 10, "Demand order already fulfilled")
	ErrNegativeTimeoutFee           = errorsmod.Register(ModuleName, 11, "Timeout fee must be greater than or equal to 0")
	ErrTooMuchTimeoutFee            = errorsmod.Register(ModuleName, 12, "Timeout fee must be less than or equal to the total amount")
	ErrNegativeErrAckFee            = errorsmod.Register(ModuleName, 13, "Error acknowledgement fee must be greater than or equal to 0")
	ErrTooMuchErrAckFee             = errorsmod.Register(ModuleName, 14, "Error acknowledgement fee must be less than or equal to the total amount")
	ErrMinFeeNotMet                 = errorsmod.Register(ModuleName, 15, "Minimum fee not met")
)
