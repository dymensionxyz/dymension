package delayedack

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	commontypes "github.com/dymensionxyz/dymension/x/common/types"
	keeper "github.com/dymensionxyz/dymension/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
	eibctypes "github.com/dymensionxyz/dymension/x/eibc/types"
)

const (
	eibcMemoObjectName = "eibc"
	eibcMemoFieldFee   = "fee"
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

	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// Check if the packet is destined for a rollapp
	chainID, err := im.keeper.ExtractChainIDFromChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		logger.Error("Failed to extract chain id from channel", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	_, found := im.keeper.GetRollapp(ctx, chainID)
	if !found {
		logger.Debug("Skipping IBC transfer OnRecvPacket for non-rollapp chain")
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
	finalizedHeight, err := im.keeper.GetRollappFinalizedHeight(ctx, chainID)
	if err == nil && finalizedHeight >= ibcClientLatestHeight.GetRevisionHeight() {
		logger.Debug("Skipping IBC transfer OnRecvPacket as the packet proof height is already finalized")
		return im.app.OnRecvPacket(ctx, packet, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := types.RollappPacket{
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: ibcClientLatestHeight.GetRevisionHeight(),
		Type:        types.RollappPacket_ON_RECV,
	}
	im.keeper.SetRollappPacket(ctx, chainID, rollappPacket)

	// Handle eibc demand order if exists
	memo := make(map[string]interface{})
	err = json.Unmarshal([]byte(data.Memo), &memo)
	if err != nil {
		logger.Info("Failed to unmarshal memo field", "err", err)
	}
	if memo[eibcMemoObjectName] != nil {
		rollappPacketStoreKey := types.GetRollappPacketKey(chainID, rollappPacket.Status, rollappPacket.ProofHeight, *rollappPacket.Packet)
		eibcDemandOrder, err := im.createDemandOrderFromIBCPacket(data, &rollappPacket, string(rollappPacketStoreKey), memo)
		if err != nil {
			return channeltypes.NewErrorAcknowledgement(fmt.Errorf("Failed to create eibc demand order, %s", err))
		}
		// Save the eibc order in the store
		im.keeper.SetDemandOrder(ctx, eibcDemandOrder)
	}

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

	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return err
	}

	// Check if the packet is destined for a rollapp
	chainID, err := im.keeper.ExtractChainIDFromChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		logger.Error("Failed to extract chain id from channel", "err", err)
		return err
	}

	_, found := im.keeper.GetRollapp(ctx, chainID)
	if !found {
		logger.Debug("Skipping IBC transfer OnAcknowledgementPacket for non-rollapp chain")
		return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// Get the light client height at this block height as a proxy for the packet proof height
	clientState, err := im.keeper.GetClientState(ctx, packet)
	if err != nil {
		return err
	}

	// TODO(omritoptix): Currently we use this height as the proofHeight as the real proofHeight from the ibc lower stack is not available.
	// using this height is secured but may cause extra delay as at best it will be equal to the proof height (but could be higher).
	ibcClientLatestHeight := clientState.GetLatestHeight()
	finalizedHeight, err := im.keeper.GetRollappFinalizedHeight(ctx, chainID)
	if err == nil && finalizedHeight >= ibcClientLatestHeight.GetRevisionHeight() {
		logger.Debug("Skipping IBC transfer OnAcknowledgementPacket as the packet proof height is already finalized")
		return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := types.RollappPacket{
		Packet:          &packet,
		Acknowledgement: acknowledgement,
		Status:          commontypes.Status_PENDING,
		Relayer:         relayer,
		ProofHeight:     ibcClientLatestHeight.GetRevisionHeight(),
		Type:            types.RollappPacket_ON_ACK,
	}
	im.keeper.SetRollappPacket(ctx, chainID, rollappPacket)

	return nil
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	if !im.keeper.IsRollappsEnabled(ctx) {
		return im.app.OnTimeoutPacket(ctx, packet, relayer)
	}
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")

	// no-op if the packet is not a fungible token packet
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return err
	}

	// Check if the packet is destined for a rollapp
	chainID, err := im.keeper.ExtractChainIDFromChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		logger.Error("Failed to extract chain id from channel", "err", err)
		return err
	}

	_, found := im.keeper.GetRollapp(ctx, chainID)
	if !found {
		logger.Debug("Skipping IBC transfer OnTimeoutPacket for non-rollapp chain")
		return im.app.OnTimeoutPacket(ctx, packet, relayer)
	}

	// Get the light client height at this block height as a proxy for the packet proof height
	clientState, err := im.keeper.GetClientState(ctx, packet)
	if err != nil {
		return err
	}

	// TODO(omritoptix): Currently we use this height as the proofHeight as the real proofHeight from the ibc lower stack is not available.
	// using this height is secured but may cause extra delay as at best it will be equal to the proof height (but could be higher).
	ibcClientLatestHeight := clientState.GetLatestHeight()
	finalizedHeight, err := im.keeper.GetRollappFinalizedHeight(ctx, chainID)
	if err == nil && finalizedHeight >= ibcClientLatestHeight.GetRevisionHeight() {
		logger.Debug("Skipping IBC transfer OnTimeoutPacket as the packet proof height is already finalized")
		return im.app.OnTimeoutPacket(ctx, packet, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := types.RollappPacket{
		Packet:      &packet,
		IsTimeout:   true,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: ibcClientLatestHeight.GetRevisionHeight(),
		Type:        types.RollappPacket_ON_TIMEOUT,
	}
	im.keeper.SetRollappPacket(ctx, chainID, rollappPacket)

	return nil
}

// createDemandOrderFromIBCPacket creates a demand order from an IBC packet.
// It validates the fungible token packet data, extracts the fee from the memo,
// calculates the demand order price, and creates a new demand order.
// It returns the created demand order or an error if there is any.
func (im IBCMiddleware) createDemandOrderFromIBCPacket(fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
	rollappPacket *types.RollappPacket, rollappPacketStoreKey string, memoObj map[string]interface{}) (*eibctypes.DemandOrder, error) {
	// Validate the fungible token packet data as we're going to use it to create the demand order
	if err := fungibleTokenPacketData.ValidateBasic(); err != nil {
		return nil, err
	}
	// Verify the original recipient is not a blocked sender otherwise could potentially use eibc to bypass it
	if im.keeper.BlockedAddr(fungibleTokenPacketData.Receiver) {
		return nil, fmt.Errorf("%s is not allowed to receive funds", fungibleTokenPacketData.Receiver)
	}
	// Get the fee from the memo
	fee := memoObj[eibcMemoObjectName].(map[string]interface{})[eibcMemoFieldFee].(string)
	// Calculate the demand order price and validate it
	amountInt, ok := sdk.NewIntFromString(fungibleTokenPacketData.Amount)
	if !ok || !amountInt.IsPositive() {
		return nil, fmt.Errorf("Failed to convert amount to positive integer, %s", fungibleTokenPacketData.Amount)
	}
	feeInt, ok := sdk.NewIntFromString(fee)
	if !ok || !feeInt.IsPositive() {
		return nil, fmt.Errorf("Failed to convert fee to positive integer, %s", fee)
	}
	if amountInt.LT(feeInt) {
		return nil, fmt.Errorf("Fee cannot be larger than amount")
	}
	demandOrderPrice := amountInt.Sub(feeInt).String()
	demandOrderDenom := im.getEIBCTransferDenom(*rollappPacket.Packet, fungibleTokenPacketData)
	// Create the demand order and validate it
	eibcDemandOrder := eibctypes.NewDemandOrder(rollappPacketStoreKey, demandOrderPrice, fee, demandOrderDenom, fungibleTokenPacketData.Receiver)
	if err := eibcDemandOrder.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to validate eibc data, %s", err)
	}
	return eibcDemandOrder, nil
}

// getEIBCTransferDenom returns the actual denom that will be credited to the eibc fulfillter.
// The denom logic follows the transfer middleware's logic and is necessary in order to prefix/non-prefix the denom
// based on the original chain it was sent from.
func (im IBCMiddleware) getEIBCTransferDenom(packet channeltypes.Packet, fungibleTokenPacketData transfertypes.FungibleTokenPacketData) string {
	var denom string
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), fungibleTokenPacketData.Denom) {
		// remove prefix added by sender chain
		voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := fungibleTokenPacketData.Denom[len(voucherPrefix):]
		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom
		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}
	} else {
		// sender chain is the source
		// since SendPacket did not prefix the denomination, we must prefix denomination here
		sourcePrefix := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		// NOTE: sourcePrefix contains the trailing "/"
		prefixedDenom := sourcePrefix + fungibleTokenPacketData.Denom
		// construct the denomination trace from the full raw denomination
		denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
		denom = denomTrace.IBCDenom()
	}
	return denom
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
