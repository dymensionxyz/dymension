package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	keeper "github.com/dymensionxyz/dymension/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}
var _ rollapptypes.RollappHooks = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks
type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
	rollapptypes.BaseRollappHook
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, keeper keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:             app,
		keeper:          keeper,
		BaseRollappHook: rollapptypes.BaseRollappHook{},
	}
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID,
		chanCap, counterparty, version)
}

// OnChanOpenTry implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	// call underlying app's OnChanOpenAck callback with the counterparty app version.
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// call underlying app's OnChanOpenConfirm callback.
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket handles the receipt of a packet and puts it into a pending queue
// until its state is finalized
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Check if the packet is destined for a rollapp
	rollappID, err := im.keeper.GetRollappIDFromPacket(ctx, packet)
	if err != nil {
		logger.Error("failed to extract rollappID from channel", "err", err)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}
	if rollappID == "" {
		logger.Debug("skipping IBC transfer OnRecvPacket for non-tendermint chain")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Get the light client height at this block height as a proxy for the packet proof height
	clientState, err := im.keeper.GetClientState(ctx, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// TODO(omritoptix): Currently we use this height as the proofHeight as the real proofHeight from the ibc lower stack is not available.
	// using this height is secured but may cause extra delay as at best it will be equal to the proof height (but could be higher).
	ibcClientLatestHeight := clientState.GetLatestHeight()

	// Save the packet data to the store for later processing
	rollappPacket := types.RollappPacket{
		Packet:      &packet,
		Status:      types.RollappPacket_PENDING,
		Relayer:     relayer,
		ProofHeight: ibcClientLatestHeight.GetRevisionHeight(),
	}
	im.keeper.SetRollappPacket(ctx, rollappID, rollappPacket)

	return nil
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// call underlying callback
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

/* ------------------------------- ICS4Wrapper ------------------------------ */

// SendPacket implements the ICS4 Wrapper interface
func (im IBCMiddleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	return im.keeper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// GetAppVersion returns the application version of the underlying application
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return im.keeper.GetAppVersion(ctx, portID, channelID)
}

/* ------------------------------- Custom Logic ------------------------------ */


// FinalizeRollappPackets finalizes the packets for the given rollapp until the given height which is
// the end height of the latest finalized state
func (im IBCMiddleware) FinalizeRollappPackets(ctx sdk.Context, rollappID string, stateEndHeight uint64) {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Get the packets for the rollapp until height
	rollappPendingPackets := im.keeper.ListRollappPendingPackets(ctx, rollappID, stateEndHeight)
	logger.Debug("Finalizing IBC rollapp packets", "rollappID", rollappID, "state end height", stateEndHeight, "num packets", len(rollappPendingPackets))
	for _, rollappPacket := range rollappPendingPackets {
		logger.Debug("Finalizing IBC rollapp packet", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
		// Update the packet status
		im.keeper.UpdateRollappPacketStatus(ctx, rollappID, rollappPacket, types.RollappPacket_ACCEPTED)
		// Call the onRecvPacket callback for each packet
		ack := im.app.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
		// Write the acknowledgement to the chain only if it is synchronous
		if ack != nil {
			_, chanCap, err := im.keeper.LookupModuleByChannel(ctx, rollappPacket.Packet.DestinationPort, rollappPacket.Packet.DestinationChannel)
			if err != nil {
				logger.Error("Error looking up module by channel", "rollappID", rollappID, "error", "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), err.Error())
				continue
			}
			err = im.keeper.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
			if err != nil {
				logger.Error("Error writing acknowledgement", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				continue
			}

		}
	}
}
