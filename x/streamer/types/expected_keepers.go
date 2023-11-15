package types

import (
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToAccount(
		ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// IncentivesKeeper creates and gets gauges, and also allows additions to gauge rewards.
type IncentivesKeeper interface {
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
	GetGauges(ctx sdk.Context) []incentivestypes.Gauge

	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error
}
