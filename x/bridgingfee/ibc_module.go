package bridgingfee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	transferapp "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	ModuleName = "bridging_fee"
)

// IBCModule is responsible for charging a bridging fee on transfers coming from rollapps
// The actual charge happens on the packet finalization
// based on ADR: https://www.notion.so/dymension/ADR-x-Bridging-Fee-Middleware-7ba8c191373f43ce81782fc759913299?pvs=4
type IBCModule struct {
	transferapp.IBCModule

	rollappKeeper    rollappkeeper.Keeper
	delayedAckKeeper delayedackkeeper.Keeper
	transferKeeper   transferkeeper.Keeper
	feeModuleAddr    sdk.AccAddress
}

func NewIBCModule(
	transfer transferapp.IBCModule,
	keeper delayedackkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	feeModuleAddr sdk.AccAddress,
	rollappKeeper rollappkeeper.Keeper,
) *IBCModule {
	return &IBCModule{
		IBCModule:        transfer,
		delayedAckKeeper: keeper,
		transferKeeper:   transferKeeper,
		feeModuleAddr:    feeModuleAddr,
		rollappKeeper:    rollappKeeper,
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

	if !w.delayedAckKeeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		l.Error("Get valid transfer.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
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
	if err == nil {
		l.Error("Charge bridging fee.", "err", err)
		// we continue as we don't want the fee charge to fail the transfer in any case
		fee = sdk.ZeroInt()
	} else {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				EventTypeBridgingFee,
				sdk.NewAttribute(AttributeKeyFee, fee.String()),
				sdk.NewAttribute(sdk.AttributeKeySender, transfer.Sender),
			),
		)
	}

	// transfer the rest to the original recipient
	transfer.Amount = transfer.MustAmountInt().Sub(fee).String()
	packet.Data = transfer.GetBytes()
	return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
