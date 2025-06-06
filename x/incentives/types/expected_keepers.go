package types

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasSupply(ctx context.Context, denom string) bool
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
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
	CalcBaseInCoin(ctx sdk.Context, inputCoin sdk.Coin, denom string) (sdk.Coin, error)
}

type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, rollappId string) (rollapptypes.Rollapp, bool)
}

// SequencerKeeper defines the expected interface needed to interact with sequencer module.
type SequencerKeeper interface {
	GetProposer(ctx sdk.Context, rollappId string) (sequencer sequencertypes.Sequencer)
	Kickable(ctx sdk.Context, proposer sequencertypes.Sequencer) bool
}

type SponsorshipKeeper interface {
	UpdateEndorsementTotalCoins(ctx sdk.Context, rollappID string, additionalCoins sdk.Coins) error
}
