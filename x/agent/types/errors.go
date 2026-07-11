package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrAgentExists            = errorsmod.Register(ModuleName, 2, "agent already exists")
	ErrAgentNotFound          = errorsmod.Register(ModuleName, 3, "agent not found")
	ErrRegistrationFeePayment = errorsmod.Register(ModuleName, 4, "agent registration fee payment error")
	ErrUnauthorized           = errorsmod.Register(ModuleName, 5, "unauthorized")
	ErrInvalidPolicy          = errorsmod.Register(ModuleName, 6, "invalid policy")
	ErrActionNotFound         = errorsmod.Register(ModuleName, 7, "action log entry not found")
	ErrSelfFeedback           = errorsmod.Register(ModuleName, 8, "agent owner cannot rate own agent")
	ErrInvalidEvidence        = errorsmod.Register(ModuleName, 9, "evidence must reference an existing action log entry")
	ErrFeedbackNotFound       = errorsmod.Register(ModuleName, 10, "feedback not found")
	ErrInvalidScore           = errorsmod.Register(ModuleName, 11, "invalid feedback score")
	ErrInvalidTag             = errorsmod.Register(ModuleName, 12, "invalid feedback tag")
	ErrFeedbackFeePayment     = errorsmod.Register(ModuleName, 13, "feedback fee payment error")
)
