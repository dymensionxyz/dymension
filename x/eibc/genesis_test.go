package eibc_test

import (
	"encoding/base64"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/nullify"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
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
				TrackingPacketStatus: commontypes.Status_PENDING,
			},
			{
				Id:                   "2",
				TrackingPacketKey:    "22/33/44",
				Price:                sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(250)}},
				Fee:                  sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(150)}},
				Recipient:            "dym15saxgqw6kvhv6k5sg6r45kmdf4sf88kfw2adcw",
				FulfillerAddress:     "dym19pas0pqwje540u5ptwnffjxeamdxc9tajmdrfa",
				TrackingPacketStatus: commontypes.Status_FINALIZED,
			},
		},
	}

	k, ctx := keepertest.EIBCKeeper(t)
	eibc.InitGenesis(ctx, *k, genesisState)
	got := eibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.ElementsMatch(t, genesisState.DemandOrders, got.DemandOrders)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.DemandOrders, got.DemandOrders)
}

func TestExportGenesis(t *testing.T) {
	k, ctx := keepertest.EIBCKeeper(t)
	params := types.Params{
		EpochIdentifier: "week",
		TimeoutFee:      math.LegacyNewDecWithPrec(4, 1),
		ErrackFee:       math.LegacyNewDecWithPrec(4, 1),
	}
	// Set some demand orders
	demandOrders := []types.DemandOrder{
		{
			Id:                   "1",
			TrackingPacketKey:    "key",
			Price:                sdk.NewCoins(sdk.NewCoin("dym", math.NewInt(100))),
			Fee:                  sdk.NewCoins(sdk.NewCoin("dym", math.NewInt(10))),
			TrackingPacketStatus: commontypes.Status_PENDING,
		},
		{
			Id:                   "2",
			TrackingPacketKey:    "key2",
			Price:                sdk.NewCoins(sdk.NewCoin("dym", math.NewInt(100))),
			Fee:                  sdk.NewCoins(sdk.NewCoin("dym", math.NewInt(10))),
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

	require.NotNil(t, got, "ExportGenesis should not return nil")

	require.Equal(t, params, got.Params, "Params should match the set params")

	expectedDemandOrders := make([]types.DemandOrder, len(demandOrders))
	for i, order := range demandOrders {
		orderCopy := order
		encodedKey := base64.StdEncoding.EncodeToString([]byte(order.TrackingPacketKey))
		orderCopy.TrackingPacketKey = encodedKey
		expectedDemandOrders[i] = orderCopy
	}

	require.ElementsMatch(t, expectedDemandOrders, got.DemandOrders, "DemandOrders should match after encoding TrackingPacketKey")
}
