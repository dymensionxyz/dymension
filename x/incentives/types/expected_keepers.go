package types

import (
	time "time"

	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	seqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin

	HasSupply(ctx sdk.Context, denom string) bool

	SendCoinsFromModuleToAccount(
		ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error

	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

// LockupKeeper defines the expected interface needed to retrieve locks.
type LockupKeeper interface {
	GetLocksLongerThanDurationDenom(ctx sdk.Context, denom string, duration time.Duration) []lockuptypes.PeriodLock
	GetPeriodLocksAccumulation(ctx sdk.Context, query lockuptypes.QueryCondition) sdk.Int
	GetAccountPeriodLocks(ctx sdk.Context, addr sdk.AccAddress) []lockuptypes.PeriodLock
	GetLockByID(ctx sdk.Context, lockID uint64) (*lockuptypes.PeriodLock, error)
}

// EpochKeeper defines the expected interface needed to retrieve epoch info.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) epochstypes.EpochInfo
}

// CommunityPoolKeeper defines the contract needed to be fulfilled for distribution keeper.
type CommunityPoolKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// TxFeesKeeper defines the expected interface needed to managing transaction fees.
type TxFeesKeeper interface {
	GetBaseDenom(ctx sdk.Context) (denom string, err error)
}

type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, rollappId string) (rollapptypes.Rollapp, bool)
}

type SequencerKeeper interface {
	GetProposer(ctx sdk.Context, rollappId string) (seqtypes.Sequencer, bool)
}
