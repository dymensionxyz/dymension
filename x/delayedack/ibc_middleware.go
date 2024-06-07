package delayedack

import (
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/tendermint/tendermint/libs/log"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}

type IBCMiddleware struct {
	porttypes.IBCModule
	*keeper.Keeper
	raKeeper rollappkeeper.Keeper
}

func NewIBCMiddleware(app porttypes.IBCModule, keeper keeper.Keeper, raK rollappkeeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		IBCModule: app,
		Keeper:    &keeper,
		raKeeper:  raK,
	}
}

func (w IBCMiddleware) logger(
	ctx sdk.Context,
	packet channeltypes.Packet,
	method string,
) log.Logger {
	return ctx.Logger().With(
		"module", types.ModuleName,
		"packet_source_port", packet.SourcePort,
		"packet_destination_port", packet.DestinationPort,
		"packet_sequence", packet.Sequence,
		"method", method,
	)
}

// OnRecvPacket handles the receipt of a packet and puts it into a pending queue
// until its state is finalized
func (w IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := w.logger(ctx, packet, "OnRecvPacket")

	if !w.Keeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}
	if types.Skip(ctx) {
		l.Info("Skipping because of skip delay ctx.")
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	transfer, err := w.GetValidTransferDataWithFinalizationInfo(ctx, packet, commontypes.RollappPacket_ON_RECV)
	if err != nil {
		l.Error("Get valid rollapp and transfer data.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if !transfer.HasRollapp() || transfer.Finalized {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// Save the packet transfer to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:   transfer.RollappID,
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: transfer.ProofHeight,
		Type:        commontypes.RollappPacket_ON_RECV,
	}

	w.Keeper.SetRollappPacket(ctx, rollappPacket)

	l.Debug("Set rollapp packet.",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	err = w.eIBCDemandOrderHandler(ctx, rollappPacket, transfer.FungibleTokenPacketData)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return nil
}

// OnAcknowledgementPacket implements the IBCMiddleware interface
func (w IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	if !w.Keeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}
	l := w.logger(ctx, packet, "OnAcknowledgementPacket")

	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		l.Error("Unmarshal acknowledgement", "err", err)
		return errorsmod.Wrapf(types.ErrUnknownRequest, "unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	transfer, err := w.GetValidTransferDataWithFinalizationInfo(ctx, packet, commontypes.RollappPacket_ON_ACK)
	if err != nil {
		l.Error("Get valid rollapp and transfer data.", "err", err)
		return err
	}

	if !transfer.HasRollapp() || transfer.Finalized {
		return w.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}
	// Run the underlying app's OnAcknowledgementPacket callback
	// with cache context to avoid state changes and report the acknowledgement result.
	// Only save the packet if the underlying app's callback succeeds.
	cacheCtx, _ := ctx.CacheContext()
	err = w.IBCModule.OnAcknowledgementPacket(cacheCtx, packet, acknowledgement, relayer)
	if err != nil {
		return err
	}
	// Save the packet transfer to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:       transfer.RollappID,
		Packet:          &packet,
		Acknowledgement: acknowledgement,
		Status:          commontypes.Status_PENDING,
		Relayer:         relayer,
		ProofHeight:     transfer.ProofHeight,
		Type:            commontypes.RollappPacket_ON_ACK,
	}
	w.Keeper.SetRollappPacket(ctx, rollappPacket)

	l.Debug("Set rollapp packet",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	switch ack.Response.(type) {
	// Only if the acknowledgement is an error, we want to create an order
	case *channeltypes.Acknowledgement_Error:
		err = w.eIBCDemandOrderHandler(ctx, rollappPacket, transfer.FungibleTokenPacketData)
		if err != nil {
			return err
		}
	}

	return nil
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (w IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	if !w.Keeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	}

	l := w.logger(ctx, packet, "OnTimeoutPacket")

	transfer, err := w.GetValidTransferDataWithFinalizationInfo(ctx, packet, commontypes.RollappPacket_ON_TIMEOUT)
	if err != nil {
		l.Error("Get valid rollapp and transfer data.", "err", err)
		return err
	}

	if !transfer.HasRollapp() || transfer.Finalized {
		return w.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	}

	// Run the underlying app's OnTimeoutPacket callback
	// with cache context to avoid state changes and report the timeout result.
	// Only save the packet if the underlying app's callback succeeds.
	cacheCtx, _ := ctx.CacheContext()
	err = w.IBCModule.OnTimeoutPacket(cacheCtx, packet, relayer)
	if err != nil {
		return err
	}
	// Save the packet transfer to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:   transfer.RollappID,
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: transfer.ProofHeight,
		Type:        commontypes.RollappPacket_ON_TIMEOUT,
	}
	w.Keeper.SetRollappPacket(ctx, rollappPacket)

	l.Debug("Set rollapp packet",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	err = w.eIBCDemandOrderHandler(ctx, rollappPacket, transfer.FungibleTokenPacketData)
	if err != nil {
		return err
	}

	return nil
}
