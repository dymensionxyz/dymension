package keeper_test

import (
	"sort"
	"testing"
	"time"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

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

	ownerA := testAddr(1).bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Add(365 * 24 * time.Hour).Unix(),
	}

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix(),
	}

	dymNameC := dymnstypes.DymName{
		Name:       "c",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Add(-365 * 24 * time.Hour).Unix(),
	}

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      ownerA,
		Controller: ownerA,
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
			Type:     dymnstypes.MarketOrderType_MOT_DYM_NAME,
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

	// TODO DymNS: add test cleanup Alias SO

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
						WhitelistedAddress: ownerA,
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
						WhitelistedAddress: ownerA,
					},
				},
			},
			wantPreservedRegistrationParams: dymnstypes.PreservedRegistrationParams{
				ExpirationEpoch: now.Add(time.Hour).Unix(),
				PreservedDymNames: []dymnstypes.PreservedDymName{
					{
						DymName:            "preserved",
						WhitelistedAddress: ownerA,
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

	ownerAcc := testAddr(1)
	ownerA := ownerAcc.bech32()

	bidderAcc := testAddr(2)
	bidderA := bidderAcc.bech32()

	dymNameA := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}

	dymNameB := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}

	dymNameC := dymnstypes.DymName{
		Name:       "c",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}

	dymNameD := dymnstypes.DymName{
		Name:       "d",
		Owner:      ownerA,
		Controller: ownerA,
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
			Type: dymnstypes.MarketOrderType_MOT_DYM_NAME,
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

	requireConfiguredAddressMappedDymNames := func(ts testSuite, cfgAddr string, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsConfiguredAddress(ts.ctx, cfgAddr)
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

	requireConfiguredAddressMappedNoDymName := func(ts testSuite, cfgAddr string) {
		requireConfiguredAddressMappedDymNames(ts, cfgAddr)
	}

	requireFallbackAddrMappedDymNames := func(ts testSuite, fallbackAddr dymnstypes.FallbackAddress, names ...string) {
		dymNames, err := ts.dk.GetDymNamesContainsFallbackAddress(ts.ctx, fallbackAddr)
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

	requireFallbackAddrMappedNoDymName := func(ts testSuite, fallbackAddr dymnstypes.FallbackAddress) {
		requireFallbackAddrMappedDymNames(ts, fallbackAddr)
	}

	// TODO DymNS: add test for Alias SOs

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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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

				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
		},
		{
			name:     "simple process expired & completed SO",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				requireOwnerChanged(dymNameA, bidderA, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				requireModuleBalance(0, ts) // should be transferred to previous owner

				requireAccountBalance(dymNameA.Owner, 200, ts) // previous owner should earn from bid

				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, bidderA, dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddrMappedDymNames(ts, bidderAcc.fallback(), dymNameA.Name)
			},
		},
		{
			name:     "simple process expired & completed SO, match by min price",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				requireOwnerChanged(dymNameA, bidderA, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				requireModuleBalance(150, ts) // 100 should be transferred to previous owner

				requireAccountBalance(dymNameA.Owner, 100, ts) // previous owner should earn from bid

				requireConfiguredAddressMappedNoDymName(ts, ownerA)
				requireConfiguredAddressMappedDymNames(ts, bidderA, dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, ownerAcc.fallback())
				requireFallbackAddrMappedDymNames(ts, bidderAcc.fallback(), dymNameA.Name)
			},
		},
		{
			name:     "process multiple - mixed SOs",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, &coin200, &dymnstypes.SellOrderBid{
					// not completed
					Bidder: bidderA,
					Price:  coin100,
				}),
				genSo(dymNameC, soExpired, &coin200, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				}),
				genSo(dymNameD, soExpired, &coin200, &dymnstypes.SellOrderBid{
					// completed by min price
					Bidder: bidderA,
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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
				requireOwnerChanged(dymNameC, bidderA, ts)
				requireNoActiveSO(dymNameC, ts)
				requireHistoricalSOs(dymNameC, 1, ts)

				// SO for Dym-Name D is completed with winner
				requireOwnerChanged(dymNameD, bidderA, ts)
				requireNoActiveSO(dymNameD, ts)
				requireHistoricalSOs(dymNameD, 1, ts)

				requireModuleBalance(150, ts)

				requireAccountBalance(ownerA, 300, ts) // price from 2 completed SO

				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedDymNames(ts, bidderA, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedDymNames(ts, bidderAcc.fallback(), dymNameC.Name, dymNameD.Name)
			},
		},
		{
			name:     "should do nothing if invalid epoch identifier",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, &coin200, &dymnstypes.SellOrderBid{
					// not completed
					Bidder: bidderA,
					Price:  coin100,
				}),
				genSo(dymNameC, soExpired, &coin200, &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				}),
				genSo(dymNameD, soExpired, &coin200, &dymnstypes.SellOrderBid{
					// completed by min price
					Bidder: bidderA,
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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

				requireAccountBalance(ownerA, 0, ts)

				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:             false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to Dym-Name A because of processed
				// removed reference to Dym-Name B because SO not exists
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnskeeper.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
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

				aSoe := dk.GetActiveSellOrdersExpiration(ctx)
				if len(tt.wantExpiryByDymName) == 0 {
					require.Empty(t, aSoe.Records)
				} else {
					require.Equal(t, tt.wantExpiryByDymName, aSoe.Records)
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

func Test_rollappHooks_RollappCreated(t *testing.T) {
	now := time.Now().UTC()

	const price1L = 9
	const price2L = 8
	const price3L = 7
	const price4L = 6
	const price5L = 5
	const price6L = 4
	const price7PL = 3

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, rollappkeeper.Keeper, sdk.Context) {
		dk, bk, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		moduleParams.Price.AliasPrice_1Letter = sdk.NewInt(price1L)
		moduleParams.Price.AliasPrice_2Letters = sdk.NewInt(price2L)
		moduleParams.Price.AliasPrice_3Letters = sdk.NewInt(price3L)
		moduleParams.Price.AliasPrice_4Letters = sdk.NewInt(price4L)
		moduleParams.Price.AliasPrice_5Letters = sdk.NewInt(price5L)
		moduleParams.Price.AliasPrice_6Letters = sdk.NewInt(price6L)
		moduleParams.Price.AliasPrice_7PlusLetters = sdk.NewInt(price7PL)
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, rk, ctx
	}

	creatorAccAddr := sdk.AccAddress(testAddr(1).bytes())
	dymNameOwnerAcc := testAddr(2)
	anotherAcc := testAddr(3)

	tests := []struct {
		name                    string
		addRollApps             []string
		preRunSetup             func(*testing.T, sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
		originalCreatorBalance  int64
		originalModuleBalance   int64
		rollAppId               string
		alias                   string
		wantErr                 bool
		wantErrContains         string
		wantSuccess             bool
		wantLaterCreatorBalance int64
		postTest                func(*testing.T, sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
	}{
		{
			name:                    "pass - register without problem",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
		},
		{
			name:                   "pass - mapping RollApp ID <=> Alias should be set",
			addRollApps:            []string{"rollapp_1-1"},
			originalCreatorBalance: price5L,
			rollAppId:              "rollapp_1-1",
			alias:                  "alias",
			wantErr:                false,
			wantSuccess:            true,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				alias, found := dk.GetAliasByRollAppId(context, "rollapp_1-1")
				require.True(t, found)
				require.Equal(t, "alias", alias)

				rollAppId, found := dk.GetRollAppIdByAlias(context, "alias")
				require.True(t, found)
				require.Equal(t, "rollapp_1-1", rollAppId)
			},
		},
		{
			name:                    "pass - Alias cost subtracted from creator and burned",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 10,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 1 char",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price1L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "a",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 2 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price2L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "ab",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 3 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price3L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "dog",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 4 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price4L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "pool",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 5 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price5L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "angel",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 6 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price6L + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 7 chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price7PL + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "academy",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "pass - charges correct price for Alias based on length, 7+ chars",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price7PL + 10,
			rollAppId:               "rollapp_1-1",
			alias:                   "dymension",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 10,
		},
		{
			name:                    "fail - RollApp not exists",
			addRollApps:             nil,
			originalCreatorBalance:  price1L,
			rollAppId:               "nad_0-0",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "not a RollApp chain-id",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
		},
		{
			name:                    "fail - mapping should not be created",
			addRollApps:             nil,
			originalCreatorBalance:  price1L,
			rollAppId:               "nad_0-0",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "not",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				_, found := dk.GetAliasByRollAppId(ctx, "nad_0-0")
				require.False(t, found)

				_, found = dk.GetRollAppIdByAlias(ctx, "alias")
				require.False(t, found)
			},
		},
		{
			name:                    "fail - reject bad alias",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  price1L,
			rollAppId:               "rollapp_1-1",
			alias:                   "@@@",
			wantErr:                 true,
			wantErrContains:         "alias candidate: invalid argument",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
		},
		{
			name:        "pass - can register if alias is not used",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)

				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "dymension_1100-1",
						Aliases: []string{"dym"},
					},
				}

				require.NoError(t, dk.SetParams(ctx, moduleParams))

				err := dk.SetAliasForRollAppId(ctx, "rollapp_2-2", "ra2")
				require.NoError(t, err)
			},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				alias, found := dk.GetAliasByRollAppId(context, "rollapp_1-1")
				require.True(t, found)
				require.Equal(t, "alias", alias)

				rollAppId, found := dk.GetRollAppIdByAlias(context, "alias")
				require.True(t, found)
				require.Equal(t, "rollapp_1-1", rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)

				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "bridge",
						Aliases: []string{"b"},
					},
				}

				require.NoError(t, dk.SetParams(ctx, moduleParams))

				err := dk.SetAliasForRollAppId(ctx, "rollapp_2-2", "ra2")
				require.NoError(t, err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				alias, found := dk.GetAliasByRollAppId(context, "rollapp_1-1")
				require.False(t, found)
				require.Empty(t, alias)

				rollAppId, found := dk.GetRollAppIdByAlias(context, "bridge")
				require.False(t, found)
				require.Empty(t, rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as alias of a chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)

				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "dymension_1100-1",
						Aliases: []string{"dym"},
					},
				}

				require.NoError(t, dk.SetParams(ctx, moduleParams))

				err := dk.SetAliasForRollAppId(ctx, "rollapp_2-2", "ra2")
				require.NoError(t, err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "dym",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				alias, found := dk.GetAliasByRollAppId(context, "rollapp_1-1")
				require.False(t, found)
				require.Empty(t, alias)

				rollAppId, found := dk.GetRollAppIdByAlias(context, "dym")
				require.False(t, found)
				require.Empty(t, rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is a RollApp-ID",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2", "rollapp"},
			preRunSetup: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)

				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "dymension_1100-1",
						Aliases: []string{"dym"},
					},
				}

				require.NoError(t, dk.SetParams(ctx, moduleParams))

				err := dk.SetAliasForRollAppId(ctx, "rollapp_2-2", "ra2")
				require.NoError(t, err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "rollapp",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				alias, found := dk.GetAliasByRollAppId(context, "rollapp_1-1")
				require.False(t, found)
				require.Empty(t, alias)

				rollAppId, found := dk.GetRollAppIdByAlias(context, "rollapp")
				require.False(t, found)
				require.Empty(t, rollAppId)
			},
		},
		{
			name:        "fail - reject if alias used by another RollApp",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2", "rollapp"},
			preRunSetup: func(t *testing.T, ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)

				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "dymension_1100-1",
						Aliases: []string{"dym"},
					},
				}

				require.NoError(t, dk.SetParams(ctx, moduleParams))

				err := dk.SetAliasForRollAppId(ctx, "rollapp_2-2", "alias")
				require.NoError(t, err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				alias, found := dk.GetAliasByRollAppId(context, "rollapp_1-1")
				require.False(t, found)
				require.Empty(t, alias)

				rollAppId, found := dk.GetRollAppIdByAlias(context, "alias")
				require.True(t, found)
				require.Equal(t, "rollapp_2-2", rollAppId)
			},
		},
		{
			name:                    "fail - reject if creator does not have enough balance to pay the fee",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  1,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "insufficient funds",
			wantSuccess:             false,
			wantLaterCreatorBalance: 1,
		},
		{
			name:                   "pass - can resolve address using alias",
			addRollApps:            []string{"rollapp_1-1"},
			preRunSetup:            nil,
			originalCreatorBalance: price5L,
			rollAppId:              "rollapp_1-1",
			alias:                  "alias",
			wantErr:                false,
			wantSuccess:            true,
			postTest: func(t *testing.T, context sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				dymName := dymnstypes.DymName{
					Name:       "my-name",
					Owner:      dymNameOwnerAcc.bech32(),
					Controller: dymNameOwnerAcc.bech32(),
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "rollapp_1-1",
							Value:   dymNameOwnerAcc.bech32(),
						},
						{
							Type:    dymnstypes.DymNameConfigType_DCT_NAME,
							ChainId: "rollapp_1-1",
							Path:    "sub",
							Value:   anotherAcc.bech32(),
						},
					},
				}
				setDymNameWithFunctionsAfter(context, dymName, t, dk)

				outputAddr, err := dk.ResolveByDymNameAddress(context, "my-name@rollapp_1-1")
				require.NoError(t, err)
				require.Equal(t, dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(context, "my-name@alias")
				require.NoError(t, err)
				require.Equal(t, dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(context, "sub.my-name@alias")
				require.NoError(t, err)
				require.Equal(t, anotherAcc.bech32(), outputAddr)

				outputs, err := dk.ReverseResolveDymNameAddress(context, anotherAcc.bech32(), "rollapp_1-1")
				require.NoError(t, err)
				require.NotEmpty(t, outputs)
				require.Equal(t, "sub.my-name@alias", outputs[0].String())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEqual(t, tt.wantSuccess, tt.wantErr, "mis-configured test case")

			dk, bk, rk, ctx := setupTest()

			if tt.originalCreatorBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName, dymnsutils.TestCoins(tt.originalCreatorBalance),
				)
				require.NoError(t, err)

				err = bk.SendCoinsFromModuleToAccount(
					ctx,
					dymnstypes.ModuleName, creatorAccAddr,
					dymnsutils.TestCoins(tt.originalCreatorBalance),
				)
				require.NoError(t, err)
			}

			if tt.originalModuleBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName, dymnsutils.TestCoins(tt.originalModuleBalance),
				)
				require.NoError(t, err)
			}

			for _, rollAppId := range tt.addRollApps {
				rk.SetRollapp(ctx, rollapptypes.Rollapp{
					RollappId: rollAppId,
					Creator:   creatorAccAddr.String(),
				})
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(t, ctx, dk, rk)
			}

			err := dk.GetRollAppHooks().RollappCreated(ctx, tt.rollAppId, tt.alias, creatorAccAddr)

			defer func() {
				if t.Failed() {
					return
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, dymnsutils.TestCoin(0).Denom)
				require.NotNil(t, laterModuleBalance)
				require.Equal(
					t,
					tt.originalModuleBalance, laterModuleBalance.Amount.Int64(),
					"module balance should not be changed regardless of success because of burn",
				)

				if tt.postTest != nil {
					tt.postTest(t, ctx, dk, rk)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)

			laterCreatorBalance := bk.GetBalance(ctx, creatorAccAddr, dymnsutils.TestCoin(0).Denom)
			require.NotNil(t, laterCreatorBalance)
			require.Equal(t, tt.wantLaterCreatorBalance, laterCreatorBalance.Amount.Int64(), "creator balance mismatch")
		})
	}
}
