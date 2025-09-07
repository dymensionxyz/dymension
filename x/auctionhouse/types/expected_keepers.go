package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected interface for the Account module.
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)

	// For creating auction-specific module accounts
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	// Balance operations
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	// Transfer operations
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error

	// Supply operations
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// AMMKeeper defines the expected interface for AMM price fetching.
// This could be a DEX module or external price oracle integration.
type AMMKeeper interface {
	// GetPrice returns the current price of a token in terms of another token
	// For example, GetPrice("DYM", "USDC") returns DYM price in USDC
	GetPrice(ctx sdk.Context, baseDenom, quoteDenom string) (math.LegacyDec, error)

	// GetSpotPrice returns spot price for immediate trades
	GetSpotPrice(ctx sdk.Context, poolID uint64, baseDenom, quoteDenom string) (math.LegacyDec, error)
}
