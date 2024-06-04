package gerr

// See doc.go for info

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const GErrors = "gerr"

var (
	// uses canonical codes defined here https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto

	ErrCancelled          = errorsmod.RegisterWithGRPCCode(GErrors, 1, 1, "cancelled")
	ErrUnknown            = errorsmod.RegisterWithGRPCCode(GErrors, 2, 2, "unknown")
	ErrInvalidArgument    = errorsmod.RegisterWithGRPCCode(GErrors, 3, 3, "invalid argument")
	ErrDeadlineExceeded   = errorsmod.RegisterWithGRPCCode(GErrors, 4, 4, "deadline exceeded")
	ErrNotFound           = sdkerrors.ErrKeyNotFound
	ErrAlreadyExist       = errorsmod.RegisterWithGRPCCode(GErrors, 5, 6, "already exist")
	ErrPermissionDenied   = errorsmod.RegisterWithGRPCCode(GErrors, 6, 7, "permission denied")
	ErrUnauthenticated    = errorsmod.RegisterWithGRPCCode(GErrors, 7, 16, "unauthenticated")
	ErrResourceExhausted  = errorsmod.RegisterWithGRPCCode(GErrors, 8, 8, "resource exhausted")
	ErrFailedPrecondition = errorsmod.RegisterWithGRPCCode(GErrors, 9, 9, "failed precondition")
	ErrAborted            = errorsmod.RegisterWithGRPCCode(GErrors, 10, 10, "aborted")
	ErrOutOfRange         = errorsmod.RegisterWithGRPCCode(GErrors, 11, 11, "out of range")
	ErrUnimplemented      = errorsmod.RegisterWithGRPCCode(GErrors, 12, 12, "unimplemented")
	ErrInternal           = errorsmod.RegisterWithGRPCCode(GErrors, 13, 13, "internal")
	ErrUnavailable        = errorsmod.RegisterWithGRPCCode(GErrors, 14, 14, "unavailable")
	ErrDataLoss           = errorsmod.RegisterWithGRPCCode(GErrors, 15, 15, "data loss")
)
