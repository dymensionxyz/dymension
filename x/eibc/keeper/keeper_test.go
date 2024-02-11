package keeper_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/stretchr/testify/require"
)

func TestListDemandOrdersByStatus(t *testing.T) {
	keeper, ctx := keepertest.EibcKeeper(t)
	demandOrdersNum := 5
	// Create and set some demand orders with status pending
	for i := 1; i <= demandOrdersNum; i++ {
		demandOrder := &types.DemandOrder{
			Id:                   strconv.Itoa(i),
			TrackingPacketKey:    "testTrackingPacketKey",
			Price:                sdktypes.Coins{sdktypes.Coin{Denom: "adym", Amount: math.NewInt(100)}},
			Fee:                  sdktypes.Coins{sdktypes.Coin{Denom: "adym", Amount: math.NewInt(10)}},
			Recipient:            "dym1zp455m6ukuq5k9kzazjpfachf6rv2ej6rcp6v8",
			IsFullfilled:         false,
			TrackingPacketStatus: commontypes.Status_PENDING,
		}
		keeper.SetDemandOrder(ctx, demandOrder)
	}

	// Get the demand orders with status active
	demandOrders, err := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING)
	require.NoError(t, err)
	require.Equal(t, demandOrdersNum, len(demandOrders))

	// Update 3 of the demand orders to status finalized
	for _, demandOrder := range demandOrders[:3] {
		keeper.UpdateDemandOrderWithStatus(ctx, demandOrder, commontypes.Status_FINALIZED)
	}
	// Retrieve the updated demand orders after status change
	updatedDemandOrders, err := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED)
	require.NoError(t, err)
	// Validate that there are exactly demandOrderNum packets in total
	pendingDemandOrders, err := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING)
	require.NoError(t, err)
	totalDemandOrders := len(updatedDemandOrders) + len(pendingDemandOrders)
	require.Equal(t, demandOrdersNum, totalDemandOrders)
}
