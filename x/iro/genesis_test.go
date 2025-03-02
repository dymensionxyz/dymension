package iro_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

var (
	fooCoin              = sdk.NewCoin("foo", math.NewInt(100))
	defaultCurve         = types.DefaultBondingCurve()
	defaultIncentives    = types.DefaultIncentivePlanParams()
	defaultLiquidityPart = types.DefaultParams().MinLiquidityPart
	defaultDuration      = time.Hour
)

var plans = []types.Plan{
	types.NewPlan(1, "rollapp1", fooCoin, defaultCurve, time.Time{}, time.Time{}, defaultIncentives, defaultLiquidityPart, defaultDuration),
	types.NewPlan(2, "rollapp2", fooCoin, defaultCurve, time.Time{}, time.Time{}, defaultIncentives, defaultLiquidityPart, defaultDuration),
}

func TestGenesis(t *testing.T) {
	t.Skip("skipped as it requires working auth keeper")

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		Plans:  plans,
	}

	k, ctx := keepertest.IROKeeper(t)
	iro.InitGenesis(ctx, *k, genesisState)

	// assertions
	require.Len(t, k.GetAllPlans(ctx, false), 2)
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
