package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"
	"github.com/pkg/errors"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	dacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// EIBCDemandOrderHandler handles the eibc packet by creating a demand order from the packet data and saving it in the store.
// the rollapp packet can be of type ON_RECV or ON_TIMEOUT/ON_ACK (with ack error).
// If the rollapp packet is of type ON_RECV, the function will validate the memo and create a demand order from the packet data.
// If the rollapp packet is of type ON_TIMEOUT/ON_ACK, the function will calculate the fee and create a demand order from the packet data.
func (k Keeper) EIBCDemandOrderHandler(ctx sdk.Context, rollappPacket commontypes.RollappPacket, data transfertypes.FungibleTokenPacketData) error {
	var (
		eibcDemandOrder *types.DemandOrder
		err             error
	)
	// Validate the fungible token packet data as we're going to use it to create the demand order
	if err := data.ValidateBasic(); err != nil {
		return err
	}
	// Verify the original recipient is not a blocked sender otherwise could potentially use eibc to bypass it
	if k.BlockedAddr(data.Receiver) {
		return types.ErrBlockedAddress
	}

	switch t := rollappPacket.Type; t {
	case commontypes.RollappPacket_ON_RECV:
		eibcDemandOrder, err = k.CreateDemandOrderOnRecv(ctx, data, &rollappPacket)
	case commontypes.RollappPacket_ON_TIMEOUT, commontypes.RollappPacket_ON_ACK:
		eibcDemandOrder, err = k.CreateDemandOrderOnErrAckOrTimeout(ctx, data, &rollappPacket)
	}
	if err != nil {
		return fmt.Errorf("create eibc demand order: %w", err)
	}
	if eibcDemandOrder == nil {
		return nil
	}
	if err := eibcDemandOrder.Validate(); err != nil {
		return fmt.Errorf("validate eibc data: %w", err)
	}
	err = k.SetDemandOrder(ctx, eibcDemandOrder)
	if err != nil {
		return fmt.Errorf("set eibc demand order: %w", err)
	}

	if err = uevent.EmitTypedEvent(ctx, eibcDemandOrder.GetCreatedEvent(rollappPacket.ProofHeight)); err != nil {
		return fmt.Errorf("emit event: %w", err)
	}

	return nil
}

// CreateDemandOrderOnRecv creates a demand order from an IBC packet.
// It extracts the fee from the memo,calculates the demand order price, and creates a new demand order.
// price calculated with the fee and the bridging fee. (price = amount - fee - bridging fee)
// It returns the created demand order or an error if there is any.
func (k *Keeper) CreateDemandOrderOnRecv(ctx sdk.Context, fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
	rollappPacket *commontypes.RollappPacket,
) (*types.DemandOrder, error) {
	// zero fee demand order by default
	eibcMetaData := dacktypes.EIBCMetadata{Fee: "0"}

	if fungibleTokenPacketData.Memo != "" {
		packetMetaData, err := dacktypes.ParsePacketMetadata(fungibleTokenPacketData.Memo)
		if err == nil {
			eibcMetaData = *packetMetaData.EIBC
		} else if !errors.Is(err, dacktypes.ErrMemoEibcEmpty) {
			return nil, fmt.Errorf("parse packet metadata: %w", err)
		}
	}
	if err := eibcMetaData.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate eibc metadata: %w", err)
	}

	// Calculate the demand order price and validate it,
	amt, _ := sdk.NewIntFromString(fungibleTokenPacketData.Amount) // guaranteed ok and positive by above validation
	fee, _ := eibcMetaData.FeeInt()                                // guaranteed ok by above validation
	demandOrderPrice, err := types.CalcPriceWithBridgingFee(amt, fee, k.dack.BridgingFee(ctx))
	if err != nil {
		return nil, err
	}

	demandOrderDenom := k.getEIBCTransferDenom(*rollappPacket.Packet, fungibleTokenPacketData)
	demandOrderRecipient := fungibleTokenPacketData.Receiver // who we tried to send to

	order := types.NewDemandOrder(*rollappPacket, demandOrderPrice, fee, demandOrderDenom, demandOrderRecipient)
	return order, nil
}

// CreateDemandOrderOnErrAckOrTimeout creates a demand order for a timeout or errack packet.
// The fee multiplier is read from params and used to calculate the fee.
func (k Keeper) CreateDemandOrderOnErrAckOrTimeout(ctx sdk.Context, fungibleTokenPacketData transfertypes.FungibleTokenPacketData,
	rollappPacket *commontypes.RollappPacket,
) (*types.DemandOrder, error) {
	// Calculate the demand order price and validate it,
	amt, _ := sdk.NewIntFromString(fungibleTokenPacketData.Amount) // guaranteed ok and positive by above validation

	// Calculate the fee by multiplying the fee by the price
	var feeMultiplier sdk.Dec
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_TIMEOUT:
		feeMultiplier = k.TimeoutFee(ctx)
	case commontypes.RollappPacket_ON_ACK:
		feeMultiplier = k.ErrAckFee(ctx)
	}
	fee := feeMultiplier.MulInt(amt).TruncateInt()
	if !fee.IsPositive() {
		ctx.Logger().Debug("fee is not positive, skipping demand order creation", "packet", rollappPacket.LogString())
		return nil, nil
	}
	demandOrderPrice := amt.Sub(fee)

	trace := transfertypes.ParseDenomTrace(fungibleTokenPacketData.Denom)
	demandOrderDenom := trace.IBCDenom()
	demandOrderRecipient := fungibleTokenPacketData.Sender // and who tried to send it (refund because it failed)

	order := types.NewDemandOrder(*rollappPacket, demandOrderPrice, fee, demandOrderDenom, demandOrderRecipient)
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
		denom = uibc.GetForeignDenomTrace(packet.GetDestChannel(), fungibleTokenPacketData.Denom).IBCDenom()
	}
	return denom
}

func (k Keeper) BlockedAddr(addr string) bool {
	account, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false
	}
	return k.bk.BlockedAddr(account)
}
