package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
	AllEpochInfos(ctx sdk.Context) []epochstypes.EpochInfo
}

type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// IncentivesKeeper creates and gets gauges, and also allows additions to gauge rewards.
type IncentivesKeeper interface {
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
	CreateRollappGauge(ctx sdk.Context, rollappId string) (uint64, error)

	GetLockableDurations(ctx sdk.Context) []time.Duration

	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
	AddToGaugeRewards(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, gaugeID uint64) error

	Distribute(ctx sdk.Context, gauges []incentivestypes.Gauge, cache incentivestypes.DenomLocksCache, epochEnd bool) (sdk.Coins, error)
	GetDistributeToBaseLocks(ctx sdk.Context, gauge incentivestypes.Gauge, cache map[string][]lockuptypes.PeriodLock) []lockuptypes.PeriodLock
}

type SponsorshipKeeper interface {
	GetDistribution(ctx sdk.Context) (types.Distribution, error)
}
