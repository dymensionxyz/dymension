package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/otcbuyback module sentinel errors
var (
	ErrInvalidAuctionID       = errorsmod.Register(ModuleName, 1, "invalid auction ID")
	ErrAuctionNotFound        = errorsmod.Wrap(gerrc.ErrNotFound, "auction not found")
	ErrNoUserPurchaseFound    = errorsmod.Wrap(gerrc.ErrNotFound, "no purchase found for user in auction")
	ErrAuctionNotActive       = errorsmod.Register(ModuleName, 3, "auction is not active")
	ErrAuctionCompleted       = errorsmod.Register(ModuleName, 4, "auction has already completed")
	ErrAuctionCancelled       = errorsmod.Register(ModuleName, 5, "auction has been cancelled")
	ErrInvalidTokenDenom      = errorsmod.Register(ModuleName, 6, "invalid token denomination")
	ErrTokenNotAccepted       = errorsmod.Register(ModuleName, 7, "token not accepted for this auction")
	ErrInsufficientAllocation = errorsmod.Register(ModuleName, 8, "insufficient tokens remaining in auction")
	ErrPriceSlippage          = errorsmod.Register(ModuleName, 9, "price slippage protection triggered")
	ErrInvalidPurchaseAmount  = errorsmod.Register(ModuleName, 10, "invalid purchase amount")
	ErrZeroPurchaseAmount     = errorsmod.Register(ModuleName, 11, "purchase amount must be greater than zero")
	ErrNoClaimableTokens      = errorsmod.Register(ModuleName, 13, "no tokens available to claim")
	ErrVestingNotStarted      = errorsmod.Register(ModuleName, 14, "vesting period has not started yet")
	ErrInvalidDiscount        = errorsmod.Register(ModuleName, 15, "invalid discount percentage")
	ErrInvalidDuration        = errorsmod.Register(ModuleName, 16, "invalid auction duration")
	ErrInvalidAllocation      = errorsmod.Register(ModuleName, 17, "invalid token allocation")
	ErrInvalidVestingPeriod   = errorsmod.Register(ModuleName, 18, "invalid vesting period")
	ErrUnauthorized           = errorsmod.Register(ModuleName, 19, "unauthorized")
	ErrInvalidAddress         = errorsmod.Register(ModuleName, 20, "invalid address")
	ErrAMMPriceFetch          = errorsmod.Register(ModuleName, 21, "failed to fetch AMM price")
	ErrTreasuryOperation      = errorsmod.Register(ModuleName, 22, "treasury operation failed")
	ErrInvalidParams          = errorsmod.Register(ModuleName, 23, "invalid module parameters")
	ErrInvalidEndTime         = errorsmod.Register(ModuleName, 24, "invalid end time")
	ErrVestingParam           = errorsmod.Register(ModuleName, 25, "invalid vesting parameter")
)
