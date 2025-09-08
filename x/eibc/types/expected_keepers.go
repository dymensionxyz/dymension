package types

import (
	context "context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins // TODO: remove, not used
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	BlockedAddr(addr sdk.AccAddress) bool
}

type DelayedAckKeeper interface {
	GetRollappPacket(ctx sdk.Context, rollappPacketKey string) (*commontypes.RollappPacket, error)
	BridgingFee(ctx sdk.Context) (res math.LegacyDec)
	ValidateCompletionHook(info commontypes.CompletionHookCall) error
}

type RollappKeeper interface {
	GetLatestStateInfo(ctx sdk.Context, rollappId string) (rollapptypes.StateInfo, bool)
	IsHeightFinalized(ctx sdk.Context, rollappID string, height uint64) bool
}
