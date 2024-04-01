package eibc_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name         string
		params       types.Params
		demandOrders []types.DemandOrder
		expPanic     bool
	}{
		{
			name: "only params - success",
			params: types.Params{
				EpochIdentifier: "week",
				TimeoutFee:      sdk.NewDecWithPrec(4, 1),
				ErrackFee:       sdk.NewDecWithPrec(4, 1),
			},
			demandOrders: []types.DemandOrder{},
			expPanic:     false,
		},
		{
			name: "params and demand order - panic",
			params: types.Params{
				EpochIdentifier: "week",
				TimeoutFee:      sdk.NewDecWithPrec(4, 1),
				ErrackFee:       sdk.NewDecWithPrec(4, 1),
			},
			demandOrders: []types.DemandOrder{{Id: "0"}},
			expPanic:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.GenesisState{Params: tt.params, DemandOrders: tt.demandOrders}
			k, ctx := keepertest.EibcKeeper(t)
			if tt.expPanic {
				require.Panics(t, func() {
					eibc.InitGenesis(ctx, *k, genesisState)
				})
			} else {
				eibc.InitGenesis(ctx, *k, genesisState)
				params := k.GetParams(ctx)
				require.Equal(t, genesisState.Params, params)
			}
		})
	}
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
