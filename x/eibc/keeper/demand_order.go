package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	"github.com/dymensionxyz/dymension/v3/utils"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (k Keeper) BlockedAddr(addr string) bool {
	account, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false
	}
	return k.bk.BlockedAddr(account)
}

// EIBCDemandOrderHandler handles the eibc packet by creating a demand order from the packet data and saving it in the store.
// the rollapp packet can be of type ON_RECV or ON_TIMEOUT.
// If the rollapp packet is of type ON_RECV, the function will validate the memo and create a demand order from the packet data.
// If the rollapp packet is of type ON_TIMEOUT, the function will calculate the fee and create a demand order from the packet data.
func (k Keeper) EIBCDemandOrderHandler(ctx sdk.Context, rollappPacket commontypes.RollappPacket, data transfertypes.FungibleTokenPacketData) error {
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	packetMetaData := &types.PacketMetadata{}

	switch t := rollappPacket.Type; t {
	case commontypes.RollappPacket_ON_RECV:
		var err error
		packetMetaData, err = types.ParsePacketMetadata(data.Memo)
		if errors.Is(err, types.ErrMemoUnmarshal) || errors.Is(err, types.ErrMemoEibcEmpty) {
			logger.Debug("skipping demand order creation - no eibc memo provided")
			return nil
		}
		if err != nil {
			return err
		}
	case commontypes.RollappPacket_ON_TIMEOUT, commontypes.RollappPacket_ON_ACK:
		// Calculate the fee by multiplying the fee by the price
		amountDec, err := sdk.NewDecFromStr(data.Amount)
		if err != nil {
			return err
		}
		// Calculate the fee by multiplying the fee by the price
		var feeMultiplier sdk.Dec
		switch t {
		case commontypes.RollappPacket_ON_TIMEOUT:
			feeMultiplier = k.TimeoutFee(ctx)
		case commontypes.RollappPacket_ON_ACK:
			feeMultiplier = k.ErrAckFee(ctx)
		}
		fee := amountDec.Mul(feeMultiplier).TruncateInt()
		if !fee.IsPositive() {
			logger.Debug("fee is not positive, skipping demand order creation", "fee type", t, "fee", fee.String(), "multiplier", feeMultiplier.String())
			return nil
		}
		packetMetaData = &types.PacketMetadata{
			EIBC: &types.EIBCMetadata{
				Fee: fee.String(),
			},
		}
	}

	eibcDemandOrder, err := k.createDemandOrderFromIBCPacket(ctx, data, &rollappPacket, *packetMetaData.EIBC)
	if err != nil {
		return fmt.Errorf("create eibc demand order: %w", err)
	}

	err = k.SetDemandOrder(ctx, eibcDemandOrder)
	if err != nil {
		return fmt.Errorf("set eibc demand order: %w", err)
	}
	return nil
}

// createDemandOrderFromIBCPacket creates a demand order from an IBC packet.
// It validates the fungible token packet data, extracts the fee from the memo,
// calculates the demand order price, and creates a new demand order.
// It returns the created demand order or an error if there is any.
func (k *Keeper) createDemandOrderFromIBCPacket(ctx sdk.Context, fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
	rollappPacket *commontypes.RollappPacket, eibcMetaData types.EIBCMetadata,
) (*eibctypes.DemandOrder, error) {
	// Validate the fungible token packet data as we're going to use it to create the demand order
	if err := fungibleTokenPacketData.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := eibcMetaData.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate eibc metadata: %w", err)
	}
	// Verify the original recipient is not a blocked sender otherwise could potentially use eibc to bypass it
	if k.BlockedAddr(fungibleTokenPacketData.Receiver) {
		return nil, fmt.Errorf("not allowed to receive funds: receiver: %s", fungibleTokenPacketData.Receiver)
	}
	// Calculate the demand order price and validate it,
	amt, _ := sdk.NewIntFromString(fungibleTokenPacketData.Amount) // guaranteed ok and positive by above validation

	// Get the fee from the memo
	fee, _ := eibcMetaData.FeeInt() // guaranteed ok by above validation
	if amt.LT(fee) {
		return nil, fmt.Errorf("fee cannot be larger than amount: fee: %s: amt :%s", fee, fungibleTokenPacketData.Amount)
	}

	// Get the bridging fee from the amount
	bridgingFee := k.dack.BridgingFeeFromAmt(ctx, amt)
	demandOrderPrice := amt.Sub(fee).Sub(bridgingFee)
	if !demandOrderPrice.IsPositive() {
		return nil, fmt.Errorf("remaining price is not positive: price: %s, bridging fee: %s, fee: %s, amount: %s",
			demandOrderPrice, bridgingFee, fee, amt)
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

	// Get the denom for the demand order
	var demandOrderDenom string
	var demandOrderRecipient string
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_TIMEOUT:
		fallthrough
	case commontypes.RollappPacket_ON_ACK:
		trace := transfertypes.ParseDenomTrace(fungibleTokenPacketData.Denom)
		demandOrderDenom = trace.IBCDenom()
		demandOrderRecipient = fungibleTokenPacketData.Sender // and who tried to send it (refund because it failed)
	case commontypes.RollappPacket_ON_RECV:
		demandOrderDenom = k.getEIBCTransferDenom(*rollappPacket.Packet, fungibleTokenPacketData)
		demandOrderRecipient = fungibleTokenPacketData.Receiver // who we tried to send to
	}

	order := eibctypes.NewDemandOrder(*rollappPacket, demandOrderPrice, fee, demandOrderDenom, demandOrderRecipient)
	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("validate eibc data: %w", err)
	}
	return order, nil
}

// getEIBCTransferDenom returns the actual denom that will be credited to the eIBC fulfiller.
// The denom logic follows the transfer middleware's logic and is necessary in order to prefix/non-prefix the denom
// based on the original chain it was sent from.
func (k *Keeper) getEIBCTransferDenom(packet channeltypes.Packet, fungibleTokenPacketData transfertypes.FungibleTokenPacketData) string {
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
