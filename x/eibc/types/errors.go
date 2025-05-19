package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/eibc module sentinel errors
var (
	ErrInvalidRecipientAddress     = gerrc.ErrInvalidArgument.Wrap("recipient address")
	ErrInvalidCreationHeight       = gerrc.ErrInvalidArgument.Wrap("creation height")
	ErrMultipleDenoms              = gerrc.ErrInvalidArgument.Wrap("multiple denoms not allowed")
	ErrEmptyPrice                  = gerrc.ErrInvalidArgument.Wrap("price must be greater than 0")
	ErrDemandAlreadyFulfilled      = gerrc.ErrFailedPrecondition.Wrap("demand order already fulfilled")
	ErrDemandOrderInactive         = gerrc.ErrInvalidArgument.Wrap("demand order inactive")
	ErrInvalidOrderID              = errorsmod.Register(ModuleName, 3, "invalid order ID")
	ErrDemandOrderAlreadyExist     = errorsmod.Register(ModuleName, 4, "demand order already exists")
	ErrDemandOrderDoesNotExist     = errorsmod.Register(ModuleName, 5, "demand order does not exist")
	ErrAccountDoesNotExist         = gerrc.ErrNotFound.Wrap("account")
	ErrBlockedAddress              = errorsmod.Register(ModuleName, 9, "cant purchase demand order for recipient with blocked address")
	ErrFeeTooHigh                  = errorsmod.Register(ModuleName, 11, "fee must be less than or equal to the total amount")
	ErrExpectedFeeNotMet           = errorsmod.Register(ModuleName, 12, "expected fee not met")
	ErrNegativeFee                 = errorsmod.Register(ModuleName, 13, "fee must be greater than or equal to 0")
	ErrRollappStateInfoNotFound    = errorsmod.Register(ModuleName, 19, "rollapp state info not found")
	ErrOrderNotSettlementValidated = errorsmod.Register(ModuleName, 20, "demand order not settlement validated")
	ErrRollappIdMismatch           = errorsmod.Register(ModuleName, 21, "rollapp ID mismatch")
	ErrPriceMismatch               = errorsmod.Register(ModuleName, 22, "price mismatch")
)
