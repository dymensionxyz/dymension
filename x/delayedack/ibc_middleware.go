package delayedack

import (
	"errors"

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
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks
type IBCMiddleware struct {
	porttypes.IBCModule
	*keeper.Keeper
	raKeeper rollappkeeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application
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

	rollappPortOnHub, rollappChannelOnHub := packet.DestinationPort, packet.DestinationChannel

	data, err := w.Keeper.GetRollappAndTransferDataFromPacket(ctx, packet, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("Get transfer data from packet.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if data.RollappID == "" {
		l.Debug("Skipping non-rollapp chain.")
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	err = w.Keeper.ValidateRollappID(ctx, data.RollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("Validate rollappID.", "rollappID", data.RollappID, "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	proofHeight, err := w.GetProofHeight(ctx, commontypes.RollappPacket_ON_RECV, rollappPortOnHub, rollappChannelOnHub, packet.Sequence)
	if err != nil {
		l.Error("Get proof height from packet.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	finalized, err := w.CheckIfFinalized(ctx, data.RollappID, proofHeight)
	if err != nil {
		l.Error("Check if packet is finalized.", "err", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	if finalized {
		l.Debug("Skipping eIBC transfer OnRecvPacket as the packet proof height is already finalized")
		return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// Save the packet data to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:   data.RollappID,
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: proofHeight,
		Type:        commontypes.RollappPacket_ON_RECV,
	}

	w.Keeper.SetRollappPacket(ctx, rollappPacket)

	l.Debug("Set rollapp packet.",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	err = w.eIBCDemandOrderHandler(ctx, rollappPacket, data.FungibleTokenPacketData)
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

	rollappPortOnHub, rollappChannelOnHub := packet.SourcePort, packet.SourceChannel

	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		l.Error("Unmarshal acknowledgement", "err", err)
		return errorsmod.Wrapf(types.ErrUnknownRequest, "unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	data, err := w.Keeper.GetRollappAndTransferDataFromPacket(ctx, packet, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("Get transfer data from packet.", "err", err)
		return err
	}

	if data.RollappID == "" {
		l.Debug("Skipping non-rollapp chain.")
		return w.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	}

	err = w.Keeper.ValidateRollappID(ctx, data.RollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("validate rollappID", "rollappID", data.RollappID, "err", err)
		return err
	}

	proofHeight, err := w.GetProofHeight(ctx, commontypes.RollappPacket_ON_ACK, rollappPortOnHub, rollappChannelOnHub, packet.Sequence)
	if err != nil {
		l.Error("get proof height from packet", "err", err)
		return err
	}

	finalized, err := w.CheckIfFinalized(ctx, data.RollappID, proofHeight)
	if err != nil {
		l.Error("check if packet is finalized", "err", err)
		return err
	}

	if finalized {
		l.Debug("Skipping eIBC transfer OnAcknowledgementPacket as the packet proof height is already finalized")
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
	// Save the packet data to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:       data.RollappID,
		Packet:          &packet,
		Acknowledgement: acknowledgement,
		Status:          commontypes.Status_PENDING,
		Relayer:         relayer,
		ProofHeight:     proofHeight,
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
		err = w.eIBCDemandOrderHandler(ctx, rollappPacket, data.FungibleTokenPacketData)
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

	rollappPortOnHub, rollappChannelOnHub := packet.SourcePort, packet.SourceChannel

	data, err := w.Keeper.GetRollappAndTransferDataFromPacket(ctx, packet, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("Get transfer data from packet.", "err", err)
		return err
	}

	if data.RollappID == "" {
		l.Debug("Skipping non-rollapp chain.")
		return w.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	}

	err = w.Keeper.ValidateRollappID(ctx, data.RollappID, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		l.Error("Validate rollappID.", "rollappID", data.RollappID, "err", err)
		return err
	}

	proofHeight, err := w.GetProofHeight(ctx, commontypes.RollappPacket_ON_TIMEOUT, rollappPortOnHub, rollappChannelOnHub, packet.Sequence)
	if err != nil {
		l.Error("Get proof height from packet.", "err", err)
		return err
	}
	finalized, err := w.CheckIfFinalized(ctx, data.RollappID, proofHeight)
	if err != nil {
		l.Error("Check if packet is finalized.", "err", err)
		return err
	}

	if finalized {
		l.Debug("Skipping because packet proof height is already finalized")
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
	// Save the packet data to the store for later processing
	rollappPacket := commontypes.RollappPacket{
		RollappId:   data.RollappID,
		Packet:      &packet,
		Status:      commontypes.Status_PENDING,
		Relayer:     relayer,
		ProofHeight: proofHeight,
		Type:        commontypes.RollappPacket_ON_TIMEOUT,
	}
	w.Keeper.SetRollappPacket(ctx, rollappPacket)

	l.Debug("Set rollapp packet",
		"rollappID", rollappPacket.RollappId,
		"proofHeight", rollappPacket.ProofHeight,
		"type", rollappPacket.Type)

	err = w.eIBCDemandOrderHandler(ctx, rollappPacket, data.FungibleTokenPacketData)
	if err != nil {
		return err
	}

	return nil
}

// GetProofHeight returns the proof height of the packet
func (w IBCMiddleware) GetProofHeight(ctx sdk.Context, packetType commontypes.RollappPacket_Type,
	rollappPortOnHub string, rollappChannelOnHub string, sequence uint64,
) (uint64, error) {
	packetId := commontypes.NewPacketUID(packetType, rollappPortOnHub, rollappChannelOnHub, sequence)
	height, ok := types.FromIBCProofContext(ctx, packetId)
	if ok {
		return height.RevisionHeight, nil
	} else {
		err := errors.New("get proof height from context")
		ctx.Logger().Error(err.Error(), "packetId", packetId)
		return 0, err
	}
}

// CheckIfFinalized checks if the packet is finalized and if so, updates the packet status
func (w IBCMiddleware) CheckIfFinalized(ctx sdk.Context, rollappID string, proofHeight uint64) (bool, error) {
	finalizedHeight, err := w.Keeper.GetRollappFinalizedHeight(ctx, rollappID)
	if err != nil {
		if errors.Is(err, rollapptypes.ErrNoFinalizedStateYetForRollapp) {
			return false, nil
		}
		return false, err
	}

	return finalizedHeight >= proofHeight, nil
}
