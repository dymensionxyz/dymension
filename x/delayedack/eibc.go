package delayedack

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/dymensionxyz/dymension/v3/utils"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

const (
	eibcMemoObjectName = "eibc"
	PFMMemoObjectName  = "forward"
)

// eIBCDemandOrderHandler handles the eibc packet by creating a demand order from the packet data and saving it in the store.
// the rollapp packet can be of type ON_RECV or ON_TIMEOUT.
// If the rollapp packet is of type ON_RECV, the function will validate the memo and create a demand order from the packet data.
// If the rollapp packet is of type ON_TIMEOUT, the function will calculate the fee and create a demand order from the packet data.
func (im IBCMiddleware) eIBCDemandOrderHandler(ctx sdk.Context, rollappPacket commontypes.RollappPacket, data transfertypes.FungibleTokenPacketData) error {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	packetMetaData := &types.PacketMetadata{}
	// If the rollapp packet is of type ON_RECV, the function will validate the memo and create a demand order from the packet data.
	if rollappPacket.Type == commontypes.RollappPacket_ON_RECV {
		// Handle eibc demand order if exists - Start by validating the memo
		memo := make(map[string]interface{})
		err := json.Unmarshal([]byte(data.Memo), &memo)
		if err != nil || memo[eibcMemoObjectName] == nil {
			logger.Debug("Memo is empty or failed to unmarshal", "memo", data.Memo)
			return nil
		}

		// Currently not supporting eibc with PFM: https://github.com/dymensionxyz/dymension/issues/599
		if memo[PFMMemoObjectName] != nil {
			err = fmt.Errorf("EIBC packet with PFM is currently not supported")
			return err
		}
		// Unmarshal the packet metadata from the memo
		err = json.Unmarshal([]byte(data.Memo), packetMetaData)
		if err != nil {
			logger.Error("error parsing packet metadata from memo", "error", err)
			return nil
		}
		// If the rollapp packet is of type ON_TIMEOUT, the function will calculate the fee and create a demand order from the packet data.
	} else if rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
		// Calculate the fee by multiplying the timeout fee by the price
		amountDec, err := sdk.NewDecFromStr(data.Amount)
		if err != nil {
			return err
		}
		// Calculate the fee by multiplying the timeout fee by the price
		timeoutFee := im.keeper.TimeoutFee(ctx)
		if timeoutFee.IsZero() {
			logger.Debug("Timeout fee is zero, skipping demand order creation")
			return nil
		}
		fee := amountDec.Mul(timeoutFee).TruncateInt().String()
		packetMetaData = &types.PacketMetadata{
			EIBC: &types.EIBCMetadata{
				Fee: fee,
			},
		}
	}
	// Validate the packet metadata
	if err := packetMetaData.ValidateBasic(); err != nil {
		logger.Error("error validating packet metadata", "error", err)
		return nil
	}
	// Create the eibc demand order
	eibcDemandOrder, err := im.createDemandOrderFromIBCPacket(data, &rollappPacket, *packetMetaData.EIBC)
	if err != nil {
		err = fmt.Errorf("Failed to create eibc demand order, %s", err)
		return err
	}
	// Save the eibc order in the store
	err = im.keeper.SetDemandOrder(ctx, eibcDemandOrder)
	if err != nil {
		err = fmt.Errorf("Failed to save eibc demand order, %s", err)
		return err
	}
	return nil
}

// createDemandOrderFromIBCPacket creates a demand order from an IBC packet.
// It validates the fungible token packet data, extracts the fee from the memo,
// calculates the demand order price, and creates a new demand order.
// It returns the created demand order or an error if there is any.
func (im IBCMiddleware) createDemandOrderFromIBCPacket(fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
	rollappPacket *commontypes.RollappPacket, eibcMetaData types.EIBCMetadata,
) (*eibctypes.DemandOrder, error) {
	// Validate the fungible token packet data as we're going to use it to create the demand order
	if err := fungibleTokenPacketData.ValidateBasic(); err != nil {
		return nil, err
	}
	// Verify the original recipient is not a blocked sender otherwise could potentially use eibc to bypass it
	if im.keeper.BlockedAddr(fungibleTokenPacketData.Receiver) {
		return nil, fmt.Errorf("%s is not allowed to receive funds", fungibleTokenPacketData.Receiver)
	}
	// Get the fee from the memo
	fee := eibcMetaData.Fee
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
	// Get the denom for the demand order. If it's a timeout packet
	// than its simply the denom we tried to send. If it's a receive packet
	// than it's the IBC denom we've got.
	var demandOrderDenom string
	var demandOrderRecipient string
	if rollappPacket.Type == commontypes.RollappPacket_ON_TIMEOUT {
		demandOrderDenom = fungibleTokenPacketData.Denom
		demandOrderRecipient = fungibleTokenPacketData.Sender
	} else if rollappPacket.Type == commontypes.RollappPacket_ON_RECV {
		demandOrderDenom = im.getEIBCTransferDenom(*rollappPacket.Packet, fungibleTokenPacketData)
		demandOrderRecipient = fungibleTokenPacketData.Receiver
	}
	// Create the demand order and validate it
	eibcDemandOrder, err := eibctypes.NewDemandOrder(*rollappPacket, demandOrderPrice, fee, demandOrderDenom, demandOrderRecipient)
	if err != nil {
		return nil, fmt.Errorf("Failed to create eibc demand order, %s", err)
	}
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
		denom = utils.GetForeignIBCDenom(packet.GetDestChannel(), fungibleTokenPacketData.Denom)
	}
	return denom
}
