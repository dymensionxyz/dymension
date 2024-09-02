package iro_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/stretchr/testify/require"
)

var (
	amt   = sdk.NewCoin("foo", math.NewInt(100))
	plans = []types.Plan{
		types.NewPlan(1, "rollapp1", amt, types.DefaultBondingCurve(), time.Time{}, time.Time{}),
		types.NewPlan(2, "rollapp2", amt, types.DefaultBondingCurve(), time.Time{}, time.Time{}),
	}
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		Plans:  plans,
	}

	k, ctx := keepertest.IROKeeper(t)
	iro.InitGenesis(ctx, *k, genesisState)

	// assertions
	require.Len(t, k.GetAllPlans(ctx), 2)
	_, found := k.GetPlanByRollapp(ctx, "rollapp1")
	require.True(t, found)
	lastPlanId := k.GetLastPlanId(ctx)
	require.Equal(t, uint64(2), lastPlanId)

	got := iro.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	for i := range genesisState.Plans {
		require.Equal(t, genesisState.Plans[i], got.Plans[i])
	}
}
