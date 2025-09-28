package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/otcbuyback module sentinel errors
var (
	ErrInvalidAuctionID       = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid auction ID")
	ErrAuctionNotFound        = errorsmod.Wrap(gerrc.ErrNotFound, "auction not found")
	ErrNoUserPurchaseFound    = errorsmod.Wrap(gerrc.ErrNotFound, "no purchase found for user in auction")
	ErrAuctionNotActive       = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "auction is not active")
	ErrAuctionCompleted       = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "auction has already completed")
	ErrTokenNotAccepted       = errorsmod.Wrap(gerrc.ErrInvalidArgument, "token not accepted for this auction")
	ErrInsufficientAllocation = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "insufficient tokens remaining in auction")
	ErrInvalidPurchaseAmount  = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid purchase amount")
	ErrNoClaimableTokens      = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no tokens available to claim")
	ErrInvalidDiscount        = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid discount percentage")
	ErrInvalidAllocation      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid token allocation")
	ErrInvalidEndTime         = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid end time")
	ErrInvalidAddress         = sdkerrors.ErrInvalidAddress
)
