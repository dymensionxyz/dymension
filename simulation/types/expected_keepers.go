package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type IncentivesKeeper interface {
	GetGauges(ctx sdk.Context) []incentivestypes.Gauge
	GetGaugeByID(ctx sdk.Context, gaugeID uint64) (*incentivestypes.Gauge, error)
}

type StreamerK interface {
	CreateStream(ctx sdk.Context, coins sdk.Coins, records []streamertypes.DistrRecord, startTime time.Time, epochIdentifier string, numEpochsPaidOver uint64, sponsored bool) (uint64, error)
	GetStreams(ctx sdk.Context) []streamertypes.Stream
	ReplaceDistrRecords(ctx sdk.Context, streamId uint64, records []streamertypes.DistrRecord) error
	UpdateDistrRecords(ctx sdk.Context, streamId uint64, records []streamertypes.DistrRecord) error
	TerminateStream(ctx sdk.Context, streamID uint64) error
}

type StakingKeeper interface {
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) []stakingtypes.Delegation
}
