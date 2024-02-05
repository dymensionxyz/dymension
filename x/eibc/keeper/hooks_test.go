package keeper_test

import (
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	types "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/stretchr/testify/require"
)

// TestAfterEpochEnd tests that the finalized of demand orders
// are deleted given the correct epoch identifier
func TestAfterEpochEnd(t *testing.T) {
	tests := []struct {
		name                 string
		pendingOrdersNum     int
		finalizedOrdersNum   int
		epochIdentifierParam string
		epochIdentifier      string
		expectedDeleted      int
		expectedTotal        int
	}{
		{
			name:                 "epoch identifier matches params set",
			pendingOrdersNum:     5,
			finalizedOrdersNum:   3,
			epochIdentifierParam: "minute",
			epochIdentifier:      "minute",
			expectedDeleted:      3,
			expectedTotal:        2,
		},
		{
			name:                 "epoch identifer does not match params set",
			pendingOrdersNum:     5,
			finalizedOrdersNum:   3,
			epochIdentifierParam: "minute",
			epochIdentifier:      "hour",
			expectedDeleted:      0,
			expectedTotal:        5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx := keepertest.EibcKeeper(t)
			for i := 1; i <= tc.pendingOrdersNum; i++ {
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

			demandOrders := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING)
			require.Equal(t, tc.pendingOrdersNum, len(demandOrders))

			for _, demandOrder := range demandOrders[:tc.finalizedOrdersNum] {
				keeper.UpdateDemandOrderWithStatus(ctx, demandOrder, commontypes.Status_FINALIZED)
			}
			finzliedDemandOrders := keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED)
			require.Equal(t, tc.finalizedOrdersNum, len(finzliedDemandOrders))

			keeper.SetParams(ctx, types.Params{EpochIdentifier: tc.epochIdentifierParam})
			epochHooks := keeper.GetEpochHooks()
			epochHooks.AfterEpochEnd(ctx, tc.epochIdentifier, 1)

			finzliedDemandOrders = keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_FINALIZED)
			require.Equal(t, tc.finalizedOrdersNum-tc.expectedDeleted, len(finzliedDemandOrders))

			totalDemandOrders := len(finzliedDemandOrders) + len(keeper.ListDemandOrdersByStatus(ctx, commontypes.Status_PENDING))
			require.Equal(t, tc.expectedTotal, totalDemandOrders)
		})
	}
}
