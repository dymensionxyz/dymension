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
	keeper.Keeper
	rollappKeeper types.RollappKeeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper, rollappKeeper types.RollappKeeper) IBCMiddleware {
	return IBCMiddleware{
		IBCModule:     app,
		Keeper:        k,
		rollappKeeper: rollappKeeper,
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
	if !im.IsRollappsEnabled(ctx) {
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	if types.Skip(ctx) {
		l.Info("Skipping because of skip delay ctx.")
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	rollappID, transferPacketData, err := im.rollappKeeper.ExtractRollappIDAndTransferPacketFromData(
		ctx,
		packet.Data,
		rollappPortOnHub,
		rollappChannelOnHub,
	)
	if err != nil {
		l.Error("Get valid rollapp and transfer.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if !transfer.IsFromRollapp() || transfer.Finalized {
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	err = im.ValidateRollappId(ctx, rollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		logger.Error("Failed to validate rollappID", "rollappID", rollappID, "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	proofHeight, err := im.GetProofHeight(ctx, commontypes.RollappPacket_ON_RECV, rollappPortOnHub, rollappChannelOnHub, packet.Sequence)
	if err != nil {
		logger.Error("Failed to get proof height from packet", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	finalized, err := im.CheckIfFinalized(ctx, rollappID, proofHeight)
	if err != nil {
		logger.Error("Failed to check if packet is finalized", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if finalized {
		logger.Debug("Skipping eIBC transfer OnRecvPacket as the packet proof height is already finalized")
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:   rollappID,
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: proofHeight,
		Type:        commontypes.RollappPacket_ON_RECV,
	}
	im.SetRollappPacket(ctx, rollappPacket)

	logger.Debug("Set rollapp packet",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	err = im.eIBCDemandOrderHandler(ctx, rollappPacket, *transferPacketData)
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
	if !im.IsRollappsEnabled(ctx) {
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}
	logger := ctx.Logger().With(
		"module", types.ModuleName,
		"packet_source", packet.SourcePort,
		"packet_destination", packet.DestinationPort,
		"packet_sequence", packet.Sequence)

	if !w.Keeper.IsRollappsEnabled(ctx) {
		return w.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		l.Error("Unmarshal acknowledgement.", "err", err)
		return errorsmod.Wrapf(types.ErrUnknownRequest, "unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	rollappID, transferPacketData, err := im.rollappKeeper.ExtractRollappIDAndTransferPacketFromData(ctx, packet.Data, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("Get valid rollapp and transfer.", "err", err)
		return err
	}

	if rollappID == "" {
		logger.Debug("Skipping IBC transfer OnAcknowledgementPacket for non-rollapp chain")
		return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}
	err = im.ValidateRollappId(ctx, rollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		logger.Error("Failed to validate rollappID", "rollappID", rollappID, "err", err)
		return err
	}

	// Run the underlying app's OnAcknowledgementPacket callback
	// with cache context to avoid state changes and report the acknowledgement result.
	// Only save the packet if the underlying app's callback succeeds.
	cacheCtx, _ := ctx.CacheContext()
	err = w.IBCModule.OnAcknowledgementPacket(cacheCtx, packet, acknowledgement, relayer)
	if err != nil {
		return err
	}
	// Save the packet data to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:       rollappID,
		Packet:          &packet,
		Acknowledgement: acknowledgement,
		Status:          commontypes.Status_PENDING,
		Relayer:         relayer,
		ProofHeight:     proofHeight,
		Type:            commontypes.RollappPacket_ON_ACK,
	}
	im.SetRollappPacket(ctx, rollappPacket)

	rollappPacket := w.getSavedPacket(ctx, l, packet, transfer, relayer, commontypes.RollappPacket_ON_ACK, acknowledgement)

	switch ack.Response.(type) {
	// Only if the acknowledgement is an error, we want to create an order
	case *channeltypes.Acknowledgement_Error:
		return w.eIBCDemandOrderHandler(ctx, rollappPacket, transfer.FungibleTokenPacketData)
	}

	return nil
}

// OnTimeoutPacket implements the IBCMiddleware interface
func (w IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	if !im.IsRollappsEnabled(ctx) {
		return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	}

	rollappPortOnHub, rollappChannelOnHub := packet.SourcePort, packet.SourceChannel

	rollappID, transferPacketData, err := im.rollappKeeper.ExtractRollappIDAndTransferPacketFromData(
		ctx,
		packet.Data,
		rollappPortOnHub,
		rollappChannelOnHub,
	)
	if err != nil {
		l.Error("Get valid rollapp and transfer.", "err", err)
		return err
	}

	if rollappID == "" {
		logger.Debug("Skipping IBC transfer OnTimeoutPacket for non-rollapp chain")
		return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	}

	err = im.ValidateRollappId(ctx, rollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		logger.Error("Failed to validate rollappID", "rollappID", rollappID, "err", err)
		return err
	}

	proofHeight, err := im.GetProofHeight(ctx, commontypes.RollappPacket_ON_TIMEOUT, rollappPortOnHub, rollappChannelOnHub, packet.Sequence)
	if err != nil {
		logger.Error("Failed to get proof height from packet", "err", err)
		return err
	}
	finalized, err := im.CheckIfFinalized(ctx, rollappID, proofHeight)
	if err != nil {
		logger.Error("Failed to check if packet is finalized", "err", err)
		return err
	}

	if finalized {
		logger.Debug("Skipping IBC transfer OnTimeoutPacket as the packet proof height is already finalized")
		return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	}

	// Run the underlying app's OnTimeoutPacket callback
	// with cache context to avoid state changes and report the timeout result.
	// Only save the packet if the underlying app's callback succeeds.
	cacheCtx, _ := ctx.CacheContext()
	err = im.IBCModule.OnTimeoutPacket(cacheCtx, packet, relayer)
	if err != nil {
		return err
	}
	// Save the packet data to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:   rollappID,
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: proofHeight,
		Type:        commontypes.RollappPacket_ON_TIMEOUT,
	}
	im.SetRollappPacket(ctx, rollappPacket)

	logger.Debug("Set rollapp packet",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	err = im.eIBCDemandOrderHandler(ctx, rollappPacket, *transferPacketData)
	if err != nil {
		return err
	}

	rollappPacket := w.getSavedPacket(ctx, l, packet, transfer, relayer, commontypes.RollappPacket_ON_TIMEOUT, nil)

	return w.eIBCDemandOrderHandler(ctx, rollappPacket, transfer.FungibleTokenPacketData)
}

// GetProofHeight returns the proof height of the packet
func (im IBCMiddleware) GetProofHeight(ctx sdk.Context, packetType commontypes.RollappPacket_Type,
	rollappPortOnHub string, rollappChannelOnHub string, sequence uint64,
) (uint64, error) {
	packetId := commontypes.NewPacketUID(packetType, rollappPortOnHub, rollappChannelOnHub, sequence)
	height, ok := types.FromIBCProofContext(ctx, packetId)
	if ok {
		return height.RevisionHeight, nil
	} else {
		err := errors.New("failed to get proof height from context")
		ctx.Logger().Error(err.Error(), "packetId", packetId)
		return 0, err
	}
}

// CheckIfFinalized checks if the packet is finalized and if so, updates the packet status
func (im IBCMiddleware) CheckIfFinalized(ctx sdk.Context, rollappID string, proofHeight uint64) (bool, error) {
	finalizedHeight, err := im.GetRollappFinalizedHeight(ctx, rollappID)
	if err != nil {
		if errors.Is(err, rollapptypes.ErrNoFinalizedStateYetForRollapp) {
			return false, nil
		}
		return false, err
	}

	w.Keeper.SetRollappPacket(ctx, p)

	l.Debug("Set rollapp packet.",
		"rollappID", p.RollappId,
		"proofHeight", p.ProofHeight,
		"type", p.Type)

	return p
}
