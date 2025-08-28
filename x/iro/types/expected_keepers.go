package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)
}

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error
	HasDenomMetadata(ctx sdk.Context, base string) bool
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// IncentivesKeeper defines the expected interface needed to retrieve account balances.
type IncentivesKeeper interface {
	GetParams(ctx sdk.Context) incentivestypes.Params
	CreateAssetGauge(ctx sdk.Context, isPerpetual bool, owner sdk.AccAddress, coins sdk.Coins, distrTo lockuptypes.QueryCondition, startTime time.Time, numEpochsPaidOver uint64) (uint64, error)
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
	SetIROPlanToRollapp(ctx sdk.Context, rollapp *rollapptypes.Rollapp, iro Plan) error
	SetPreLaunchTime(ctx sdk.Context, rollapp *rollapptypes.Rollapp, preLaunchTime time.Time)
	MustGetRollappOwner(ctx sdk.Context, rollappID string) sdk.AccAddress
}

type TxFeesKeeper interface {
	ChargeFeesFromPayer(ctx sdk.Context, payer sdk.AccAddress, takerFeeCoin sdk.Coin, beneficiary *sdk.AccAddress) error
	CalcCoinInBaseDenom(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error)
	CalcBaseInCoin(ctx sdk.Context, inputCoin sdk.Coin, denom string) (sdk.Coin, error)
}
