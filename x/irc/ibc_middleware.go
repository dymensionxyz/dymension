package irc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	keeper "github.com/dymensionxyz/dymension/x/irc/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/x/rollapp/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the fee middleware given the
// fee keeper and the underlying application.
type IBCMiddleware struct {
	app            porttypes.IBCModule
	keeper         keeper.Keeper
	transferkeeper transferkeeper.Keeper
	rollappkeeper  rollappkeeper.Keeper
	bankkeeper     bankkeeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper, tk transferkeeper.Keeper, rk rollappkeeper.Keeper, bk bankkeeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:            app,
		keeper:         k,
		transferkeeper: tk,
		rollappkeeper:  rk,
		bankkeeper:     bk,
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
// If the channel is not fee enabled the underlying application version will be returned
// If the channel is fee enabled we merge the underlying application version with the ics29 version
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

// OnRecvPacket implements the IBCMiddleware interface.
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	im.keeper.Logger(ctx).Info("GREATEST MESSAGE ARRIVED OnRecvPacket",
		"sequence", packet.Sequence,
		"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
		"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
		"amount", data.Amount, "denom", data.Denom, "memo", data.Memo,
	)

	rollappID, err := im.keeper.ExtractRollappIDFromChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		ctx.Logger().Error("failed to extract channelID from channel", "err", err)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}
	// no-op if rollappID is empty (i.e transfer from non-dymint chain)
	if rollappID == "" {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// no-op if the receiver chain is the source chain
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + data.Denom
	// construct the denomination trace from the full raw denomination
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	traceHash := denomTrace.Hash()
	voucherDenom := denomTrace.IBCDenom()

	// no-op if token already exist
	if im.transferkeeper.HasDenomTrace(ctx, traceHash) {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	im.keeper.Logger(ctx).Info("Extracted rollappID from channel", "channelID", packet.GetDestChannel(), "rollappID", rollappID)
	rollapp, found := im.rollappkeeper.GetRollapp(ctx, rollappID)
	if !found {
		panic("handling IBC transfer packet for non-existing rollapp")
	}

	for i := range rollapp.TokenMetadata {
		if rollapp.TokenMetadata[i].Base == data.Denom {
			metadata := banktypes.Metadata{
				Description: "auto-generated metadata for " + voucherDenom + " from rollapp " + rollappID,
				Base:        voucherDenom,
				DenomUnits:  make([]*banktypes.DenomUnit, len(rollapp.TokenMetadata[i].DenomUnits)),
				Display:     rollapp.TokenMetadata[i].Display,
				Name:        rollapp.TokenMetadata[i].Name,
				Symbol:      rollapp.TokenMetadata[i].Symbol,
				URI:         rollapp.TokenMetadata[i].URI,
				URIHash:     rollapp.TokenMetadata[i].URIHash,
			}
			// Copy DenomUnits slice
			for j, du := range rollapp.TokenMetadata[i].DenomUnits {
				newDu := banktypes.DenomUnit{
					Aliases:  du.Aliases,
					Denom:    du.Denom,
					Exponent: du.Exponent,
				}
				metadata.DenomUnits[j] = &newDu
			}

			im.bankkeeper.SetDenomMetaData(ctx, metadata)
		}
	}

	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCMiddleware interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
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
	packet exported.PacketI,
) error {
	return im.keeper.SendPacket(ctx, chanCap, packet)
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
