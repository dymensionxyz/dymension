package transfersenabled

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
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

type ctxKeySkip struct{}

// SkipContext returns a context which can be used when this middleware
// processes received packets in order to skip the transfer enabled check.
func SkipContext(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(ctxKeySkip{}, true)
}

// skip returns if the context contains the skip directive
// Not intended to be used outside of module
func skip(ctx sdk.Context) bool {
	val, ok := ctx.Value(ctxKeySkip{}).(bool)
	return ok && val
}

// OnRecvPacket will block any packet from a rollapp for which transfers are not enabled
// for that rollapp. Pass a skip context to skip the check.
func (w IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := w.logger(ctx, packet)

	_ = l

	if skip(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	if !w.delayedackKeeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.rollappKeeper.GetValidTransferFromReceivedPacket(ctx, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(errorsmod.Wrap(err, "get valid transfer"))
	}

	if !transfer.IsRollapp() {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	ra := w.rollappKeeper.MustGetRollapp(ctx, transfer.RollappID)

	if !ra.GenesisState.TransfersEnabled {
		err = errorsmod.Wrapf(gerr.ErrFailedPrecondition, "transfers are disabled: rollapp id: %s", ra.RollappId)
		// Someone on the RA tried to send a transfer before the bridge is open! Return an err ack and they will get refunded
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
