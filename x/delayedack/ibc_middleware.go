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
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks
type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, keeper keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: keeper,
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
	if !im.keeper.IsRollappsEnabled(ctx) {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	rollappID, err := im.ExtractRollappID(ctx, packet)
	if err != nil {
		logger.Error("Failed to extract rollappID from packet", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}
	if rollappID == "" {
		logger.Debug("Skipping IBC transfer OnRecvPacket for non-rollapp chain")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	finalized, proofHeight, err := im.CheckIfFinalized(ctx, rollappID, packet)
	if err != nil {
		logger.Error("Failed to check if packet is finalized", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}
	if finalized {
		logger.Debug("Skipping IBC transfer OnRecvPacket as the packet proof height is already finalized")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := types.RollappPacket{
		Packet:      &packet,
		Status:      types.RollappPacket_PENDING,
		Relayer:     relayer,
		ProofHeight: proofHeight,
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
	if !im.keeper.IsRollappsEnabled(ctx) {
		return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	rollappID, err := im.ExtractRollappID(ctx, packet)
	if err != nil {
		logger.Error("Failed to extract rollappID from packet", "err", err)
		return err
	}
	if rollappID == "" {
		logger.Debug("Skipping IBC transfer OnAcknowledgementPacket for non-rollapp chain")
		return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	finalized, proofHeight, err := im.CheckIfFinalized(ctx, rollappID, packet)
	if err != nil {
		logger.Error("Failed to check if packet is finalized", "err", err)
		return err
	}
	if finalized {
		logger.Debug("Skipping IBC transfer OnAcknowledgementPacket as the packet proof height is already finalized")
		return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := types.RollappPacket{
		Packet:      &packet,
		Status:      types.RollappPacket_PENDING,
		Relayer:     relayer,
		ProofHeight: proofHeight,
	}
	im.keeper.SetRollappPacket(ctx, rollappID, rollappPacket)
	return nil
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

// extract rollapp from packet
func (im IBCMiddleware) ExtractRollappID(ctx sdk.Context, packet channeltypes.Packet) (string, error) {
	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return "", err
	}

	// Check if the packet is destined for a rollapp
	chainID, err := im.keeper.ExtractChainIDFromChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		return "", err
	}

	_, found := im.keeper.GetRollapp(ctx, chainID)
	if !found {
		return "", nil
	}

	return chainID, nil
}

// CheckIfFinalized checks if the packet is finalized and if so, updates the packet status
func (im IBCMiddleware) CheckIfFinalized(ctx sdk.Context, rollappID string, packet channeltypes.Packet) (bool, uint64, error) {
	// Get the light client height at this block height as a proxy for the packet proof height
	clientState, err := im.keeper.GetClientState(ctx, packet)
	if err != nil {
		return false, 0, err
	}
	proofHeight := clientState.GetLatestHeight().GetRevisionHeight()

	// TODO(omritoptix): Currently we use this height as the proofHeight as the real proofHeight from the ibc lower stack is not available.
	// using this height is secured but may cause extra delay as at best it will be equal to the proof height (but could be higher).
	finalizedHeight, err := im.keeper.GetRollappFinalizedHeight(ctx, rollappID)
	if err != nil {
		return false, 0, err
	}

	return finalizedHeight >= proofHeight, proofHeight, nil
}
