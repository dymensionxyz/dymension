package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

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
	GetNextProposerAddr(ctx sdk.Context, rollappId string) (address string, found bool)
}
