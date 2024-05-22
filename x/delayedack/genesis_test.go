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
	var defBridgingFee = types.DefaultParams().BridgingFee

	tests := []struct {
		name           string
		params         types.Params
		rollappPackets []commontypes.RollappPacket
		expPanic       bool
	}{
		{
			name: "only params - success",
			params: types.Params{
				EpochIdentifier: "week",
				BridgingFee:     defBridgingFee,
			},
			rollappPackets: []commontypes.RollappPacket{},
			expPanic:       false,
		},
		{
			name: "only params - missing bridging fee - fail",
			params: types.Params{
				EpochIdentifier: "week",
			},
			rollappPackets: []commontypes.RollappPacket{},
			expPanic:       true,
		},
		{
			name: "params and rollapp packets - panic",
			params: types.Params{
				EpochIdentifier: "week",
				BridgingFee:     defBridgingFee,
			},
			rollappPackets: []commontypes.RollappPacket{{RollappId: "0"}},
			expPanic:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.GenesisState{Params: tt.params, RollappPackets: tt.rollappPackets}
			k, ctx := keepertest.DelayedackKeeper(t)
			if tt.expPanic {
				require.Panics(t, func() {
					delayedack.InitGenesis(ctx, *k, genesisState)
				})
			} else {
				delayedack.InitGenesis(ctx, *k, genesisState)
				params := k.GetParams(ctx)
				require.Equal(t, genesisState.Params, params)
			}
		})
	}
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
		k.SetRollappPacket(ctx, rollappPacket)
	}
	got := delayedack.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, params, got.Params)
	require.Equal(t, rollappPackets, got.RollappPackets)
}
