package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
)

var (
	ErrValidationFailed      = errorsmod.Register(ModuleName, 1, "validation failed")
	ErrInvalidOwner          = errorsmod.Register(ModuleName, 2, "invalid owner")
	ErrInvalidState          = errorsmod.Register(ModuleName, 3, "invalid state")
	ErrDymNameNotFound       = errorsmod.Register(ModuleName, 4, "Dym-Name could not be found")
	ErrSellOrderNotFound     = errorsmod.Register(ModuleName, 5, "sell order could not be found")
	ErrGracePeriod           = errorsmod.Register(ModuleName, 6, "expired Dym-Name still in grace period")
	ErrBadDymNameAddress     = errorsmod.Register(ModuleName, 7, "bad format Dym-Name address")
	ErrDymNameTooLong        = errorsmod.Register(ModuleName, 8, fmt.Sprintf("Dym-Name is too long, maximum %d characters", MaxDymNameLength))
	ErrUnAcknowledgedPayment = errorsmod.Register(ModuleName, 9, "Un-acknowledged payment")
	ErrOfferToBuyNotFound    = errorsmod.Register(ModuleName, 10, "offer to buy could not be found")
)
