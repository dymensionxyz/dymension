package keeper_test

import (
	"sort"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func Test_epochHooks_BeforeEpochStart(t *testing.T) {
	now := time.Now().UTC()
	const daysKeepHistorical = 1
	require.Greater(t, daysKeepHistorical, 0, "mis-configured test case")

	setupTest := func() (dymnskeeper.Keeper, sdk.Context) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		params := dk.GetParams(ctx)
		params.Misc.PreservedClosedSellOrderDuration = daysKeepHistorical * 24 * time.Hour
		err := dk.SetParams(ctx, params)
		require.NoError(t, err)

		return dk, ctx
	}

	t.Run("should do something even nothing to do", func(t *testing.T) {
		dk, ctx := setupTest()

		params := dk.GetParams(ctx)

		originalGas := ctx.GasMeter().GasConsumed()

		err := dk.GetEpochHooks().BeforeEpochStart(ctx, params.Misc.BeginEpochHookIdentifier, 1)
		require.NoError(t, err)

		// gas should be changed because it should at least reading the params to check epoch identifier
		require.Less(t, originalGas, ctx.GasMeter().GasConsumed(), "should do something")
	})

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Add(365 * 24 * time.Hour).Unix(),
	}

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix(),
	}

	dymNameC := dymnstypes.DymName{
		Name:       "c",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Add(-365 * 24 * time.Hour).Unix(),
	}

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   1,
	}

	getEpochWithOffset := func(offset int64) int64 {
		return now.Unix() + offset
	}
	genSo := func(
		dymName dymnstypes.DymName, offsetExpiry int64,
	) dymnstypes.SellOrder {
		return dymnstypes.SellOrder{
			Name:     dymName.Name,
			ExpireAt: getEpochWithOffset(offsetExpiry),
			MinPrice: dymnsutils.TestCoin(100),
		}
	}

	type testSuite struct {
		t   *testing.T
		dk  dymnskeeper.Keeper
		ctx sdk.Context
	}

	nts := func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) testSuite {
		return testSuite{
			t:   t,
			dk:  dk,
			ctx: ctx,
		}
	}

	requireDymNameNotChanged := func(dymName dymnstypes.DymName, ts testSuite) {
		laterDymName := ts.dk.GetDymName(ts.ctx, dymName.Name)
		require.NotNil(t, laterDymName)

		require.Equal(t, dymName, *laterDymName, "nothing changed")
	}

	requireNoActiveSO := func(dymName dymnstypes.DymName, ts testSuite) {
		so := ts.dk.GetSellOrder(ts.ctx, dymName.Name)
		require.Nil(t, so)
	}

	requireActiveSO := func(dymName dymnstypes.DymName, ts testSuite) {
		so := ts.dk.GetSellOrder(ts.ctx, dymName.Name)
		require.NotNil(t, so)
	}

	requireHistoricalSOs := func(dymName dymnstypes.DymName, wantCount int, ts testSuite) {
		historicalSOs := ts.dk.GetHistoricalSellOrders(ts.ctx, dymName.Name)
		require.Lenf(t, historicalSOs, wantCount, "should have %d historical SOs", wantCount)
	}

	testsCleanupHistoricalSellOrders := []struct {
		name                           string
		dymNames                       []dymnstypes.DymName
		historicalSOs                  []dymnstypes.SellOrder
		activeSOs                      []dymnstypes.SellOrder
		minExpiryByDymName             map[string]int64
		customEpochIdentifier          string
		wantErr                        bool
		wantErrContains                string
		wantMinExpiryPerDymNameRecords []dymnstypes.HistoricalSellOrderMinExpiry
		preHookTestFunc                func(*testing.T, dymnskeeper.Keeper, sdk.Context)
		afterHookTestFunc              func(*testing.T, dymnskeeper.Keeper, sdk.Context)
	}{
		{
			name:     "simple cleanup",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -daysKeepHistorical*86400-1),
			},
			activeSOs: nil,
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-daysKeepHistorical*86400 - 1),
			},
			wantErr:                        false,
			wantMinExpiryPerDymNameRecords: nil,
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 1, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 0, ts)

				requireDymNameNotChanged(dymNameB, ts)
				requireHistoricalSOs(dymNameB, 0, ts)

				requireDymNameNotChanged(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 0, ts)

				requireDymNameNotChanged(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
		},
		{
			name:     "mis-match epoch will clean nothing",
			dymNames: []dymnstypes.DymName{dymNameA},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -daysKeepHistorical*86400-1),
			},
			activeSOs: nil,
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-daysKeepHistorical*86400 - 1),
			},
			customEpochIdentifier: "not-match",
			wantErr:               false,
			wantMinExpiryPerDymNameRecords: []dymnstypes.HistoricalSellOrderMinExpiry{
				{
					DymName:   dymNameA.Name,
					MinExpiry: getEpochWithOffset(-daysKeepHistorical*86400 - 1),
				},
			},
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 1, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 1, ts)
			},
		},
		{
			name:     "simple cleanup, with active SO",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -daysKeepHistorical*86400-1),
			},
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, +1),
			},
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-daysKeepHistorical*86400 - 1),
			},
			wantErr:                        false,
			wantMinExpiryPerDymNameRecords: nil,
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 1, ts)
				requireActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 0, ts)
				requireActiveSO(dymNameA, ts)

				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
		},
		{
			name:          "simple cleanup, no historical record to prune",
			dymNames:      []dymnstypes.DymName{dymNameA},
			historicalSOs: nil,
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, +1),
			},
			minExpiryByDymName:             nil,
			wantErr:                        false,
			wantMinExpiryPerDymNameRecords: nil,
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 0, ts)
				requireActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 0, ts)
				requireActiveSO(dymNameA, ts)
			},
		},
		{
			name:     "simple cleanup, nothing to prune",
			dymNames: []dymnstypes.DymName{dymNameA},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -1),
			},
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, +1),
			},
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-1),
			},
			wantErr: false,
			wantMinExpiryPerDymNameRecords: []dymnstypes.HistoricalSellOrderMinExpiry{
				{
					DymName:   dymNameA.Name,
					MinExpiry: getEpochWithOffset(-1),
				},
			},
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 1, ts)
				requireActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)
				requireActiveSO(dymNameA, ts)
			},
		},
		{
			name:     "cleanup multiple Historical SO, all need to prune",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -(daysKeepHistorical+0)*86400-1),
				genSo(dymNameA, -(daysKeepHistorical+2)*86400-1),
				genSo(dymNameA, -(daysKeepHistorical+1)*86400-1),
				genSo(dymNameC, -(daysKeepHistorical+3)*86400-1),
				genSo(dymNameC, -(daysKeepHistorical+5)*86400-1),
			},
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameC, +1),
			},
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-(daysKeepHistorical+2)*86400 - 1),
				dymNameC.Name: getEpochWithOffset(-(daysKeepHistorical+5)*86400 - 1),
			},
			wantErr:                        false,
			wantMinExpiryPerDymNameRecords: nil,
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 3, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 2, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 0, ts)
				requireNoActiveSO(dymNameA, ts)

				requireDymNameNotChanged(dymNameB, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireNoActiveSO(dymNameB, ts)

				requireDymNameNotChanged(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireActiveSO(dymNameC, ts)

				requireDymNameNotChanged(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
				requireNoActiveSO(dymNameD, ts)
			},
		},
		{
			name:     "cleanup multiple Historical SO, some need to prune while some not",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -(daysKeepHistorical+0)*86400-1),
				genSo(dymNameA, -(daysKeepHistorical+2)*86400-1),
				genSo(dymNameA, -9),
				genSo(dymNameC, -(daysKeepHistorical+3)*86400-1),
				genSo(dymNameC, -(daysKeepHistorical+5)*86400-1),
			},
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, +1), genSo(dymNameB, +1), genSo(dymNameC, +1),
			},
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-(daysKeepHistorical+2)*86400 - 1),
				dymNameC.Name: getEpochWithOffset(-(daysKeepHistorical+5)*86400 - 1),
			},
			wantErr: false,
			wantMinExpiryPerDymNameRecords: []dymnstypes.HistoricalSellOrderMinExpiry{
				{
					DymName:   dymNameA.Name,
					MinExpiry: getEpochWithOffset(-9),
				},
			},
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 3, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireHistoricalSOs(dymNameC, 2, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)
				requireActiveSO(dymNameA, ts)

				requireDymNameNotChanged(dymNameB, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireActiveSO(dymNameB, ts)

				requireDymNameNotChanged(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireActiveSO(dymNameC, ts)

				requireDymNameNotChanged(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
				requireNoActiveSO(dymNameD, ts)
			},
		},
		{
			name:     "should update min expiry correctly",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, -9),
				genSo(dymNameA, -(daysKeepHistorical+2)*86400-1),
				genSo(dymNameA, -10),
			},
			activeSOs: nil,
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-(daysKeepHistorical+2)*86400 - 1),
			},
			wantErr: false,
			wantMinExpiryPerDymNameRecords: []dymnstypes.HistoricalSellOrderMinExpiry{
				{
					DymName:   dymNameA.Name,
					MinExpiry: getEpochWithOffset(-10),
				},
			},
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 3, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 2, ts)
			},
		},
		{
			name:     "mixed cleanup",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				// Dym-Name A has some historical SO, some need to prune, some not
				genSo(dymNameA, -(daysKeepHistorical+0)*86400-1),
				genSo(dymNameA, -(daysKeepHistorical+2)*86400-1),
				genSo(dymNameA, -9),
				// Dym-Name B has some historical SO, no need to prune
				genSo(dymNameB, -8),
				genSo(dymNameB, -7),
				// Dym-Name C has some historical SO, all need to prune
				genSo(dymNameC, -(daysKeepHistorical+3)*86400-1),
				genSo(dymNameC, -(daysKeepHistorical+5)*86400-1),
				// Dym-Name D has no historical SO
			},
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, +1), genSo(dymNameB, +1), genSo(dymNameC, +1), genSo(dymNameD, +1),
			},
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-(daysKeepHistorical+2)*86400 - 1),
				dymNameB.Name: getEpochWithOffset(-8),
				dymNameC.Name: getEpochWithOffset(-(daysKeepHistorical+5)*86400 - 1),
			},
			wantErr: false,
			wantMinExpiryPerDymNameRecords: []dymnstypes.HistoricalSellOrderMinExpiry{
				{
					DymName:   dymNameA.Name,
					MinExpiry: getEpochWithOffset(-9),
				},
				{
					DymName:   dymNameB.Name,
					MinExpiry: getEpochWithOffset(-8),
				},
			},
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 3, ts)
				requireHistoricalSOs(dymNameB, 2, ts)
				requireHistoricalSOs(dymNameC, 2, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)
				requireActiveSO(dymNameA, ts)

				requireDymNameNotChanged(dymNameB, ts)
				requireHistoricalSOs(dymNameB, 2, ts)
				requireActiveSO(dymNameB, ts)

				requireDymNameNotChanged(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireActiveSO(dymNameC, ts)

				requireDymNameNotChanged(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
				requireActiveSO(dymNameD, ts)
			},
		},
		{
			name:          "case no historical SO but has min expiry",
			dymNames:      []dymnstypes.DymName{dymNameA},
			historicalSOs: nil,
			activeSOs:     nil,
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: 1, // incorrect state: no historical SO but has min expiry
			},
			wantErr:                        false,
			wantMinExpiryPerDymNameRecords: nil,
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 0, ts)
			},
		},
		{
			name:     "mixed cleanup with incorrect state",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			historicalSOs: []dymnstypes.SellOrder{
				// Dym-Name A has some SO, no need to prune
				genSo(dymNameA, -9),
				// Dym-Name D has no historical SO
			},
			activeSOs: []dymnstypes.SellOrder{
				genSo(dymNameA, +1), genSo(dymNameB, +1), genSo(dymNameC, +1), genSo(dymNameD, +1),
			},
			minExpiryByDymName: map[string]int64{
				dymNameA.Name: getEpochWithOffset(-daysKeepHistorical*86400 - 1), // incorrect state: has historical SO, no need to prune but min-expiry indicates need to prune
				dymNameD.Name: 1,                                                 // incorrect state: no historical SO but has min expiry
			},
			wantErr: false,
			wantMinExpiryPerDymNameRecords: []dymnstypes.HistoricalSellOrderMinExpiry{
				{
					DymName:   dymNameA.Name,
					MinExpiry: getEpochWithOffset(-9), // corrected value
				},
				// incorrect of Dym-Name D was removed
			},
			preHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireHistoricalSOs(dymNameA, 1, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, ctx sdk.Context) {
				ts := nts(t, dk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)
				requireActiveSO(dymNameA, ts)

				requireDymNameNotChanged(dymNameB, ts)
				requireHistoricalSOs(dymNameB, 0, ts)
				requireActiveSO(dymNameB, ts)

				requireDymNameNotChanged(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 0, ts)
				requireActiveSO(dymNameC, ts)

				requireDymNameNotChanged(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 0, ts)
				requireActiveSO(dymNameD, ts)
			},
		},
	}
	for _, tt := range testsCleanupHistoricalSellOrders {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.preHookTestFunc, "mis-configured test case")
			require.NotNil(t, tt.afterHookTestFunc, "mis-configured test case")

			dk, ctx := setupTest()

			for _, dymName := range tt.dymNames {
				err := dk.SetDymName(ctx, dymName)
				require.NoError(t, err)
			}

			for _, so := range tt.historicalSOs {
				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)
				err = dk.MoveSellOrderToHistorical(ctx, so.Name)
				require.NoError(t, err)
			}

			for _, so := range tt.activeSOs {
				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)
			}

			meh := dk.GetMinExpiryOfAllHistoricalSellOrders(ctx)
			if len(meh) > 0 {
				// clear existing records to simulate cases of malformed state
				for _, record := range meh {
					dk.SetMinExpiryHistoricalSellOrder(ctx, record.DymName, 0)
				}
			}
			if len(tt.minExpiryByDymName) > 0 {
				for dymName, minExpiry := range tt.minExpiryByDymName {
					dk.SetMinExpiryHistoricalSellOrder(ctx, dymName, minExpiry)
				}
			}

			tt.preHookTestFunc(t, dk, ctx)

			moduleParams := dk.GetParams(ctx)
			useEpochIdentifier := moduleParams.Misc.BeginEpochHookIdentifier
			if tt.customEpochIdentifier != "" {
				useEpochIdentifier = tt.customEpochIdentifier
			}
			err := dk.GetEpochHooks().BeforeEpochStart(ctx, useEpochIdentifier, 1)

			defer func() {
				if t.Failed() {
					return
				}

				tt.afterHookTestFunc(t, dk, ctx)

				meh := dk.GetMinExpiryOfAllHistoricalSellOrders(ctx)
				if len(tt.wantMinExpiryPerDymNameRecords) == 0 {
					require.Empty(t, meh)
				} else {
					require.Equal(t, tt.wantMinExpiryPerDymNameRecords, meh, "lists mismatch")
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}

	testsClearPreservedRegistration := []struct {
		name                            string
		preservedRegistrationParams     dymnstypes.PreservedRegistrationParams
		wantPreservedRegistrationParams dymnstypes.PreservedRegistrationParams
	}{
		{
			name: "pass - can clear when expired",
			preservedRegistrationParams: dymnstypes.PreservedRegistrationParams{
				ExpirationEpoch: now.Add(-time.Second).Unix(),
				PreservedDymNames: []dymnstypes.PreservedDymName{
					{
						DymName:            "preserved",
						WhitelistedAddress: owner,
					},
				},
			},
			wantPreservedRegistrationParams: dymnstypes.PreservedRegistrationParams{
				ExpirationEpoch:   0,
				PreservedDymNames: nil,
			},
		},
		{
			name: "pass - keep when not expired",
			preservedRegistrationParams: dymnstypes.PreservedRegistrationParams{
				ExpirationEpoch: now.Add(time.Hour).Unix(),
				PreservedDymNames: []dymnstypes.PreservedDymName{
					{
						DymName:            "preserved",
						WhitelistedAddress: owner,
					},
				},
			},
			wantPreservedRegistrationParams: dymnstypes.PreservedRegistrationParams{
				ExpirationEpoch: now.Add(time.Hour).Unix(),
				PreservedDymNames: []dymnstypes.PreservedDymName{
					{
						DymName:            "preserved",
						WhitelistedAddress: owner,
					},
				},
			},
		},
	}
	for _, tt := range testsClearPreservedRegistration {
		t.Run(tt.name, func(t *testing.T) {
			dk, ctx := setupTest()

			moduleParams := dk.GetParams(ctx)
			moduleParams.PreservedRegistration = tt.preservedRegistrationParams
			require.NoError(t, dk.SetParams(ctx, moduleParams))

			err := dk.GetEpochHooks().BeforeEpochStart(ctx, moduleParams.Misc.BeginEpochHookIdentifier, 1)
			require.NoError(t, err)

			require.Equal(t, tt.wantPreservedRegistrationParams, dk.GetParams(ctx).PreservedRegistration)
		})
	}
}

//goland:noinspection SpellCheckingInspection
func Test_epochHooks_AfterEpochEnd(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, bk, ctx
	}

	t.Run("should do something even nothing to do", func(t *testing.T) {
		dk, _, ctx := setupTest()

		params := dk.GetParams(ctx)

		originalGas := ctx.GasMeter().GasConsumed()

		err := dk.GetEpochHooks().AfterEpochEnd(ctx, params.Misc.EndEpochHookIdentifier, 1)
		require.NoError(t, err)

		// gas should be changed because it should at least reading the params to check epoch identifier
		require.Less(t, originalGas, ctx.GasMeter().GasConsumed(), "should do something")
	})

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const bidder = "dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv"
	dymNsModuleAccAddr := authtypes.NewModuleAddress(dymnstypes.ModuleName)

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix() + 1,
	}

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix() + 1,
	}

	dymNameC := dymnstypes.DymName{
		Name:       "c",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix() + 1,
	}

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix() + 1,
	}

	coin100 := dymnsutils.TestCoin(100)
	coin200 := dymnsutils.TestCoin(200)
	denom := dymnsutils.TestCoin(0).Denom

	soExpiredEpoch := now.Unix() - 1
	soNotExpiredEpoch := now.Unix() + 1

	const soExpired = true
	const soNotExpired = false
	genSo := func(
		dymName dymnstypes.DymName,
		expired bool, sellPrice *sdk.Coin, highestBid *dymnstypes.SellOrderBid,
	) dymnstypes.SellOrder {
		return dymnstypes.SellOrder{
			Name: dymName.Name,
			ExpireAt: func() int64 {
				if expired {
					return soExpiredEpoch
				}
				return soNotExpiredEpoch
			}(),
			MinPrice:   coin100,
			SellPrice:  sellPrice,
			HighestBid: highestBid,
		}
	}

	type testSuite struct {
		t   *testing.T
		dk  dymnskeeper.Keeper
		bk  dymnskeeper.BankKeeper
		ctx sdk.Context
	}

	nts := func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) testSuite {
		return testSuite{
			t:   t,
			dk:  dk,
			bk:  bk,
			ctx: ctx,
		}
	}

	requireOwnerChanged := func(dymName dymnstypes.DymName, newOwner string, ts testSuite) {
		require.NotEmpty(t, newOwner, "mis-configured test case")

		laterDymName := ts.dk.GetDymName(ts.ctx, dymName.Name)
		require.NotNil(t, laterDymName)

		require.Equal(t, newOwner, laterDymName.Owner, "ownership must be transferred")
		require.Equal(t, newOwner, laterDymName.Controller, "controller must be changed")
		require.Equal(t, dymName.ExpireAt, laterDymName.ExpireAt, "expiry must not be changed")
		require.Empty(t, laterDymName.Configs, "configs must be cleared")
	}

	requireDymNameNotChanged := func(dymName dymnstypes.DymName, ts testSuite) {
		laterDymName := ts.dk.GetDymName(ts.ctx, dymName.Name)
		require.NotNil(t, laterDymName)

		require.Equal(t, dymName, *laterDymName, "nothing changed")
	}

	requireNoActiveSO := func(dymName dymnstypes.DymName, ts testSuite) {
		so := ts.dk.GetSellOrder(ts.ctx, dymName.Name)
		require.Nil(t, so)
	}

	requireHistoricalSOs := func(dymName dymnstypes.DymName, wantCount int, ts testSuite) {
		historicalSOs := ts.dk.GetHistoricalSellOrders(ts.ctx, dymName.Name)
		require.Lenf(t, historicalSOs, wantCount, "should have %d historical SOs", wantCount)
	}

	requireModuleBalance := func(wantAmount int64, ts testSuite) {
		moduleBalance := ts.bk.GetBalance(ts.ctx, dymNsModuleAccAddr, denom)
		require.NotNil(t, moduleBalance)

		require.Equalf(t, wantAmount, moduleBalance.Amount.Int64(), "module balance should be %d", wantAmount)
	}

	requireAccountBalance := func(bech32Addr string, wantAmount int64, ts testSuite) {
		accountBalance := ts.bk.GetBalance(ts.ctx, sdk.MustAccAddressFromBech32(bech32Addr), denom)
		require.NotNil(t, accountBalance)

		require.Equalf(t, wantAmount, accountBalance.Amount.Int64(), "account balance should be %d", wantAmount)
	}

	requireConfiguredAddressMappedDymNames := func(ts testSuite, bech32Addr string, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsConfiguredAddress(ts.ctx, bech32Addr)
		require.NoError(ts.t, err)
		require.Len(ts.t, dymNames, len(names))
		sort.Strings(names)
		sort.Slice(dymNames, func(i, j int) bool {
			return dymNames[i].Name < dymNames[j].Name
		})
		for i, name := range names {
			require.Equal(ts.t, name, dymNames[i].Name)
		}
	}

	requireConfiguredAddressMappedNoDymName := func(ts testSuite, bech32Addr string) {
		requireConfiguredAddressMappedDymNames(ts, bech32Addr)
	}

	require0xMappedDymNames := func(ts testSuite, bech32Addr string, names ...string) {
		_, bz, err := bech32.DecodeAndConvert(bech32Addr)
		require.NoError(ts.t, err)

		dymNames, err := ts.dk.GetDymNamesContainsHexAddress(ts.ctx, bz)
		require.NoError(ts.t, err)
		require.Len(ts.t, dymNames, len(names))
		sort.Strings(names)
		sort.Slice(dymNames, func(i, j int) bool {
			return dymNames[i].Name < dymNames[j].Name
		})
		for i, name := range names {
			require.Equal(ts.t, name, dymNames[i].Name)
		}
	}

	require0xMappedNoDymName := func(ts testSuite, bech32Addr string) {
		require0xMappedDymNames(ts, bech32Addr)
	}

	tests := []struct {
		name                  string
		dymNames              []dymnstypes.DymName
		sellOrders            []dymnstypes.SellOrder
		expiryByDymName       []dymnstypes.ActiveSellOrdersExpirationRecord
		preMintModuleBalance  int64
		customEpochIdentifier string
		beforeHookTestFunc    func(*testing.T, dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context)
		wantErr               bool
		wantErrContains       string
		wantExpiryByDymName   []dymnstypes.ActiveSellOrdersExpirationRecord
		afterHookTestFunc     func(*testing.T, dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context)
	}{
		{
			name:       "simple process expired SO",
			dymNames:   []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, nil)},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				requireModuleBalance(200, ts)

				requireAccountBalance(dymNameA.Owner, 0, ts)

				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				require0xMappedNoDymName(ts, bidder)
			},
		},
		{
			name:     "simple process expired & completed SO",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidder,
				Price:  coin200,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				requireOwnerChanged(dymNameA, bidder, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				requireModuleBalance(0, ts) // should be transferred to previous owner

				requireAccountBalance(dymNameA.Owner, 200, ts) // previous owner should earn from bid

				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, bidder, dymNameA.Name)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, bidder, dymNameA.Name)
			},
		},
		{
			name:     "simple process expired & completed SO, match by min price",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidder,
				Price:  coin100,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 250,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				requireOwnerChanged(dymNameA, bidder, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				requireModuleBalance(150, ts) // 100 should be transferred to previous owner

				requireAccountBalance(dymNameA.Owner, 100, ts) // previous owner should earn from bid

				requireConfiguredAddressMappedNoDymName(ts, owner)
				requireConfiguredAddressMappedDymNames(ts, bidder, dymNameA.Name)
				require0xMappedNoDymName(ts, owner)
				require0xMappedDymNames(ts, bidder, dymNameA.Name)
			},
		},
		{
			name:     "process multiple - mixed SOs",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, &coin200, &dymnstypes.SellOrderBid{
					// not completed
					Bidder: bidder,
					Price:  coin100,
				}),
				genSo(dymNameC, soExpired, &coin200, &dymnstypes.SellOrderBid{
					Bidder: bidder,
					Price:  coin200,
				}),
				genSo(dymNameD, soExpired, &coin200, &dymnstypes.SellOrderBid{
					// completed by min price
					Bidder: bidder,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					Name:     dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					Name:     dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					Name:     dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 450,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				// SO for Dym-Name A is expired without any bid/winner
				requireDymNameNotChanged(dymNameA, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				// SO for Dym-Name B not yet finished
				requireDymNameNotChanged(dymNameB, ts)
				soB := ts.dk.GetSellOrder(ts.ctx, dymNameB.Name)
				require.NotNil(t, soB)
				requireHistoricalSOs(dymNameB, 0, ts)

				// SO for Dym-Name C is completed with winner
				requireOwnerChanged(dymNameC, bidder, ts)
				requireNoActiveSO(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 1, ts)

				// SO for Dym-Name D is completed with winner
				requireOwnerChanged(dymNameD, bidder, ts)
				requireNoActiveSO(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 1, ts)

				requireModuleBalance(150, ts)

				requireAccountBalance(owner, 300, ts) // price from 2 completed SO

				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedDymNames(ts, bidder, dymNameC.Name, dymNameD.Name)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				require0xMappedDymNames(ts, bidder, dymNameC.Name, dymNameD.Name)
			},
		},
		{
			name:     "should do nothing if invalid epoch identifier",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, &coin200, &dymnstypes.SellOrderBid{
					// not completed
					Bidder: bidder,
					Price:  coin100,
				}),
				genSo(dymNameC, soExpired, &coin200, &dymnstypes.SellOrderBid{
					Bidder: bidder,
					Price:  coin200,
				}),
				genSo(dymNameD, soExpired, &coin200, &dymnstypes.SellOrderBid{
					// completed by min price
					Bidder: bidder,
					Price:  coin100,
				}),
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					Name:     dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					Name:     dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					Name:     dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance:  450,
			customEpochIdentifier: "another",
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					Name:     dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					Name:     dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					Name:     dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				requireDymNameNotChanged(dymNameA, ts)
				requireDymNameNotChanged(dymNameB, ts)
				requireDymNameNotChanged(dymNameC, ts)
				requireDymNameNotChanged(dymNameD, ts)

				requireModuleBalance(450, ts)

				requireAccountBalance(owner, 0, ts)

				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				require0xMappedNoDymName(ts, bidder)
			},
		},
		{
			name:     "should remove expiry reference to non-exists SO",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				// no SO for Dym-Name B
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					// no SO for Dym-Name B but still have reference
					Name:     dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr:             false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to Dym-Name A because of processed
				// removed reference to Dym-Name B because SO not exists
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				require0xMappedNoDymName(ts, bidder)
			},
		},
		{
			name:     "update expiry if in-correct",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, nil, nil), // SO not expired
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					// incorrect, SO not expired
					Name:     dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				require0xMappedNoDymName(ts, bidder)
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidder)
				require0xMappedDymNames(ts, owner, dymNameA.Name, dymNameB.Name)
				require0xMappedNoDymName(ts, bidder)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.beforeHookTestFunc, "mis-configured test case")
			require.NotNil(t, tt.afterHookTestFunc, "mis-configured test case")

			dk, bk, ctx := setupTest()

			if tt.preMintModuleBalance > 0 {
				err := bk.MintCoins(ctx, dymnstypes.ModuleName, dymnsutils.TestCoins(tt.preMintModuleBalance))
				require.NoError(t, err)
			}

			err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: tt.expiryByDymName,
			})
			require.NoError(t, err)

			for _, dymName := range tt.dymNames {
				setDymNameWithFunctionsAfter(ctx, dymName, t, dk)
			}

			for _, so := range tt.sellOrders {
				err = dk.SetSellOrder(ctx, so)
				require.NoError(t, err)
			}

			moduleParams := dk.GetParams(ctx)

			useEpochIdentifier := moduleParams.Misc.EndEpochHookIdentifier
			if tt.customEpochIdentifier != "" {
				useEpochIdentifier = tt.customEpochIdentifier
			}

			tt.beforeHookTestFunc(t, dk, bk, ctx)

			err = dk.GetEpochHooks().AfterEpochEnd(ctx, useEpochIdentifier, 1)

			defer func() {
				if t.Failed() {
					return
				}

				tt.afterHookTestFunc(t, dk, bk, ctx)

				aope := dk.GetActiveSellOrdersExpiration(ctx)
				if len(tt.wantExpiryByDymName) == 0 {
					require.Empty(t, aope.Records)
				} else {
					require.Equal(t, tt.wantExpiryByDymName, aope.Records)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				return
			}

			require.NoError(t, err)
		})
	}
}
