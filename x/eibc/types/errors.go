package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/eibc module sentinel errors
var (
	ErrInvalidOrderID              = errorsmod.Register(ModuleName, 3, "invalid order ID")
	ErrDemandOrderAlreadyExist     = errorsmod.Register(ModuleName, 4, "demand order already exists")
	ErrDemandOrderDoesNotExist     = errorsmod.Register(ModuleName, 5, "demand order does not exist")
	ErrDemandOrderInactive         = errorsmod.Register(ModuleName, 6, "demand order inactive")
	ErrAccountDoesNotExist         = gerrc.ErrNotFound.Wrap("account")
	ErrInvalidRecipientAddress     = errorsmod.Register(ModuleName, 8, "invalid recipient address")
	ErrBlockedAddress              = errorsmod.Register(ModuleName, 9, "cant purchase demand order for recipient with blocked address")
	ErrDemandAlreadyFulfilled      = errorsmod.Register(ModuleName, 10, "demand order already fulfilled")
	ErrFeeTooHigh                  = errorsmod.Register(ModuleName, 11, "fee must be less than or equal to the total amount")
	ErrExpectedFeeNotMet           = errorsmod.Register(ModuleName, 12, "expected fee not met")
	ErrNegativeFee                 = errorsmod.Register(ModuleName, 13, "fee must be greater than or equal to 0")
	ErrMultipleDenoms              = errorsmod.Register(ModuleName, 15, "multiple denoms not allowed")
	ErrEmptyPrice                  = errorsmod.Register(ModuleName, 16, "price must be greater than 0")
	ErrRollappStateInfoNotFound    = errorsmod.Register(ModuleName, 19, "rollapp state info not found")
	ErrOrderNotSettlementValidated = errorsmod.Register(ModuleName, 20, "demand order not settlement validated")
	ErrRollappIdMismatch           = errorsmod.Register(ModuleName, 21, "rollapp ID mismatch")
	ErrPriceMismatch               = errorsmod.Register(ModuleName, 22, "price mismatch")
	ErrInvalidCreationHeight       = errorsmod.Register(ModuleName, 23, "invalid creation height")
)
