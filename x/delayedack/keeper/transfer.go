package keeper

import (
	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
	types3 "github.com/dymensionxyz/dymension/v3/x/common/types"
	types4 "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	types5 "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// GetValidTransferWithFinalizationInfo does GetValidTransferFromReceivedPacket, but additionally it gets the finalization status and proof height
// of the packet.
func (k Keeper) GetValidTransferWithFinalizationInfo(
	ctx types.Context,
	packet types2.Packet,
	packetType types3.RollappPacket_Type,
) (data types4.TransferDataWithFinalization, err error) {
	switch packetType {
	case types3.RollappPacket_ON_RECV:
		data.TransferData, err = k.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetDestPort(), packet.GetDestChannel())
	case types3.RollappPacket_ON_TIMEOUT, types3.RollappPacket_ON_ACK:
		data.TransferData, err = k.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), packet.GetSourcePort(), packet.GetSourceChannel())
	}
	if err != nil {
		err = errors.Wrap(err, "get valid transfer data")
		return
	}

	packetId := types3.NewPacketUID(packetType, packet.DestinationPort, packet.DestinationChannel, packet.Sequence)
	height, ok := types4.PacketProofHeightFromCtx(ctx, packetId)
	if !ok {
		// TODO: should probably be a panic
		err = errors.Wrapf(gerr.ErrNotFound, "get proof height from context: packetID: %s", packetId)
		return
	}
	data.ProofHeight = height.RevisionHeight

	if !data.IsRollapp() {
		return
	}

	finalizedHeight, err := k.getRollappFinalizedHeight(ctx, data.Rollapp.RollappId)
	if errors.IsOf(err, types5.ErrNoFinalizedStateYetForRollapp) {
		err = nil
	} else if err != nil {
		err = errors.Wrap(err, "get rollapp finalized height")
	} else {
		data.Finalized = data.ProofHeight <= finalizedHeight
	}

	return
}
