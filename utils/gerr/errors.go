package gerr

// See doc.go for info

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const GErrors = "gerr"

var (
	// uses canonical codes defined here https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto

	ErrCancelled          = errorsmod.RegisterWithGRPCCode(GErrors, 1, 1, "cancelled") // no obvious sdk mapping exists
	ErrUnknown            = errorsmod.RegisterWithGRPCCode(GErrors, 2, 2, "unknown")   // no obvious sdk mapping exists
	ErrInvalidArgument    = errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid argument")
	ErrDeadlineExceeded   = errorsmod.RegisterWithGRPCCode(GErrors, 4, 4, "deadline exceeded") // no obvious sdk mapping exists
	ErrNotFound           = sdkerrors.ErrNotFound
	ErrAlreadyExist       = errorsmod.RegisterWithGRPCCode(GErrors, 5, 6, "already exist") // no obvious sdk mapping exists
	ErrPermissionDenied   = errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "permission denied")
	ErrUnauthenticated    = errorsmod.Wrapf(sdkerrors.ErrWrongPassword, "unauthenticated")
	ErrResourceExhausted  = errorsmod.RegisterWithGRPCCode(GErrors, 8, 8, "resource exhausted")  // no obvious sdk mapping exists
	ErrFailedPrecondition = errorsmod.RegisterWithGRPCCode(GErrors, 9, 9, "failed precondition") // no obvious sdk mapping exists
	ErrAborted            = errorsmod.RegisterWithGRPCCode(GErrors, 10, 10, "aborted")           // no obvious sdk mapping exists
	ErrOutOfRange         = errorsmod.RegisterWithGRPCCode(GErrors, 11, 11, "out of range")      // no obvious sdk mapping exists
	ErrUnimplemented      = errorsmod.RegisterWithGRPCCode(GErrors, 12, 12, "unimplemented")     // no obvious sdk mapping exists
	ErrInternal           = errorsmod.RegisterWithGRPCCode(GErrors, 13, 13, "internal")          // no obvious sdk mapping exists
	ErrUnavailable        = errorsmod.RegisterWithGRPCCode(GErrors, 14, 14, "unavailable")       // no obvious sdk mapping exists
	ErrDataLoss           = errorsmod.RegisterWithGRPCCode(GErrors, 15, 15, "data loss")         // no obvious sdk mapping exists
)
