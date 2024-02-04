package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

var _ rollapptypes.RollappHooks = &IBCMiddleware{}

func (im IBCMiddleware) BeforeUpdateState(ctx sdk.Context, seqAddr string, rollappId string) error {
	return nil
}

// AfterStateFinalized implements the RollappHooks interface
func (im IBCMiddleware) AfterStateFinalized(ctx sdk.Context, rollappID string, stateInfo *rollapptypes.StateInfo) error {
	// Finalize the packets for the rollapp at the given height
	stateEndHeight := stateInfo.StartHeight + stateInfo.NumBlocks - 1
	im.FinalizeRollappPackets(ctx, rollappID, stateEndHeight)
	return nil
}

// FinalizeRollappPackets finalizes the packets for the given rollapp until the given height which is
// the end height of the latest finalized state
func (im IBCMiddleware) FinalizeRollappPackets(ctx sdk.Context, rollappID string, stateEndHeight uint64) {
	rollappPendingPackets := im.keeper.ListRollappPendingPackets(ctx, rollappID, stateEndHeight)
	if len(rollappPendingPackets) == 0 {
		return
	}

	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Get the packets for the rollapp until height
	logger.Debug("Finalizing IBC rollapp packets", "rollappID", rollappID, "state end height", stateEndHeight, "num packets", len(rollappPendingPackets))
	for _, rollappPacket := range rollappPendingPackets {
		logger.Debug("Finalizing IBC rollapp packet", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "type", rollappPacket.Type)
		// Update the packet status
		rollappPacket := im.keeper.UpdateRollappPacketStatus(ctx, rollappID, rollappPacket, types.RollappPacket_ACCEPTED)
		// Call the relevant callback for each packet
		switch rollappPacket.Type {
		case types.RollappPacket_ON_RECV:
			logger.Debug("Calling OnRecvPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
			wrappedFunc := func(ctx sdk.Context) error {
				ack := im.app.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
				if !ack.Success() {
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, string(ack.Acknowledgement()))
				}
				// Write the acknowledgement to the chain only if it is synchronous
				if ack != nil {
					_, chanCap, err := im.keeper.LookupModuleByChannel(ctx, rollappPacket.Packet.DestinationPort, rollappPacket.Packet.DestinationChannel)
					if err != nil {
						return err
					}
					err = im.keeper.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
					if err != nil {
						return err
					}
				}
				return nil
			}
			err := osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				logger.Error("Error writing acknowledgement", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				// Update the packet with the error
				rollappPacket.Error = err.Error()
				im.keeper.SetRollappPacket(ctx, rollappPacket)
				continue

			}
		case types.RollappPacket_ON_ACK:
			logger.Debug("Calling OnAcknowledgementPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
			wrappedFunc := func(ctx sdk.Context) error {
				err := im.app.OnAcknowledgementPacket(ctx, *rollappPacket.Packet, rollappPacket.Acknowledgement, rollappPacket.Relayer)
				if err != nil {
					return err
				}
				return nil
			}
			err := osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				logger.Error("Error calling OnAcknowledgementPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				rollappPacket.Error = err.Error()
				im.keeper.SetRollappPacket(ctx, rollappPacket)
				continue
			}
		case types.RollappPacket_ON_TIMEOUT:
			logger.Debug("Calling OnTimeoutPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
			wrappedFunc := func(ctx sdk.Context) error {
				err := im.app.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
				if err != nil {
					return err
				}
				return nil
			}
			err := osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				logger.Error("Error calling OnTimeoutPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				rollappPacket.Error = err.Error()
				im.keeper.SetRollappPacket(ctx, rollappPacket)
			}
		default:
			logger.Error("Unknown rollapp packet type", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "type", rollappPacket.Type)
		}

	}
}
