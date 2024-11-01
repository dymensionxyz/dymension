package bridgingfee

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

const (
	ModuleName = "bridging_fee"
)

// IBCModule is responsible for charging a bridging fee on transfers coming from rollapps
// The actual charge happens on the packet finalization
// based on ADR: https://www.notion.so/dymension/ADR-x-Bridging-Fee-Middleware-7ba8c191373f43ce81782fc759913299?pvs=4
type IBCModule struct {
	ibctransfer.IBCModule

	rollappKeeper    rollappkeeper.Keeper
	delayedAckKeeper delayedackkeeper.Keeper
	transferKeeper   transferkeeper.Keeper
	txFeesKeeper     txfeeskeeper.Keeper
	feeModuleAddr    sdk.AccAddress
}

func NewIBCModule(
	next ibctransfer.IBCModule,
	rollappKeeper rollappkeeper.Keeper,
	delayedAckKeeper delayedackkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	txFeesKeeper txfeeskeeper.Keeper,
	feeModuleAddr sdk.AccAddress,
) *IBCModule {
	return &IBCModule{
		IBCModule:        next,
		rollappKeeper:    rollappKeeper,
		delayedAckKeeper: delayedAckKeeper,
		transferKeeper:   transferKeeper,
		txFeesKeeper:     txFeesKeeper,
		feeModuleAddr:    feeModuleAddr,
	}
}

func (w IBCModule) logger(
	ctx sdk.Context,
	packet channeltypes.Packet,
	method string,
) log.Logger {
	return ctx.Logger().With(
		"module", ModuleName,
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence,
		"method", method,
	)
}

func (w *IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	l := w.logger(ctx, packet, "OnRecvPacket")

	if commontypes.SkipRollappMiddleware(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		l.Error("Get valid transfer.", "err", err)
		err = errorsmod.Wrapf(err, "%s: get valid transfer", ModuleName)
		return uevent.NewErrorAcknowledgement(ctx, err)
	}

	if !transfer.IsRollapp() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// Use the packet as a basis for a fee transfer
	feeData := transfer
	fee := w.delayedAckKeeper.BridgingFeeFromAmt(ctx, transfer.MustAmountInt())
	feeData.Amount = fee.String()
	feeData.Receiver = w.feeModuleAddr.String()

	// No event emitted, as we called the transfer keeper directly (vs the transfer middleware)
	err = w.transferKeeper.OnRecvPacket(ctx, packet, feeData.FungibleTokenPacketData)
	if err != nil {
		l.Error("Charge bridging fee.", "err", err)
		// We continue as we don't want the fee charge to fail the transfer in any case
		fee = sdk.ZeroInt()
	} else {
		// Charge the fee from the txfees module account: construct the IBC denom and use it for the fee coin.
		denomTrace := uibc.GetForeignDenomTrace(packet.GetDestChannel(), feeData.Denom)
		feeCoin := sdk.NewCoin(denomTrace.IBCDenom(), fee)

		err = w.txFeesKeeper.ChargeFees(ctx, feeCoin, nil, transfer.Receiver)
		if err != nil {
			// We continue as we don't want the fee charge to fail the transfer in any case.
			// Also, the fee was already successfully sent to x/txfees and charging will be retried at the epoch end.
			w.logger(ctx, packet, "OnRecvPacket").Error("Charge bridging fee from x/txfees account.", "err", err)
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				EventTypeBridgingFee,
				sdk.NewAttribute(AttributeKeyFee, fee.String()),
				sdk.NewAttribute(sdk.AttributeKeySender, transfer.Sender),
				sdk.NewAttribute(transfertypes.AttributeKeyReceiver, transfer.Receiver),
				sdk.NewAttribute(transfertypes.AttributeKeyDenom, transfer.Denom),
				sdk.NewAttribute(transfertypes.AttributeKeyAmount, transfer.Amount),
			),
		)
	}

	// transfer the rest to the original recipient
	transfer.Amount = transfer.MustAmountInt().Sub(fee).String()
	packet.Data = transfer.GetBytes()
	return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
