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
	denomutils "github.com/dymensionxyz/dymension/v3/utils/denom"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	txfeeskeeper "github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"

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
}

func NewIBCModule(
	next ibctransfer.IBCModule,
	rollappKeeper rollappkeeper.Keeper,
	delayedAckKeeper delayedackkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	txFeesKeeper txfeeskeeper.Keeper,
) *IBCModule {
	return &IBCModule{
		IBCModule:        next,
		rollappKeeper:    rollappKeeper,
		delayedAckKeeper: delayedAckKeeper,
		transferKeeper:   transferKeeper,
		txFeesKeeper:     txFeesKeeper,
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

// OnRecvPacket implements the IBCModule interface. It processes IBC transfer packets,
// charging bridging fees for transfers from rollapps to the hub. The fee is charged
// in the denomination of the incoming tokens, which is determined by:
// - For tokens originating from the hub: the original denomination
// - For tokens originating on the rollapp: the IBC denomination on the hub
func (w IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	l := w.logger(ctx, packet, "OnRecvPacket")

	transfer, err := w.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		l.Error("Get valid transfer.", "err", err)
		err = errorsmod.Wrapf(err, "%s: get valid transfer", ModuleName)
		return uevent.NewErrorAcknowledgement(ctx, err)
	}

	if !transfer.IsRollapp() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// handle the packet transfer
	ack := w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}
	receiver := sdk.MustAccAddressFromBech32(transfer.Receiver)

	feeAmt := w.delayedAckKeeper.BridgingFeeFromAmt(ctx, transfer.MustAmountInt())
	denom := denomutils.GetIncomingTransferDenom(packet, transfer.FungibleTokenPacketData)
	feeCoin := sdk.NewCoin(denom, feeAmt)

	// since transfer worked, then receiver should have enough balance to pay
	// (unless param increased since the delayedck packet was created)
	err = osmoutils.ApplyFuncIfNoError(ctx, func(ctx sdk.Context) error {
		return w.txFeesKeeper.ChargeFeesFromPayer(ctx, receiver, feeCoin, nil)
	})
	if err != nil {
		// We continue as we don't want the fee charge to fail the transfer in any case.
		l.Error("Charge bridging fee from payer.", "receiver", receiver, "err", err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			EventTypeBridgingFee,
			sdk.NewAttribute(AttributeKeyFee, feeAmt.String()),
			sdk.NewAttribute(sdk.AttributeKeySender, transfer.Sender),
			sdk.NewAttribute(transfertypes.AttributeKeyReceiver, transfer.Receiver),
			sdk.NewAttribute(transfertypes.AttributeKeyDenom, transfer.Denom),
			sdk.NewAttribute(transfertypes.AttributeKeyAmount, transfer.Amount),
		),
	)

	return ack
}
