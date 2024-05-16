package eibc_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		DemandOrders: []types.DemandOrder{
			{
				Id:                   "1",
				TrackingPacketKey:    "11/22/33",
				Price:                sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(150)}},
				Fee:                  sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(50)}},
				Recipient:            "dym17g9cn4ss0h0dz5qhg2cg4zfnee6z3ftg3q6v58",
				IsFullfilled:         false,
				TrackingPacketStatus: commontypes.Status_PENDING,
			},
			{
				Id:                   "2",
				TrackingPacketKey:    "22/33/44",
				Price:                sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(250)}},
				Fee:                  sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(150)}},
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

func TestExportGenesis(t *testing.T) {
	k, ctx := keepertest.EibcKeeper(t)
	params := types.Params{
		EpochIdentifier: "week",
		TimeoutFee:      sdk.NewDecWithPrec(4, 1),
		ErrackFee:       sdk.NewDecWithPrec(4, 1),
	}
	// Set some demand orders
	demandOrders := []types.DemandOrder{
		{
			Id:                   "1",
			TrackingPacketKey:    "key",
			Price:                sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(100))),
			Fee:                  sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(10))),
			TrackingPacketStatus: commontypes.Status_PENDING,
		},
		{
			Id:                   "2",
			TrackingPacketKey:    "key2",
			Price:                sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(100))),
			Fee:                  sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(10))),
			TrackingPacketStatus: commontypes.Status_PENDING,
		},
	}
	for _, demandOrder := range demandOrders {
		demandOrderCopy := demandOrder
		err := k.SetDemandOrder(ctx, &demandOrderCopy)
		require.NoError(t, err)
	}
	k.SetParams(ctx, params)
	// Verify the exported genesis
	got := eibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.ElementsMatch(t, demandOrders, got.DemandOrders)
	require.Equal(t, params, got.Params)
}
