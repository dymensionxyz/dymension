package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GetValidTransferWithFinalizationInfo does GetValidTransferFromReceivedPacket, but additionally it gets the finalization status and proof height
// of the packet.
func (k Keeper) GetValidTransferWithFinalizationInfo(
	ctx sdk.Context,
	packet channeltypes.Packet,
	packetType commontypes.RollappPacket_Type,
) (data types.TransferDataWithFinalization, err error) {
	var port string
	var channel string

	switch packetType {
	case commontypes.RollappPacket_ON_RECV:
		port, channel = packet.GetDestPort(), packet.GetDestChannel()
	case commontypes.RollappPacket_ON_TIMEOUT, commontypes.RollappPacket_ON_ACK:
		port, channel = packet.GetSourcePort(), packet.GetSourceChannel()
	}

	data.TransferData, err = k.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), port, channel)
	if err != nil {
		err = errorsmod.Wrap(err, "get valid transfer data")
		return
	}

	packetID := commontypes.NewPacketUID(packetType, port, channel, packet.Sequence)
	height, ok := types.PacketProofHeightFromCtx(ctx, packetID)
	if !ok {
		// TODO: should probably be a panic
		err = errorsmod.Wrapf(gerrc.ErrNotFound, "get proof height from context: packetID: %s", packetID)
		return
	}
	data.ProofHeight = height.RevisionHeight

	if !data.IsRollapp() {
		return
	}

	finalizedHeight, err := k.getRollappFinalizedHeight(ctx, data.Rollapp.RollappId)
	if errorsmod.IsOf(err, rollapptypes.ErrNoFinalizedStateYetForRollapp) {
		err = nil
	} else if err != nil {
		err = errorsmod.Wrap(err, "get rollapp finalized height")
	} else {
		data.Finalized = data.ProofHeight <= finalizedHeight
	}

	return
}
