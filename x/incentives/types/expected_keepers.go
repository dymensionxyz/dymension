package types

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasSupply(ctx context.Context, denom string) bool
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) math.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

// TxFeesKeeper defines the expected interface needed to managing transaction fees.
type TxFeesKeeper interface {
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
	ChargeFeesFromPayer(ctx sdk.Context, payer sdk.AccAddress, takerFeeCoin sdk.Coin, beneficiary *sdk.AccAddress) error
}

type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, rollappId string) (rollapptypes.Rollapp, bool)
}

type PoolManagerKeeper interface {
	GetPoolModule(ctx sdk.Context, poolId uint64) (poolmanagertypes.SwapI, error)
	GetNextPoolId(ctx sdk.Context) uint64
}

type GammKeeper interface {
	GetPool(ctx sdk.Context, poolId uint64) (poolmanagertypes.PoolI, error)
}

type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
