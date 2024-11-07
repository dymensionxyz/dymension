package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func migrateDelayedAckPacketIndex(ctx sdk.Context, dk delayedackkeeper.Keeper) error {
	pendingPackets := dk.ListRollappPackets(ctx, delayedacktypes.ByStatus(commontypes.Status_PENDING))
	for _, packet := range pendingPackets {
		pd, err := packet.GetTransferPacketData()
		if err != nil {
			return err
		}

		switch packet.Type {
		case commontypes.RollappPacket_ON_RECV:
			dk.MustSetPendingPacketByAddress(ctx, pd.Receiver, packet.RollappPacketKey())
		case commontypes.RollappPacket_ON_ACK, commontypes.RollappPacket_ON_TIMEOUT:
			dk.MustSetPendingPacketByAddress(ctx, pd.Sender, packet.RollappPacketKey())
		}
	}
	return nil
}
