package delayedack_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	k, ctx := keepertest.DelayedackKeeper(t)
	delayedack.InitGenesis(ctx, *k, genesisState)
	require.Equal(t, genesisState.Params, k.GetParams(ctx))
}

func TestExportGenesis(t *testing.T) {
	k, ctx := keepertest.DelayedackKeeper(t)
	// Set params
	params := types.Params{
		EpochIdentifier: "week",
	}
	k.SetParams(ctx, params)
	// Set some demand orders
	rollappPackets := []commontypes.RollappPacket{
		{
			RollappId: "testRollappID",
			Packet: &channeltypes.Packet{
				SourcePort:         "testSourcePort",
				SourceChannel:      "testSourceChannel",
				DestinationPort:    "testDestinationPort",
				DestinationChannel: "testDestinationChannel",
				Data:               []byte("testData"),
				Sequence:           1,
			},
			Status:      commontypes.Status_PENDING,
			ProofHeight: 1,
		},
		{
			RollappId: "testRollappID",
			Packet: &channeltypes.Packet{
				SourcePort:         "testSourcePort",
				SourceChannel:      "testSourceChannel",
				DestinationPort:    "testDestinationPort",
				DestinationChannel: "testDestinationChannel",
				Data:               []byte("testData2"),
				Sequence:           2,
			},
			Status:      commontypes.Status_PENDING,
			ProofHeight: 2,
		},
	}
	for _, rollappPacket := range rollappPackets {
		err := k.SetRollappPacket(ctx, rollappPacket)
		require.NoError(t, err)
	}
	got := delayedack.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, params, got.Params)
	require.Equal(t, rollappPackets, got.RollappPackets)
}
