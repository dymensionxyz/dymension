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

	switch t := rollappPacket.Type; t {
	case commontypes.RollappPacket_ON_RECV:
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
			logger.Error("parsing packet metadata from memo", "error", err)
			return nil
		}
	case commontypes.RollappPacket_ON_TIMEOUT, commontypes.RollappPacket_ON_ACK:
		// Calculate the fee by multiplying the fee by the price
		amountDec, err := sdk.NewDecFromStr(data.Amount)
		if err != nil {
			return err
		}
		// Calculate the fee by multiplying the fee by the price
		var feeMultiplier sdk.Dec
		if t == commontypes.RollappPacket_ON_TIMEOUT {
			feeMultiplier = im.keeper.TimeoutFee(ctx)
		}
		if t == commontypes.RollappPacket_ON_ACK {
			feeMultiplier = im.keeper.ErrAckFee(ctx)
		}

		fee := amountDec.Mul(feeMultiplier).TruncateInt()
		if !fee.IsPositive() {
			logger.Debug("fee is not positive, skipping demand order creation", "fee type", t, "fee", fee.String())
			return nil
		}
		packetMetaData = &types.PacketMetadata{
			EIBC: &types.EIBCMetadata{
				Fee: fee.String(),
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
		err = fmt.Errorf("create eibc demand order: %s", err)
		return err
	}
	// Save the eibc order in the store
	err = im.keeper.SetDemandOrder(ctx, eibcDemandOrder)
	if err != nil {
		err = fmt.Errorf("save eibc demand order: %s", err)
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
		return nil, fmt.Errorf("not allowed to receive funds: receiver: %s", fungibleTokenPacketData.Receiver)
	}
	// Calculate the demand order price and validate it
	amountInt, ok := sdk.NewIntFromString(fungibleTokenPacketData.Amount)
	if !ok || !amountInt.IsPositive() {
		return nil, fmt.Errorf("convert amount to positive integer: %s", fungibleTokenPacketData.Amount)
	}
	// Get the fee from the memo
	fee := eibcMetaData.Fee
	feeInt, ok := sdk.NewIntFromString(fee)
	if !ok || !feeInt.IsPositive() {
		return nil, fmt.Errorf("convert fee to positive integer: %s", fee)
	}
	if amountInt.LT(feeInt) {
		return nil, fmt.Errorf("fee cannot be larger than amount")
	}

	/*
		   In case of timeout/errack:
		       fee = fee_multiplier*transfer_amount
		       price = transfer_amount-fee
		       order is created with (price,fee)
		       the order creator immediately receives price on fulfillment from fulfiller
		       when the ack/timeout finalizes, the fulfiller receives the transfer_amount

		       therefore:
		           fulfiller balance += (transfer_amount - (transfer_amount-fee))
					   equivalent to += fee_multiplier*transfer_amount
		           demander balance += (transfer_amount - fee)
		              equivalent to += (1-fee_multiplier)*transfer_amount
	*/
	demandOrderPrice := amountInt.Sub(feeInt).String()

	var demandOrderDenom string
	var demandOrderRecipient string
	// Get the denom for the demand order
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_TIMEOUT:
		fallthrough
	case commontypes.RollappPacket_ON_ACK:
		demandOrderDenom = fungibleTokenPacketData.Denom      // It's what we tried to send
		demandOrderRecipient = fungibleTokenPacketData.Sender // and who tried to send it (refund because it failed)
	case commontypes.RollappPacket_ON_RECV:
		demandOrderDenom = im.getEIBCTransferDenom(*rollappPacket.Packet, fungibleTokenPacketData)
		demandOrderRecipient = fungibleTokenPacketData.Receiver // who we tried to send to
	}
	// Create the demand order and validate it
	eibcDemandOrder, err := eibctypes.NewDemandOrder(*rollappPacket, demandOrderPrice, fee, demandOrderDenom, demandOrderRecipient)
	if err != nil {
		return nil, fmt.Errorf("create eibc demand order: %s", err)
	}
	if err := eibcDemandOrder.Validate(); err != nil {
		return nil, fmt.Errorf("validate eibc data: %s", err)
	}
	return eibcDemandOrder, nil
}

// getEIBCTransferDenom returns the actual denom that will be credited to the eIBC fulfiller.
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
