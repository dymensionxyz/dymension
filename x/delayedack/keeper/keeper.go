package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		rollappKeeper    types.RollappKeeper
		ics4Wrapper      porttypes.ICS4Wrapper
		channelKeeper    types.ChannelKeeper
		connectionKeeper types.ConnectionKeeper
		clientKeeper     types.ClientKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,

	rollappKeeper types.RollappKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	connectionKeeper types.ConnectionKeeper,
	clientKeeper types.ClientKeeper,

) *Keeper {
	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		rollappKeeper:    rollappKeeper,
		ics4Wrapper:      ics4Wrapper,
		channelKeeper:    channelKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ExtractChainIDFromChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", fmt.Errorf("failed to extract clientID from channel: %w", err)
	}

	tmClientState, ok := clientState.(*ibctypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmClientState.ChainId, nil
}

func (k Keeper) GetRollapp(ctx sdk.Context, chainID string) (rollapptypes.Rollapp, bool) {
	return k.rollappKeeper.GetRollapp(ctx, chainID)
}

func (k Keeper) GetRollappFinalizedHeight(ctx sdk.Context, chainID string) (uint64, error) {
	res, err := k.rollappKeeper.StateInfo(ctx, &rollapptypes.QueryGetStateInfoRequest{
		RollappId: chainID,
		Finalized: true,
	})
	if err != nil {
		return 0, err
	}

	return (res.StateInfo.StartHeight + res.StateInfo.NumBlocks - 1), nil
}

// GetClientState retrieves the client state for a given packet.
func (k Keeper) GetClientState(ctx sdk.Context, packet channeltypes.Packet) (exported.ClientState, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if !found {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelNotFound, packet.SourceChannel)
	}
	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.GetClientID())
	if !found {
		return nil, clienttypes.ErrConsensusStateNotFound
	}

	return clientState, nil
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (k Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper WriteAcknowledgement function.
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements.
func (k *Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, acknowledgement exported.Acknowledgement) error {
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper GetAppVersion function.
func (k *Keeper) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// LookupModuleByChannel wraps ChannelKeeper LookupModuleByChannel function.
func (k *Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return k.channelKeeper.LookupModuleByChannel(ctx, portID, channelID)
}
