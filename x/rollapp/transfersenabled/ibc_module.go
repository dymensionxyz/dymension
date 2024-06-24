package transfersenabled

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/tendermint/tendermint/libs/log"
)

type IBCModule struct {
	porttypes.IBCModule // next one
	rollappKeeper       rollappkeeper.Keeper
	delayedackKeeper    delayedackkeeper.Keeper
}

func NewIBCModule(
	next porttypes.IBCModule,
	rollappKeeper rollappkeeper.Keeper,
	delayedAckKeeper delayedackkeeper.Keeper,
) IBCModule {
	return IBCModule{
		IBCModule:        next,
		rollappKeeper:    rollappKeeper,
		delayedackKeeper: delayedAckKeeper,
	}
}

func (w IBCModule) logger(
	ctx sdk.Context,
	packet channeltypes.Packet,
) log.Logger {
	return ctx.Logger().With(
		"module", "transferEnabled",
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence,
		"method", "OnRecvPacket",
	)
}

// OnRecvPacket will block any packet from a rollapp for which transfers are not enabled
// for that rollapp. Pass a skip context to skip the check.
func (w IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := w.logger(ctx, packet)

	if commontypes.SkipRollappMiddleware(ctx) || !w.delayedackKeeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "transfer enabled?: get valid transfer"))
	}

	if !transfer.IsRollapp() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	if !transfer.Rollapp.GenesisState.TransfersEnabled {
		// Someone on the RA tried to send a transfer before the bridge is open! Return an err ack and they will get refunded
		err = errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "transfers are disabled: rollapp id: %s", transfer.Rollapp.RollappId)
		l.Debug("Returning error ack.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
