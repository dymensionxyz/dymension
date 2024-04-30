package denommetadata

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"

	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks
type IBCMiddleware struct {
	app            porttypes.IBCModule
	channelKeeper  types.ChannelKeeper
	ics4Wrapper    porttypes.ICS4Wrapper
	transferkeeper types.TransferKeeper
	rollappkeeper  types.RollappKeeper
	bankkeeper     types.BankKeeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, ck types.ChannelKeeper, tk transferkeeper.Keeper, rk rollappkeeper.Keeper, bk bankkeeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:            app,
		channelKeeper:  ck,
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
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	if !im.rollappkeeper.GetParams(ctx).RollappsEnabled {
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	logger := ctx.Logger().With("module", "DenomMiddleware")

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// no-op if the receiver chain is the source chain
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		logger.Debug("Skipping IBC transfer OnRecvPacket for receiver chain being the source chain")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	_, clientState, err := im.channelKeeper.GetChannelClientState(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		logger.Error("failed to extract clientID from channel", "err", err)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	tmClientState, ok := clientState.(*ibctypes.ClientState)
	if !ok {
		logger.Error("failed to extract chainID from clientState")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	chainID := tmClientState.ChainId
	rollapp, found := im.rollappkeeper.GetRollapp(ctx, chainID)
	if !found {
		logger.Debug("Skipping denommetadata middleware. Chain is not a rollapp. ", "chain_id", chainID, "err", err)
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

	if len(rollapp.TokenMetadata) == 0 {
		logger.Info("skipping new IBC token for rollapp with no metadata", "rollappID", chainID, "denom", voucherDenom)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	if im.bankkeeper.HasDenomMetaData(ctx, voucherDenom) {
		logger.Info("denom metadata already registered", "rollappID", chainID, "denom", voucherDenom)
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	for i := range rollapp.TokenMetadata {
		if rollapp.TokenMetadata[i].Base == data.Denom {
			metadata := banktypes.Metadata{
				Description: "auto-generated metadata for " + voucherDenom + " from rollapp " + chainID,
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
				//base denom_unit should be the same as baseDenom
				if newDu.Exponent == 0 {
					newDu.Denom = voucherDenom
					newDu.Aliases = append(newDu.Aliases, du.Denom)
				}
				metadata.DenomUnits[j] = &newDu
			}

			im.bankkeeper.SetDenomMetaData(ctx, metadata)

			logger.Info("registered denom metadata for IBC token", "rollappID", chainID, "denom", voucherDenom)
		}
	}

	return im.app.OnRecvPacket(ctx, packet, relayer)
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
	return im.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements the ICS4 Wrapper interface
func (im IBCMiddleware) WriteAcknowledgement(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
	ack exported.Acknowledgement,
) error {
	return im.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}

// GetAppVersion returns the application version of the underlying application
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return im.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
