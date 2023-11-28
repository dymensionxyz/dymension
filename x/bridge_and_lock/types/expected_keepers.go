package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
}

type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
}

type LockupKeeper interface {
	HasLock(ctx sdk.Context, owner sdk.AccAddress, denom string, duration time.Duration) bool
	CreateLock(ctx sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (lockuptypes.PeriodLock, error)
	AddToExistingLock(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin, duration time.Duration) (uint64, error)
}
