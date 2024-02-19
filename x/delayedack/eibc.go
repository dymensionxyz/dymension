package delayedack

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (im IBCMiddleware) handleEIBCPacket(ctx sdk.Context, chainID string, rollappPacket commontypes.RollappPacket, data transfertypes.FungibleTokenPacketData) error {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Handle eibc demand order if exists - Start by validating the memo
	memo := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &memo)
	if err != nil || memo[eibcMemoObjectName] == nil {
		logger.Debug("Memo is empty or failed to unmarshal", "memo", data.Memo)
		return nil
	}
	packetMetaData := &types.PacketMetadata{}
	err = json.Unmarshal([]byte(data.Memo), packetMetaData)
	if err != nil || packetMetaData.ValidateBasic() != nil {
		logger.Error("error parsing packet metadata from memo", "error", err)
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
	rollappPacket *commontypes.RollappPacket, eibcMetaData types.EIBCMetadata) (*eibctypes.DemandOrder, error) {
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
	demandOrderDenom := im.getEIBCTransferDenom(*rollappPacket.Packet, fungibleTokenPacketData)
	// Create the demand order and validate it
	eibcDemandOrder, err := eibctypes.NewDemandOrder(*rollappPacket, demandOrderPrice, fee, demandOrderDenom, fungibleTokenPacketData.Receiver)
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
