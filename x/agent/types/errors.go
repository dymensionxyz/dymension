package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrAgentExists            = errorsmod.Register(ModuleName, 2, "agent already exists")
	ErrAgentNotFound          = errorsmod.Register(ModuleName, 3, "agent not found")
	ErrRegistrationFeePayment = errorsmod.Register(ModuleName, 4, "agent registration fee payment error")
	ErrUnauthorized           = errorsmod.Register(ModuleName, 5, "unauthorized")
	ErrInvalidPolicy          = errorsmod.Register(ModuleName, 6, "invalid policy")
	ErrActionNotFound         = errorsmod.Register(ModuleName, 7, "action log entry not found")
	ErrSpendingDisabled       = errorsmod.Register(ModuleName, 8, "agent spending is disabled")
	ErrSpendBudgetExceeded    = errorsmod.Register(ModuleName, 9, "spend budget exceeded for the current window")
	ErrInsufficientEscrow     = errorsmod.Register(ModuleName, 10, "insufficient escrow balance")
)
