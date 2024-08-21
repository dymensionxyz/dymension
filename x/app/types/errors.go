package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	ErrAppExists             = errorsmod.Register(ModuleName, 1000, "app already exists")
	ErrInvalidCreatorAddress = errorsmod.Register(ModuleName, 1001, "invalid creator address")
	ErrFeePayment            = errorsmod.Register(ModuleName, 1002, "app creation fee payment error")
	ErrNotFound              = errorsmod.Register(ModuleName, 1003, "not found")
	ErrRollappNotFound       = errorsmod.Register(ModuleName, 1004, "rollapp not found")
	ErrInvalidURL            = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid url")
	ErrInvalidName           = errorsmod.Wrap(gerrc.ErrInvalidArgument, "name")
	ErrInvalidRollappId      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "rollapp id")
	ErrInvalidImage          = errorsmod.Wrap(gerrc.ErrInvalidArgument, "image path")
	ErrInvalidDescription    = errorsmod.Wrap(gerrc.ErrInvalidArgument, "description")
	ErrUnauthorizedSigner    = errorsmod.Wrap(gerrc.ErrPermissionDenied, "unauthorized signer")
)
