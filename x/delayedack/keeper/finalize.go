package keeper

import (
	"fmt"

	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// FinalizeRollappPackets finalizes the packets for the given rollapp until the given height which is
// the end height of the latest finalized state
func (k Keeper) FinalizeRollappPackets(ctx sdk.Context, ibc porttypes.IBCModule, rollappID string, stateEndHeight uint64) error {
	rollappPendingPackets := k.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappID, stateEndHeight))
	if len(rollappPendingPackets) == 0 {
		return nil
	}
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Get the packets for the rollapp until height
	logger.Debug("finalizing IBC rollapp packets",
		"rollappID", rollappID,
		"state end height", stateEndHeight,
		"num packets", len(rollappPendingPackets))
	for _, rollappPacket := range rollappPendingPackets {
		if err := k.finalizeRollappPacket(ctx, ibc, rollappID, logger, rollappPacket); err != nil {
			return fmt.Errorf("finalize rollapp packet: %w", err)
		}
	}
	return nil
}

type wrappedFunc func(ctx sdk.Context) error

func (k Keeper) finalizeRollappPacket(
	ctx sdk.Context,
	ibc porttypes.IBCModule,
	rollappID string,
	logger log.Logger,
	rollappPacket commontypes.RollappPacket,
) error {
	logContext := []interface{}{
		"rollappID", rollappID,
		"sequence", rollappPacket.Packet.Sequence,
		"source channel", rollappPacket.Packet.SourceChannel,
		"destination channel", rollappPacket.Packet.DestinationChannel,
		"type", rollappPacket.Type,
	}

	var packetErr error
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		// TODO: makes more sense to modify the packet when calling the handler, instead storing in db "wrong" packet
		ack := ibc.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
		/*
				We only write the ack if writing it succeeds:
				1. Transfer fails and writing ack fails - In this case, the funds will never be refunded on the RA.
						non-eibc: sender will never get the funds back
						eibc: the fulfiller will never get the funds back, the original target has already been paid
				2. Transfer succeeds and writing ack fails - In this case, the packet is never cleared on the RA.
				3. Transfer succeeds and writing succeeds - happy path
				4. Transfer fails and ack succeeds - we write the err ack and the funds will be refunded on the RA
					 non-eibc: sender will get the funds back
			            eibc: effective transfer from fulfiller to original target
		*/
		if ack != nil {
			packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.writeRecvAck(rollappPacket, ack))
		}
	case commontypes.RollappPacket_ON_ACK:
		packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.onAckPacket(rollappPacket, ibc))
	case commontypes.RollappPacket_ON_TIMEOUT:
		packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.onTimeoutPacket(rollappPacket, ibc))
	default:
		logger.Error("Unknown rollapp packet type", logContext...)
	}
	// Update the packet with the error
	if packetErr != nil {
		rollappPacket.Error = packetErr.Error()
	}
	// Update status to finalized
	_, err := k.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
	if err != nil {
		// If we failed finalizing the packet we return an error to abort the end blocker otherwise it's
		// invariant breaking
		return err
	}

	logger.Debug("finalized IBC rollapp packet", logContext...)
	return nil
}

func (k Keeper) writeRecvAck(rollappPacket commontypes.RollappPacket, ack exported.Acknowledgement) wrappedFunc {
	return func(ctx sdk.Context) (err error) {
		var chanCap *capabilitytypes.Capability
		_, chanCap, err = k.LookupModuleByChannel(
			ctx,
			rollappPacket.Packet.DestinationPort,
			rollappPacket.Packet.DestinationChannel,
		)
		if err != nil {
			return
		}
		/*
			Here, we do the inverse of what we did when we updated the packet transfer address, when we fulfilled the order
			to ensure the ack matches what the rollapp expects.
		*/
		rollappPacket, err = rollappPacket.RestoreOriginalTransferTarget()
		if err != nil {
			return fmt.Errorf("restore original transfer target: %w", err)
		}
		err = k.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
		return
	}
}

func (k Keeper) onAckPacket(rollappPacket commontypes.RollappPacket, ibc porttypes.IBCModule) wrappedFunc {
	return func(ctx sdk.Context) (err error) {
		return ibc.OnAcknowledgementPacket(
			ctx,
			*rollappPacket.Packet,
			rollappPacket.Acknowledgement,
			rollappPacket.Relayer,
		)
	}
}

func (k Keeper) onTimeoutPacket(rollappPacket commontypes.RollappPacket, ibc porttypes.IBCModule) wrappedFunc {
	return func(ctx sdk.Context) (err error) {
		return ibc.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
	}
}
