package bridging_fee

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	dakeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
)

const (
	ModuleName = "bridging_fee"
)

var _ porttypes.Middleware = &BridgingFeeMiddleware{}

// BridgingFeeMiddleware implements the ICS26 callbacks
type BridgingFeeMiddleware struct {

	//FIXME: can be embedded?
	transfer transfer.IBCModule
	porttypes.ICS4Wrapper

	keeper         dakeeper.Keeper
	transferKeeper transferkeeper.Keeper
	feeModuleAddr  sdk.AccAddress
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCMiddleware(transfer transfer.IBCModule, channelKeeper porttypes.ICS4Wrapper, keeper dakeeper.Keeper, transferKeeper transferkeeper.Keeper, feeModuleAddr sdk.AccAddress) *BridgingFeeMiddleware {
	return &BridgingFeeMiddleware{
		transfer:       transfer,
		ICS4Wrapper:    channelKeeper,
		keeper:         keeper,
		transferKeeper: transferKeeper,
		feeModuleAddr:  feeModuleAddr,
	}
}

func (im BridgingFeeMiddleware) GetBridgingFee(ctx sdk.Context) sdk.Dec {
	// FIXME: move to param
	return sdk.NewDecWithPrec(1, 1) // 10%
}

// Get the bridging fee param
func (im BridgingFeeMiddleware) GetFeeRecipient(ctx sdk.Context) sdk.AccAddress {
	// FIXME: move to param
	return im.feeModuleAddr
}

func (im *BridgingFeeMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	// skip check if source chain is not a rollapp
	if !im.keeper.IsRollappsEnabled(ctx) {
		return im.transfer.OnRecvPacket(ctx, packet, relayer)
	}
	logger := ctx.Logger().With(
		"module", ModuleName,
		"packet_source", packet.SourcePort,
		"packet_destination", packet.DestinationPort,
		"packet_sequence", packet.Sequence)

	rollappPortOnHub, rollappChannelOnHub := packet.DestinationPort, packet.DestinationChannel
	rollappID, transferPacketData, err := im.keeper.ExtractRollappIDAndTransferPacket(ctx, packet, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		logger.Error("Failed to extract rollapp id from packet", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if rollappID == "" {
		logger.Debug("Skipping IBC transfer OnRecvPacket for non-rollapp chain")
		return im.transfer.OnRecvPacket(ctx, packet, relayer)
	}

	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(transferPacketData.Amount)
	if !ok {
		err = errors.New("unable to parse transfer amount into math.Int")
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// get fee
	feePercentage := im.GetBridgingFee(ctx)
	fee := feePercentage.MulInt(transferAmount).RoundInt()

	// update packet data for the fee charge
	feePacket := *transferPacketData
	feePacket.Amount = fee.String()
	feePacket.Receiver = im.GetFeeRecipient(ctx).String()

	// No event emitted, as we called the transfer keeper directly (vs the transfer middleware)
	err = im.transferKeeper.OnRecvPacket(ctx, packet, feePacket)
	if err != nil {
		logger.Error("Failed to charge bridging fee", "err", err)
		// we continue as we don't want the fee charge to fail the transfer in any case
		fee = sdk.ZeroInt()
	} else {
		logger.Debug("Charged bridging fee", "fee", fee)
		// TODO: Emit events for bridging fee
	}

	transferPacketData.Amount = transferAmount.Sub(fee).String()
	packet.Data = transferPacketData.GetBytes()
	return im.transfer.OnRecvPacket(ctx, packet, relayer)
}

/* -------------------------- unmodified interfaces ------------------------- */
// OnChanCloseConfirm implements types.Middleware.
func (im *BridgingFeeMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.transfer.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements types.Middleware.
func (im *BridgingFeeMiddleware) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return im.transfer.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanOpenAck implements types.Middleware.
func (im *BridgingFeeMiddleware) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return im.transfer.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements types.Middleware.
func (im *BridgingFeeMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.transfer.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanOpenInit implements types.Middleware.
func (im *BridgingFeeMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return im.transfer.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, version)
}

// OnChanOpenTry implements types.Middleware.
func (im *BridgingFeeMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (version string, err error) {
	return im.transfer.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnAcknowledgementPacket implements types.Middleware.
func (im *BridgingFeeMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	return im.transfer.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware.
func (im *BridgingFeeMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.transfer.OnTimeoutPacket(ctx, packet, relayer)
}
