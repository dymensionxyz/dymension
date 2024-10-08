package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// FinalizeRollappPacketsUntilHeight finalizes the packets for the given rollapp until the given height inclusively.
// Returns the number of finalized packets. stateEndHeight is inclusive.
func (k Keeper) FinalizeRollappPacketsUntilHeight(ctx sdk.Context, ibc porttypes.IBCModule, rollappID string, stateEndHeight uint64) (int, error) {
	// Verify the height is finalized
	err := k.VerifyHeightFinalized(ctx, rollappID, stateEndHeight)
	if err != nil {
		return 0, fmt.Errorf("verify height is not finalized: rollapp '%s': %w", rollappID, err)
	}

	// Get all pending rollapp packets until the specified height
	rollappPendingPackets := k.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappID, stateEndHeight))

	// Finalize the packets
	for _, packet := range rollappPendingPackets {
		if err = k.finalizeRollappPacket(ctx, ibc, rollappID, packet); err != nil {
			return 0, fmt.Errorf("finalize packet: rollapp '%s': %w", rollappID, err)
		}
	}

	return len(rollappPendingPackets), nil
}

type FinalizeRollappPacketsBySenderResult struct {
	latestFinalizedHeight uint64 // the latest finalized height of the rollup until which packets are finalized
	finalizedNum          uint64 // the number of finalized packets
}

// FinalizeRollappPacketsByReceiver finalizes the rollapp packets from the specified sender until the latest finalized
// height inclusively. Returns the number of finalized packets.
func (k Keeper) FinalizeRollappPacketsByReceiver(ctx sdk.Context, ibc porttypes.IBCModule, rollappID string, receiver string) (FinalizeRollappPacketsBySenderResult, error) {
	// Get rollapp's latest finalized height. All packets until this height with the specified receiver will be finalized.
	latestFinalizedHeight, err := k.GetRollappLatestFinalizedHeight(ctx, rollappID)
	if err != nil {
		return FinalizeRollappPacketsBySenderResult{}, fmt.Errorf("get latest finalized height: rollapp '%s': %w", rollappID, err)
	}

	// Get all pending rollapp packets until the latest finalized height
	rollappPendingPackets := k.ListRollappPackets(ctx, types.PendingByRollappIDByMaxHeight(rollappID, latestFinalizedHeight))

	// Finalize the packets
	for _, packet := range rollappPendingPackets {
		// Get packet data
		pd, err := packet.GetTransferPacketData()
		if err != nil {
			return FinalizeRollappPacketsBySenderResult{}, fmt.Errorf("get transfer packet data: rollapp '%s', packet: %w", rollappID, err)
		}
		// Finalize a packet if its receiver matches the one specified
		if pd.Receiver == receiver {
			if err = k.finalizeRollappPacket(ctx, ibc, rollappID, packet); err != nil {
				return FinalizeRollappPacketsBySenderResult{}, fmt.Errorf("finalize packet: rollapp '%s': %w", rollappID, err)
			}
		}
	}

	return FinalizeRollappPacketsBySenderResult{
		latestFinalizedHeight: latestFinalizedHeight,
		finalizedNum:          uint64(len(rollappPendingPackets)),
	}, nil
}

// FinalizeRollappPacket finalizes a singe packet by its rollapp packet key.
func (k Keeper) FinalizeRollappPacket(ctx sdk.Context, ibc porttypes.IBCModule, rollappPacketKey string) (*commontypes.RollappPacket, error) {
	// Get a rollapp packet
	packet, err := k.GetRollappPacket(ctx, rollappPacketKey)
	if err != nil {
		return nil, fmt.Errorf("get rollapp packet: %w", err)
	}

	// Verify the height is finalized
	err = k.VerifyHeightFinalized(ctx, packet.RollappId, packet.ProofHeight)
	if err != nil {
		return packet, fmt.Errorf("verify height is finalized: rollapp '%s': %w", packet.RollappId, err)
	}

	// Finalize the packet
	err = k.finalizeRollappPacket(ctx, ibc, packet.RollappId, *packet)
	if err != nil {
		return packet, fmt.Errorf("finalize rollapp packet: %w", err)
	}

	return packet, nil
}

type wrappedFunc func(ctx sdk.Context) error

func (k Keeper) finalizeRollappPacket(
	ctx sdk.Context,
	ibc porttypes.IBCModule,
	rollappID string,
	rollappPacket commontypes.RollappPacket,
) error {
	logger := k.Logger(ctx).With(
		"rollappID", rollappID,
		"sequence", rollappPacket.Packet.Sequence,
		"source channel", rollappPacket.Packet.SourceChannel,
		"destination channel", rollappPacket.Packet.DestinationChannel,
		"type", rollappPacket.Type,
	)

	var packetErr error
	switch rollappPacket.Type {
	case commontypes.RollappPacket_ON_RECV:
		// TODO: makes more sense to modify the packet when calling the handler, instead storing in db "wrong" packet
		ack := ibc.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
		/*
				We only write the ack if writing it succeeds:
				1. Transfer fails and writing ack fails - In this case, the funds will never be refunded on the RA.
						non-eibc: sender will never get the funds back
						eibc:     the fulfiller will never get the funds back, the original target has already been paid
				2. Transfer succeeds and writing ack fails - In this case, the packet is never cleared on the RA.
				3. Transfer succeeds and writing succeeds - happy path
				4. Transfer fails and ack succeeds - we write the err ack and the funds will be refunded on the RA
						non-eibc: sender will get the funds back
			            eibc:     effective transfer from fulfiller to original target
		*/
		if ack != nil {
			packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.writeRecvAck(rollappPacket, ack))
		}
	case commontypes.RollappPacket_ON_ACK:
		packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.onAckPacket(rollappPacket, ibc))
	case commontypes.RollappPacket_ON_TIMEOUT:
		packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.onTimeoutPacket(rollappPacket, ibc))
	default:
		logger.Error("Unknown rollapp packet type")
	}
	// Update the packet with the error
	if packetErr != nil {
		rollappPacket.Error = packetErr.Error()
	}

	// Update status to finalized
	_, err := k.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
	if err != nil {
		return fmt.Errorf("update rollapp packet: %w", err)
	}

	logger.Debug("finalized IBC rollapp packet")

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

func (k Keeper) VerifyHeightFinalized(ctx sdk.Context, rollappID string, height uint64) error {
	// Get the latest state info of the rollapp
	latestIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, rollappID)
	if !found {
		return gerrc.ErrNotFound.Wrapf("latest finalized state index is not found")
	}
	stateInfo, found := k.rollappKeeper.GetStateInfo(ctx, rollappID, latestIndex.Index)
	if !found {
		return gerrc.ErrNotFound.Wrapf("state info is not found")
	}
	// Check the latest finalized height of the rollapp is higher than the height specified
	if height > stateInfo.GetLatestHeight() {
		return gerrc.ErrInvalidArgument.Wrapf("packet height is not finalized yet: height '%d', latest height '%d'", height, stateInfo.GetLatestHeight())
	}
	return nil
}

func (k Keeper) GetRollappLatestFinalizedHeight(ctx sdk.Context, rollappID string) (uint64, error) {
	latestIndex, found := k.rollappKeeper.GetLatestFinalizedStateIndex(ctx, rollappID)
	if !found {
		return 0, gerrc.ErrNotFound.Wrapf("latest finalized state index is not found")
	}
	stateInfo, found := k.rollappKeeper.GetStateInfo(ctx, rollappID, latestIndex.Index)
	if !found {
		return 0, gerrc.ErrNotFound.Wrapf("state info is not found")
	}
	return stateInfo.GetLatestHeight(), nil
}
