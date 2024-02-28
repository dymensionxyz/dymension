package eibc_test

import (
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		DemandOrders: []types.DemandOrder{
			{
				Id:                   "1",
				TrackingPacketKey:    "11/22/33",
				Price:                sdktypes.Coins{sdktypes.Coin{Denom: "adym", Amount: math.NewInt(150)}},
				Fee:                  sdktypes.Coins{sdktypes.Coin{Denom: "adym", Amount: math.NewInt(50)}},
				Recipient:            "dym17g9cn4ss0h0dz5qhg2cg4zfnee6z3ftg3q6v58",
				IsFullfilled:         false,
				TrackingPacketStatus: commontypes.Status_PENDING,
			},
			{
				Id:                   "2",
				TrackingPacketKey:    "22/33/44",
				Price:                sdktypes.Coins{sdktypes.Coin{Denom: "adym", Amount: math.NewInt(250)}},
				Fee:                  sdktypes.Coins{sdktypes.Coin{Denom: "adym", Amount: math.NewInt(150)}},
				Recipient:            "dym15saxgqw6kvhv6k5sg6r45kmdf4sf88kfw2adcw",
				IsFullfilled:         true,
				TrackingPacketStatus: commontypes.Status_REVERTED,
			},
		},
	}

	k, ctx := keepertest.EibcKeeper(t)
	eibc.InitGenesis(ctx, *k, genesisState)
	got := eibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.ElementsMatch(t, genesisState.DemandOrders, got.DemandOrders)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.DemandOrders, got.DemandOrders)

}
