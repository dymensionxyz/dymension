package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

// TransferKeeper defines the expected transfer keeper
type TransferKeeper interface {
	Transfer(ctx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error)
	DenomPathFromHash(ctx sdk.Context, denom string) (string, error)
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte
	GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool)
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}
