package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrAgentExists            = errorsmod.Register(ModuleName, 2, "agent already exists")
	ErrAgentNotFound          = errorsmod.Register(ModuleName, 3, "agent not found")
	ErrRegistrationFeePayment = errorsmod.Register(ModuleName, 4, "agent registration fee payment error")
	ErrUnauthorized           = errorsmod.Register(ModuleName, 5, "unauthorized")
	ErrInvalidPolicy          = errorsmod.Register(ModuleName, 6, "invalid policy")
)
