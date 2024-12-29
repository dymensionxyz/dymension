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
)

func (k Keeper) FinalizeRollappPacket(ctx sdk.Context, ibc porttypes.IBCModule, rollappPacketKey string) (*commontypes.RollappPacket, error) {
	packet, err := k.GetRollappPacket(ctx, rollappPacketKey)
	if err != nil {
		return nil, fmt.Errorf("get rollapp packet: %s: %w", rollappPacketKey, err)
	}

	err = k.VerifyHeightFinalized(ctx, packet.RollappId, packet.ProofHeight)
	if err != nil {
		return packet, fmt.Errorf("verify height: rollapp '%s': %w", packet.RollappId, err)
	}

	err = k.finalizeRollappPacket(ctx, ibc, packet.RollappId, *packet)
	if err != nil {
		return packet, fmt.Errorf("finalize rollapp packet: %w", err)
	}

	return packet, nil
}

// used with osmo helper
type wrappedFunc func(ctx sdk.Context) error

// ibc = the next IBC module in the stack, to 'resume' the rest of the transfer flow
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
		// Because we intercepted the packet, the core ibc library wasn't able to write the ack when we first
		// got the packet. So we try to write it here.

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
		if ack != nil { // NOTE: in practice ack should not be nil, since ibc transfer core module always returns something
			packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.writeRecvAck(rollappPacket, ack))
		}
	case commontypes.RollappPacket_ON_ACK:
		packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.onAckPacket(rollappPacket, ibc))
	case commontypes.RollappPacket_ON_TIMEOUT:
		packetErr = osmoutils.ApplyFuncIfNoError(ctx, k.onTimeoutPacket(rollappPacket, ibc))
	default:
		logger.Error("Unknown rollapp packet type")
	}
	if packetErr != nil {
		// NOTE (timeout,ack): in regular (non dymension) IBC, timeout and ack errors are actually supposed
		//  to cause the delivery transaction to be rejected.
		//  Here, we already accepted the original msg delivery transaction, we can't retroactively reject it.
		rollappPacket.Error = packetErr.Error()
	}

	// Update status to finalized
	_, err := k.UpdateRollappPacketAfterFinalization(ctx, rollappPacket)
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
		// TODO: makes more sense to modify the packet when calling the handler, instead storing in db "wrong" packet (big change)
		rollappPacket = rollappPacket.RestoreOriginalTransferTarget()
		return k.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
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
	// TODO: can extract rollapp keeper IsHeightFinalized method
	latestFinalizedHeight, err := k.getRollappLatestFinalizedHeight(ctx, rollappID)
	if err != nil {
		return err
	}

	// Check the latest finalized height of the rollapp is higher than the height specified
	if height > latestFinalizedHeight {
		return gerrc.ErrInvalidArgument.Wrapf("packet height is not finalized yet: height '%d', latest finalized height '%d'", height, latestFinalizedHeight)
	}
	return nil
}
