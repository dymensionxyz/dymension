package ante

import (
	context "context"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	AddressCodec() address.Codec
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	IterateAccounts(ctx context.Context, cb func(account sdk.AccountI) bool)
	SetAccount(ctx context.Context, account sdk.AccountI)
	RemoveAccount(ctx context.Context, account sdk.AccountI)
	GetParams(ctx context.Context) (params authtypes.Params)
	GetSequence(ctx context.Context, account sdk.AccAddress) (uint64, error)
	GetAllAccounts(ctx context.Context) (accounts []sdk.AccountI)
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}

// BankKeeper defines the contract needed for supply related APIs (noalias)
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// FeeMarketKeeper defines the expected keeper interface used on the AnteHandler
type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (params feemarkettypes.Params)
	AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error)
	GetBaseFeeEnabled(ctx sdk.Context) bool
	GetMinGasPrice(ctx sdk.Context) (minGasPrice math.LegacyDec)
}

// TxFeesKeeper defines the expected transaction fee keeper
type TxFeesKeeper interface {
	ConvertToBaseToken(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error)
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
	GetFeeToken(ctx sdk.Context, denom string) (txfeestypes.FeeToken, error)
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
