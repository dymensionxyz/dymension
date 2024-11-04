package delayedack

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	for _, packet := range genState.RollappPackets {
		transferPacketData := packet.MustGetTransferPacketData()
		switch packet.Type {
		case commontypes.RollappPacket_ON_RECV:
			k.MustSetPendingPacketByAddress(ctx, transferPacketData.Receiver, packet.RollappPacketKey())
		case commontypes.RollappPacket_ON_ACK, commontypes.RollappPacket_ON_TIMEOUT:
			k.MustSetPendingPacketByAddress(ctx, transferPacketData.Sender, packet.RollappPacketKey())
		case commontypes.RollappPacket_UNDEFINED:
			panic("invalid rollapp packet type")
		}
		k.SetRollappPacket(ctx, packet)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:         k.GetParams(ctx),
		RollappPackets: k.GetAllRollappPackets(ctx),
	}
}
