package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GetValidTransferWithFinalizationInfo does GetValidTransferFromReceivedPacket, but additionally it gets the finalization status and proof height
// of the packet.
func (k Keeper) GetValidTransferWithFinalizationInfo(
	ctx sdk.Context,
	packet channeltypes.Packet,
	packetType commontypes.RollappPacket_Type,
) (data types.TransferDataWithFinalization, err error) {
	port, channel := commontypes.PacketHubPortChan(packetType, packet)

	height, err := commontypes.UnpackPacketProofHeight(ctx, packet, packetType)
	if err != nil {
		err = errorsmod.Wrap(err, "unpack packet proof height")
		return
	}
	data.ProofHeight = height

	data.TransferData, err = k.rollappKeeper.GetValidTransfer(ctx, packet.GetData(), port, channel)
	if err != nil {
		err = errorsmod.Wrap(err, "get valid transfer data")
		return
	}

	if !data.IsRollapp() {
		return
	}

	// TODO: can extract rollapp keeper IsHeightFinalized method
	finalizedHeight, err := k.getRollappLatestFinalizedHeight(ctx, data.Rollapp.RollappId)
	if errorsmod.IsOf(err, gerrc.ErrNotFound) {
		err = nil
	} else if err != nil {
		err = errorsmod.Wrap(err, "get rollapp finalized height")
	} else {
		data.Finalized = data.ProofHeight <= finalizedHeight
	}

	return
}
