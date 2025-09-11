package types

import (
	context "context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	otctypes "github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
	AllEpochInfos(ctx sdk.Context) []epochstypes.EpochInfo
}

type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// IncentivesKeeper creates and gets gauges, and also allows additions to gauge rewards.
type IncentivesKeeper interface {
	CreateAssetGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
	CreateRollappGauge(ctx sdk.Context, rollappId string) (uint64, error)
	GetParams(ctx sdk.Context) incentivestypes.Params
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
	Distribute(ctx sdk.Context, gauges []incentivestypes.Gauge, cache incentivestypes.DenomLocksCache, epochEnd bool) (sdk.Coins, error)
	GetDistributeToBaseLocks(ctx sdk.Context, gauge incentivestypes.Gauge, cache incentivestypes.DenomLocksCache) []lockuptypes.PeriodLock
}

type SponsorshipKeeper interface {
	GetDistribution(ctx sdk.Context) (sponsorshiptypes.Distribution, error)
	SaveEndorsement(ctx sdk.Context, e sponsorshiptypes.Endorsement) error
	GetEndorsement(ctx sdk.Context, rollappID string) (sponsorshiptypes.Endorsement, error)
	ClearAllVotes(ctx sdk.Context) error
	GetEndorserPosition(ctx sdk.Context, voterAddr sdk.AccAddress, rollappID string) (sponsorshiptypes.EndorserPosition, error)
	GetVote(ctx sdk.Context, voterAddr sdk.AccAddress) (sponsorshiptypes.Vote, error)
}

type MintParamsGetter interface {
	Get(ctx context.Context) (minttypes.Params, error)
}

type IROKeeper interface {
	GetPlanByRollapp(ctx sdk.Context, rollappId string) (irotypes.Plan, bool)
	BuyExactSpend(ctx sdk.Context, planId string, buyer sdk.AccAddress, amountToSpend, minTokensAmt math.Int) (math.Int, error)
}

type PoolManagerKeeper interface {
	RouteExactAmountIn(
		ctx sdk.Context,
		sender sdk.AccAddress,
		routes []poolmanagertypes.SwapAmountInRoute,
		tokenIn sdk.Coin,
		tokenOutMinAmount math.Int,
	) (tokenOutAmount math.Int, err error)
}

type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, rollappId string) (rollapptypes.Rollapp, bool)
}

type TxFeesKeeper interface {
	GetFeeToken(ctx sdk.Context, denom string) (txfeestypes.FeeToken, error)
	GetBaseDenom(ctx sdk.Context) (string, error)
}

// OtcbuybackKeeper defines the expected interface for the Otcbuyback module.
type OtcbuybackKeeper interface {
	CreateAuction(
		ctx sdk.Context,
		allocation sdk.Coin,
		startTime time.Time,
		endTime time.Time,
		initialDiscount math.LegacyDec,
		maxDiscount math.LegacyDec,
		vestingPlan otctypes.Auction_VestingParams,
		pumpParams otctypes.Auction_PumpParams,
	) (uint64, error)
	EndAuction(ctx sdk.Context, auctionID uint64, reason string) error
}
