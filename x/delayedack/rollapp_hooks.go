package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
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
	return im.FinalizeRollappPackets(ctx, rollappID, stateEndHeight)
}

func (im IBCMiddleware) FraudSubmitted(ctx sdk.Context, rollappID string, height uint64, seqAddr string) error {
	return im.keeper.HandleFraud(ctx, rollappID)
}

// FinalizeRollappPackets finalizes the packets for the given rollapp until the given height which is
// the end height of the latest finalized state
func (im IBCMiddleware) FinalizeRollappPackets(ctx sdk.Context, rollappID string, stateEndHeight uint64) error {
	rollappPendingPackets := im.keeper.ListRollappPacketsByStatus(ctx, commontypes.Status_PENDING, stateEndHeight)
	if len(rollappPendingPackets) == 0 {
		return nil
	}
	logger := ctx.Logger().With("module", "DelayedAckMiddleware")
	// Get the packets for the rollapp until height
	logger.Debug("Finalizing IBC rollapp packets", "rollappID", rollappID, "state end height", stateEndHeight, "num packets", len(rollappPendingPackets))
	for _, rollappPacket := range rollappPendingPackets {
		logger.Debug("Finalizing IBC rollapp packet", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "type", rollappPacket.Type)
		var err error
		switch rollappPacket.Type {
		case commontypes.RollappPacket_ON_RECV:
			logger.Debug("Calling OnRecvPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
			wrappedFunc := func(ctx sdk.Context) error {
				ack := im.IBCModule.OnRecvPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
				// If async, return
				if ack == nil {
					return nil
				}
				// If sync, check if the acknowledgement is successful
				if !ack.Success() {
					return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, string(ack.Acknowledgement()))
				}
				// Write the acknowledgement to the chain only if it is synchronous
				_, chanCap, err := im.keeper.LookupModuleByChannel(ctx, rollappPacket.Packet.DestinationPort, rollappPacket.Packet.DestinationChannel)
				if err != nil {
					return err
				}
				err = im.keeper.WriteAcknowledgement(ctx, chanCap, rollappPacket.Packet, ack)
				if err != nil {
					return err
				}
				return nil
			}
			err = osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				logger.Error("Error writing acknowledgement", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				rollappPacket.Error = err.Error()
			}
		case commontypes.RollappPacket_ON_ACK:
			logger.Debug("Calling OnAcknowledgementPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
			wrappedFunc := func(ctx sdk.Context) error {
				err := im.IBCModule.OnAcknowledgementPacket(ctx, *rollappPacket.Packet, rollappPacket.Acknowledgement, rollappPacket.Relayer)
				if err != nil {
					return err
				}
				return nil
			}
			err := osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				logger.Error("Error calling OnAcknowledgementPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				rollappPacket.Error = err.Error()
			}
		case commontypes.RollappPacket_ON_TIMEOUT:
			logger.Debug("Calling OnTimeoutPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel())
			wrappedFunc := func(ctx sdk.Context) error {
				err := im.IBCModule.OnTimeoutPacket(ctx, *rollappPacket.Packet, rollappPacket.Relayer)
				if err != nil {
					return err
				}
				return nil
			}
			err := osmoutils.ApplyFuncIfNoError(ctx, wrappedFunc)
			if err != nil {
				logger.Error("Error calling OnTimeoutPacket", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "error", err.Error())
				rollappPacket.Error = err.Error()
			}
		default:
			logger.Error("Unknown rollapp packet type", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "type", rollappPacket.Type)
		}
		// Update the packet with the error
		if err != nil {
			rollappPacket.Error = err.Error()
		}
		// Update status to finalized
		rollappPacket, err = im.keeper.UpdateRollappPacketWithStatus(ctx, rollappPacket, commontypes.Status_FINALIZED)
		if err != nil {
			logger.Error("Error finalizing IBC rollapp packet", "rollappID", rollappID, "sequence", rollappPacket.Packet.GetSequence(), "destination channel", rollappPacket.Packet.GetDestChannel(), "type", rollappPacket.Type, "error", err.Error())
			return err
		}
	}
	return nil
}
