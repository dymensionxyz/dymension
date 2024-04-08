package delayedack

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	"github.com/tendermint/tendermint/libs/log"
)

var _ rollapptypes.RollappHooks = &IBCMiddleware{}

func (im IBCMiddleware) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
	return nil
}

// AfterStateFinalized implements the RollappHooks interface
func (im IBCMiddleware) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	// Finalize the packets for the rollapp at the given height
	stateEndHeight := stateInfo.StartHeight + stateInfo.NumBlocks - 1
	return im.FinalizeRollappPackets(ctx, rollappID, stateEndHeight)
}

func (im IBCMiddleware) FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	return im.keeper.HandleFraud(ctx, rollappID)
}

// FinalizeRollappPackets finalizes the packets for the given rollapp until the given height which is
// the end height of the latest finalized state
func (im IBCMiddleware) FinalizeRollappPackets(ctx sdk.Context, rollappID string, stateEndHeight uint64) error {
	rollappPendingPackets := im.keeper.ListRollappPackets(ctx, keeper.PendingByRollappIDByMaxHeight(rollappID, stateEndHeight))
	if len(rollappPendingPackets) == 0 {
		return nil
	}
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Get the packets for the rollapp until height
	logger.Debug("Finalizing IBC rollapp packets",
		"rollappID", rollappID,
		"state end height", stateEndHeight,
		"num packets", len(rollappPendingPackets))
	for _, rollappPacket := range rollappPendingPackets {
		if err := im.finalizeRollappPacket(ctx, rollappID, logger, rollappPacket); err != nil {
			return fmt.Errorf("finalize rollapp packet: %w", err)
		}
	}
	return nil
}

type wrappedFunc func(ctx sdk.Context) error

func (im IBCMiddleware) finalizeRollappPacket(
	ctx sdk.Context,
	rollappID string,
	logger log.Logger,
	rollappPacket commontypes.RollappPacket,
) (err error) {
	logger.Debug("Finalizing IBC rollapp packet",
		"rollappID", rollappID,
		"sequence", rollappPacket.Packet.Sequence,
		"source channel", rollappPacket.Packet.SourceChannel,
		"type", rollappPacket.Type)

	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		err = osmoutils.ApplyFuncIfNoError(ctx, im.onRecvPacket(rollappPacket, logger))
	case commontypes.RollappPacket_ON_ACK:
		err = osmoutils.ApplyFuncIfNoError(ctx, im.onAckPacket(rollappPacket, logger))
	case commontypes.RollappPacket_ON_TIMEOUT:
		err = osmoutils.ApplyFuncIfNoError(ctx, im.onTimeoutPacket(rollappPacket, logger))
	default:
		logger.Error("Unknown rollapp packet type",
			"rollappID", rollappID,
			"sequence", rollappPacket.Packet.Sequence,
			"source channel", rollappPacket.Packet.SourceChannel,
			"type", rollappPacket.Type)
	}
	// Update the packet with the error
	if err != nil {
		rollappPacket.Error = err.Error()
	}
	// Update status to finalized
	rollappPacket, err = im.keeper.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
	if err != nil {
		logger.Error("Error finalizing IBC rollapp packet",
			"rollappID", rollappID,
			"sequence", rollappPacket.Packet.Sequence,
			"source channel", rollappPacket.Packet.SourceChannel,
			"type", rollappPacket.Type,
			"error", err.Error())
	}
	return
}

func (im IBCMiddleware) onRecvPacket(rollappPacket commontypes.RollappPacket, logger log.Logger) wrappedFunc {
	return func(ctx sdk.Context) (err error) {
		defer func() {
			if err != nil {
				logger.Error("writing acknowledgement",
					"rollappID", rollappPacket.RollappId,
					"sequence", rollappPacket.Packet.Sequence,
					"source channel", rollappPacket.Packet.SourceChannel,
					"type", rollappPacket.Type,
					"error", err.Error())
			}
		}()
		logger.Debug("Calling OnRecvPacket",
			"rollappID", rollappPacket.RollappId,
			"sequence", rollappPacket.Packet.Sequence,
			"source channel", rollappPacket.Packet.SourceChannel,
			"type", rollappPacket.Type,
		)

		ack := im.IBCModule.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
		// If async, return
		if ack == nil {
			return
		}
		// If sync, check if the acknowledgement is successful
		if !ack.Success() {
			err = sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, string(ack.Acknowledgement()))
			return
		}
		// Write the acknowledgement to the chain only if it is synchronous
		var chanCap *capabilitytypes.Capability
		_, chanCap, err = im.keeper.LookupModuleByChannel(
			ctx,
			rollappPacket.Packet.DestinationPort,
			rollappPacket.Packet.DestinationChannel,
		)
		if err != nil {
			return
		}
		err = im.keeper.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
		return
	}
}

func (im IBCMiddleware) onAckPacket(rollappPacket commontypes.RollappPacket, logger log.Logger) wrappedFunc {
	return func(ctx sdk.Context) (err error) {
		logger.Debug("Calling OnAcknowledgementPacket",
			"rollappID", rollappPacket.RollappId,
			"sequence", rollappPacket.Packet.Sequence,
			"source channel", rollappPacket.Packet.SourceChannel,
			"type", rollappPacket.Type)

		err = im.IBCModule.OnAcknowledgementPacket(
			ctx,
			*rollappPacket.Packet,
			rollappPacket.Acknowledgement,
			rollappPacket.Relayer,
		)
		if err != nil {
			logger.Error("calling OnAcknowledgementPacket",
				"rollappID", rollappPacket.RollappId,
				"sequence", rollappPacket.Packet.Sequence,
				"source channel", rollappPacket.Packet.SourceChannel,
				"type", rollappPacket.Type,
				"error", err.Error())
		}
		return
	}
}

func (im IBCMiddleware) onTimeoutPacket(rollappPacket commontypes.RollappPacket, logger log.Logger) wrappedFunc {
	return func(ctx sdk.Context) (err error) {
		logger.Debug("Calling OnTimeoutPacket",
			"rollappID", rollappPacket.RollappId,
			"sequence", rollappPacket.Packet.Sequence,
			"source channel", rollappPacket.Packet.SourceChannel,
			"type", rollappPacket.Type)

		err = im.IBCModule.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
		if err != nil {
			logger.Error("calling OnTimeoutPacket",
				"rollappID", rollappPacket.RollappId,
				"sequence", rollappPacket.Packet.Sequence,
				"source channel", rollappPacket.Packet.SourceChannel,
				"type", rollappPacket.Type,
				"error", err.Error())
		}
		return
	}
}
