package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	NewAccount(ctx sdk.Context, acc authtypes.AccountI) authtypes.AccountI
	SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasDenomMetaData(ctx sdk.Context, denom string) bool
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)

	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

// IncentivesKeeper defines the expected interface needed to retrieve account balances.
type IncentivesKeeper interface {
	GetLockableDurations(ctx sdk.Context) []time.Duration
	CreateGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
}

// GammKeeper defines the expected interface needed to retrieve account balances.
type GammKeeper interface {
	GetParams(ctx sdk.Context) (params gammtypes.Params)
}

// PoolManagerKeeper defines the expected interface needed to retrieve account balances.
type PoolManagerKeeper interface {
	CreatePool(ctx sdk.Context, msg poolmanagertypes.CreatePoolMsg) (uint64, error)
}

// RollappKeeper defines the expected interface needed to retrieve account balances.
type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, rollappId string) (rollapp rollapptypes.Rollapp, found bool)
	UpdateRollappWithIROPlanAndSeal(ctx sdk.Context, rollappId string, preLaunchTime time.Time)
	MustGetRollapp(ctx sdk.Context, rollappId string) rollapptypes.Rollapp
}
