package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
	GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool)
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool)
}

type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

type RollappKeeper interface {
	GetParams(ctx sdk.Context) rollapptypes.Params
	GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (val rollapptypes.StateInfo, found bool)
	MustGetStateInfo(ctx sdk.Context, rollappId string, index uint64) rollapptypes.StateInfo
	GetLatestFinalizedStateIndex(ctx sdk.Context, rollappId string) (val types.StateInfoIndex, found bool)
	GetAllRollapps(ctx sdk.Context) (list []types.Rollapp)
	GetValidTransferFromReceivedPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
	) (data rollapptypes.TransferData, err error)
	GetValidTransferFromSentPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
	) (data rollapptypes.TransferData, err error)
}

type SequencerKeeper interface {
	GetSequencer(ctx sdk.Context, sequencerAddress string) (val sequencertypes.Sequencer, found bool)
}
type EIBCKeeper interface {
	SetDemandOrder(ctx sdk.Context, order *eibctypes.DemandOrder) error
	TimeoutFee(ctx sdk.Context) sdk.Dec
	ErrAckFee(ctx sdk.Context) sdk.Dec
}

type BankKeeper interface {
	BlockedAddr(addr sdk.AccAddress) bool
}
