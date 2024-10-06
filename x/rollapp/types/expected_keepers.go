package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

type IBCClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	SetClientState(ctx sdk.Context, clientID string, clientState exported.ClientState)
}

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type SequencerKeeper interface {
	SlashLiveness(ctx sdk.Context, rollappID string) error
	JailLiveness(ctx sdk.Context, rollappID string) error
	UnbondingTime(ctx sdk.Context) (res time.Duration)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
}

type CanonicalLightClientKeeper interface {
	GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool)
}
