package types

import (
	errorsmod "cosmossdk.io/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidGaugeWeight  = errorsmod.RegisterWithGRPCCode(ModuleName, 1, codes.InvalidArgument, "invalid gauge weight")
	ErrInvalidDistribution = errorsmod.RegisterWithGRPCCode(ModuleName, 2, codes.InvalidArgument, "invalid gauge weight distribution")
	ErrInvalidParams       = errorsmod.RegisterWithGRPCCode(ModuleName, 3, codes.InvalidArgument, "invalid params")
	ErrInvalidGenesis      = errorsmod.RegisterWithGRPCCode(ModuleName, 4, codes.InvalidArgument, "invalid genesis")
	ErrInvalidVote         = errorsmod.RegisterWithGRPCCode(ModuleName, 5, codes.InvalidArgument, "invalid vote")
	ErrInvalidVoterInfo    = errorsmod.RegisterWithGRPCCode(ModuleName, 6, codes.InvalidArgument, "invalid voter info")
	ErrNoEndorsers         = errorsmod.RegisterWithGRPCCode(ModuleName, 7, codes.FailedPrecondition, "no endorsers")
)
