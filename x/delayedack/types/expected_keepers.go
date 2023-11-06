package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
}

type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

type RollappKeeper interface {
	GetRollapp(ctx sdk.Context, chainID string) (rollapp rollapptypes.Rollapp, found bool)
}
