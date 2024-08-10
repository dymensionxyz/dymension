package keeper_test

import (
	"sort"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

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
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  getEpochWithOffset(offsetExpiry),
			MinPrice:  dymnsutils.TestCoin(100),
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
		so := ts.dk.GetSellOrder(ts.ctx, dymName.Name, dymnstypes.TypeName)
		require.Nil(t, so)
	}

	requireActiveSO := func(dymName dymnstypes.DymName, ts testSuite) {
		so := ts.dk.GetSellOrder(ts.ctx, dymName.Name, dymnstypes.TypeName)
		require.NotNil(t, so)
	}

	requireHistoricalSOs := func(dymName dymnstypes.DymName, wantCount int, ts testSuite) {
		historicalSOs := ts.dk.GetHistoricalSellOrders(ts.ctx, dymName.Name, dymnstypes.TypeName)
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
				err = dk.MoveSellOrderToHistorical(ctx, so.AssetId, so.AssetType)
				require.NoError(t, err)
			}

			for _, so := range tt.activeSOs {
				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)
			}

			meh := dk.GetMinExpiryOfAllHistoricalDymNameSellOrders(ctx)
			if len(meh) > 0 {
				// clear existing records to simulate cases of malformed state
				for _, record := range meh {
					dk.SetMinExpiryHistoricalSellOrder(ctx, record.DymName, dymnstypes.TypeName, 0)
				}
			}
			if len(tt.minExpiryByDymName) > 0 {
				for dymName, minExpiry := range tt.minExpiryByDymName {
					dk.SetMinExpiryHistoricalSellOrder(ctx, dymName, dymnstypes.TypeName, minExpiry)
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

				meh := dk.GetMinExpiryOfAllHistoricalDymNameSellOrders(ctx)
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

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_epochHooks_AfterEpochEnd() {
	s.Run("should do something even nothing to do", func() {
		s.SetupTest()

		moduleParams := s.moduleParams()

		originalGas := s.ctx.GasMeter().GasConsumed()

		err := s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(
			s.ctx,
			moduleParams.Misc.EndEpochHookIdentifier, 1,
		)
		s.Require().NoError(err)

		// gas should be changed because it should at least reading the params to check epoch identifier
		s.Require().Less(originalGas, s.ctx.GasMeter().GasConsumed(), "should do something")
	})

	s.Run("process active mixed Dym-Name and alias Sell-Orders", func() {
		s.SetupTest()

		dymNameOwner := testAddr(1).bech32()
		dymNameBuyer := testAddr(2).bech32()

		creator1_asOwner := testAddr(3).bech32()
		creator2_asBuyer := testAddr(4).bech32()

		dymName1 := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      dymNameOwner,
			Controller: dymNameOwner,
			ExpireAt:   s.now.Add(2 * 365 * 24 * time.Hour).Unix(),
		}
		err := s.dymNsKeeper.SetDymName(s.ctx, dymName1)
		s.Require().NoError(err)

		rollApp1_asSrc := *newRollApp("rollapp_1-1").WithOwner(creator1_asOwner).WithAlias("one")
		s.persistRollApp(rollApp1_asSrc)
		s.requireRollApp(rollApp1_asSrc.rollAppId).HasAlias("one")
		rollApp2_asDst := *newRollApp("rollapp_2-2").WithOwner(creator2_asBuyer)
		s.persistRollApp(rollApp2_asDst)
		s.requireRollApp(rollApp2_asDst.rollAppId).HasNoAlias()

		const dymNameOrderPrice = 100
		const aliasOrderPrice = 200

		s.mintToModuleAccount(dymNameOrderPrice + aliasOrderPrice + 1)

		dymNameSO := s.newDymNameSellOrder(dymName1.Name).
			WithMinPrice(dymNameOrderPrice).
			WithDymNameBid(dymNameBuyer, dymNameOrderPrice).
			Expired().Build()
		err = s.dymNsKeeper.SetSellOrder(s.ctx, dymNameSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameSO.AssetId,
					ExpireAt: dymNameSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeName)
		s.Require().NoError(err)

		aliasSO := s.newAliasSellOrder(rollApp1_asSrc.alias).
			WithMinPrice(aliasOrderPrice).
			WithAliasBid(rollApp2_asDst.owner, aliasOrderPrice, rollApp2_asDst.rollAppId).
			Expired().Build()
		err = s.dymNsKeeper.SetSellOrder(s.ctx, aliasSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasSO.AssetId,
					ExpireAt: aliasSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		moduleParams := s.moduleParams()

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))
		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp1_asSrc.alias, dymnstypes.TypeAlias))

		s.Equal(int64(1), s.moduleBalance())
		s.Equal(int64(dymNameOrderPrice), s.balance(dymNameOwner))
		s.Equal(int64(aliasOrderPrice), s.balance(rollApp1_asSrc.owner))

		laterDymName := s.dymNsKeeper.GetDymName(s.ctx, dymName1.Name)
		if s.NotNil(laterDymName) {
			s.Equal(dymNameBuyer, laterDymName.Owner)
			s.Equal(dymNameBuyer, laterDymName.Controller)
		}

		s.requireRollApp(rollApp1_asSrc.rollAppId).HasNoAlias()
		s.requireRollApp(rollApp2_asDst.rollAppId).HasAlias("one")
	})

	s.Run("should not process Dym-Name SO if trading is disabled", func() {
		s.SetupTest()

		dymNameOwner := testAddr(1).bech32()
		dymNameBuyer := testAddr(2).bech32()

		dymName1 := dymnstypes.DymName{
			Name:       "my-name",
			Owner:      dymNameOwner,
			Controller: dymNameOwner,
			ExpireAt:   s.now.Add(2 * 365 * 24 * time.Hour).Unix(),
		}
		err := s.dymNsKeeper.SetDymName(s.ctx, dymName1)
		s.Require().NoError(err)

		const dymNameOrderPrice = 100

		s.mintToModuleAccount(dymNameOrderPrice + 1)

		dymNameSO := s.newDymNameSellOrder(dymName1.Name).
			WithMinPrice(dymNameOrderPrice).
			WithDymNameBid(dymNameBuyer, dymNameOrderPrice).
			Expired().Build()
		err = s.dymNsKeeper.SetSellOrder(s.ctx, dymNameSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameSO.AssetId,
					ExpireAt: dymNameSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeName)
		s.Require().NoError(err)

		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingName = false
			return p
		})

		moduleParams := s.moduleParams()

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		// the SellOrder should still be there
		s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))

		// re-enable and test again to make sure it not processes just because trading was disabled
		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingName = true
			return p
		})

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))
	})

	s.Run("should not process Alias SO if trading is disabled", func() {
		s.SetupTest()

		creator1_asOwner := testAddr(3).bech32()
		creator2_asBuyer := testAddr(4).bech32()

		rollApp1_asSrc := *newRollApp("rollapp_1-1").WithOwner(creator1_asOwner).WithAlias("one")
		s.persistRollApp(rollApp1_asSrc)
		s.requireRollApp(rollApp1_asSrc.rollAppId).HasAlias("one")
		rollApp2_asDst := *newRollApp("rollapp_2-2").WithOwner(creator2_asBuyer)
		s.persistRollApp(rollApp2_asDst)
		s.requireRollApp(rollApp2_asDst.rollAppId).HasNoAlias()

		const aliasOrderPrice = 200

		s.mintToModuleAccount(aliasOrderPrice + 1)

		aliasSO := s.newAliasSellOrder(rollApp1_asSrc.alias).
			WithMinPrice(aliasOrderPrice).
			WithAliasBid(rollApp2_asDst.owner, aliasOrderPrice, rollApp2_asDst.rollAppId).
			Expired().Build()
		err := s.dymNsKeeper.SetSellOrder(s.ctx, aliasSO)
		s.Require().NoError(err)
		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasSO.AssetId,
					ExpireAt: aliasSO.ExpireAt,
				},
			},
		}, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingAlias = false
			return p
		})

		moduleParams := s.moduleParams()

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		// the SellOrder should still be there
		s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp1_asSrc.alias, dymnstypes.TypeAlias))

		// re-enable and test again to make sure it not processes just because trading was disabled
		s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
			p.Misc.EnableTradingAlias = true
			return p
		})

		err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, moduleParams.Misc.EndEpochHookIdentifier, 1)
		s.Require().NoError(err)

		s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp1_asSrc.alias, dymnstypes.TypeAlias))
	})
}

func Test_epochHooks_AfterEpochEnd_processActiveDymNameSellOrders(t *testing.T) {
	now := time.Now().UTC()

	setupTest := func() (dymnskeeper.Keeper, dymnstypes.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		return dk, bk, ctx
	}

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
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
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
		bk  dymnstypes.BankKeeper
		ctx sdk.Context
	}

	nts := func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) testSuite {
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
		so := ts.dk.GetSellOrder(ts.ctx, dymName.Name, dymnstypes.TypeName)
		require.Nil(t, so)
	}

	requireHistoricalSOs := func(dymName dymnstypes.DymName, wantCount int, ts testSuite) {
		historicalSOs := ts.dk.GetHistoricalSellOrders(ts.ctx, dymName.Name, dymnstypes.TypeName)
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

	tests := []struct {
		name                  string
		dymNames              []dymnstypes.DymName
		sellOrders            []dymnstypes.SellOrder
		expiryByDymName       []dymnstypes.ActiveSellOrdersExpirationRecord
		preMintModuleBalance  int64
		customEpochIdentifier string
		beforeHookTestFunc    func(*testing.T, dymnskeeper.Keeper, dymnstypes.BankKeeper, sdk.Context)
		wantErr               bool
		wantErrContains       string
		wantExpiryByDymName   []dymnstypes.ActiveSellOrdersExpirationRecord
		afterHookTestFunc     func(*testing.T, dymnskeeper.Keeper, dymnstypes.BankKeeper, sdk.Context)
	}{
		{
			name:       "pass - simple process expired SO",
			dymNames:   []dymnstypes.DymName{dymNameA, dymNameB, dymNameC, dymNameD},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, nil)},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
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
			name:     "pass - simple process expired & completed SO",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  coin200,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
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
			name:     "pass - simple process expired & completed SO, match by min price",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, &coin200, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  coin100,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 250,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:             false,
			wantExpiryByDymName: nil,
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
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
			name:     "pass - process multiple - mixed SOs",
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
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 450,
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)

				// SO for Dym-Name A is expired without any bid/winner
				requireDymNameNotChanged(dymNameA, ts)
				requireNoActiveSO(dymNameA, ts)
				requireHistoricalSOs(dymNameA, 1, ts)

				// SO for Dym-Name B not yet finished
				requireDymNameNotChanged(dymNameB, ts)
				soB := ts.dk.GetSellOrder(ts.ctx, dymNameB.Name, dymnstypes.TypeName)
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
			name:     "pass - should do nothing if invalid epoch identifier",
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
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance:  450,
			customEpochIdentifier: "another",
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name, dymNameC.Name, dymNameD.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  dymNameC.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  dymNameD.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
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
			name:     "pass - should remove expiry reference to non-exists SO",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				// no SO for Dym-Name B
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					// no SO for Dym-Name B but still have reference
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
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
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
		},
		{
			name:     "pass - update expiry if in-correct",
			dymNames: []dymnstypes.DymName{dymNameA, dymNameB},
			sellOrders: []dymnstypes.SellOrder{
				genSo(dymNameA, soExpired, nil, nil),
				genSo(dymNameB, soNotExpired, nil, nil), // SO not expired
			},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
				{
					// incorrect, SO not expired
					AssetId:  dymNameB.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr: false,
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameB.Name,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name, dymNameB.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name, dymNameB.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
		},
		{
			name:     "fail - returns error when can not process complete order",
			dymNames: []dymnstypes.DymName{dymNameA},
			sellOrders: []dymnstypes.SellOrder{genSo(dymNameA, soExpired, nil, &dymnstypes.SellOrderBid{
				Bidder: bidderA,
				Price:  coin100,
			})},
			expiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 1, // not enough balance
			beforeHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name)
				requireFallbackAddrMappedNoDymName(ts, bidderAcc.fallback())
			},
			wantErr:         true,
			wantErrContains: "insufficient funds",
			wantExpiryByDymName: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  dymNameA.Name,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(t *testing.T, dk dymnskeeper.Keeper, bk dymnstypes.BankKeeper, ctx sdk.Context) {
				// unchanged
				ts := nts(t, dk, bk, ctx)
				requireConfiguredAddressMappedDymNames(ts, ownerA, dymNameA.Name)
				requireConfiguredAddressMappedNoDymName(ts, bidderA)
				requireFallbackAddrMappedDymNames(ts, ownerAcc.fallback(), dymNameA.Name)
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
			}, dymnstypes.TypeName)
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

				aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeName)
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

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_epochHooks_AfterEpochEnd_processActiveAliasSellOrders() {
	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBidder := testAddr(2).bech32()

	rollApp_1_byOwner_asSrc := *newRollApp("rollapp_1-1").WithAlias("one").WithOwner(creator_1_asOwner)
	rollApp_2_byBuyer_asDst := *newRollApp("rollapp_2-2").WithOwner(creator_2_asBidder)
	rollApp_3_byOwner_asSrc := *newRollApp("rollapp_3-1").WithAlias("three").WithOwner(creator_1_asOwner)
	rollApp_4_byOwner_asSrc := *newRollApp("rollapp_4-1").WithAlias("four").WithOwner(creator_1_asOwner)
	rollApp_5_byOwner_asSrc := *newRollApp("rollapp_5-1").WithAlias("five").WithOwner(creator_1_asOwner)

	const aliasProhibitedTrading = "prohibited"

	const minPrice = 100
	const soExpiredEpoch = 1
	soNotExpiredEpoch := s.now.Add(time.Hour).Unix()

	requireNoActiveSO := func(alias string) {
		so := s.dymNsKeeper.GetSellOrder(s.ctx, alias, dymnstypes.TypeAlias)
		s.Nil(so)
	}

	requireNoHistoricalSO := func(alias string) {
		historicalSOs := s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, alias, dymnstypes.TypeAlias)
		s.Empty(historicalSOs, "should have no historical SOs since Alias SOs are not supported")
	}

	tests := []struct {
		name                  string
		rollApps              []rollapp
		sellOrders            []dymnstypes.SellOrder
		expiryByAlias         []dymnstypes.ActiveSellOrdersExpirationRecord
		preMintModuleBalance  int64
		customEpochIdentifier string
		beforeHookTestFunc    func(s *KeeperTestSuite)
		wantErr               bool
		wantErrContains       string
		wantExpiryByAlias     []dymnstypes.ActiveSellOrdersExpirationRecord
		afterHookTestFunc     func(s *KeeperTestSuite)
	}{
		{
			name:     "pass - simple process expired SO without bid",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
			},
			wantErr:           false,
			wantExpiryByAlias: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				requireNoActiveSO(rollApp_1_byOwner_asSrc.alias)
				requireNoHistoricalSO(rollApp_1_byOwner_asSrc.alias)

				// unchanged

				s.Equal(int64(200), s.moduleBalance())
				s.Zero(s.balance(rollApp_1_byOwner_asSrc.owner))

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
		},
		{
			name:     "pass - simple process expired & completed SO",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, 200, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 200,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
			wantErr:           false,
			wantExpiryByAlias: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)

				requireNoActiveSO(rollApp_1_byOwner_asSrc.alias)
				requireNoHistoricalSO(rollApp_1_byOwner_asSrc.alias)

				s.Zero(s.moduleBalance())                                     // should be transferred to previous owner
				s.Equal(int64(200), s.balance(rollApp_1_byOwner_asSrc.owner)) // previous owner should earn from bid
			},
		},
		{
			name:     "pass - simple process expired & completed SO, match by min price",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 250,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
			wantErr:           false,
			wantExpiryByAlias: nil,
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)

				requireNoActiveSO(rollApp_1_byOwner_asSrc.alias)
				requireNoHistoricalSO(rollApp_1_byOwner_asSrc.alias)

				s.Equal(int64(250-minPrice), s.moduleBalance())                    // should be transferred to previous owner
				s.Equal(int64(minPrice), s.balance(rollApp_1_byOwner_asSrc.owner)) // previous owner should earn from bid
			},
		},
		{
			name: "pass - refunds records that alias presents in params",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst,
			},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(aliasProhibitedTrading).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  aliasProhibitedTrading,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 500,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp_1_byOwner_asSrc.rollAppId, aliasProhibitedTrading)
				s.NoError(err)

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(
					rollApp_1_byOwner_asSrc.alias, aliasProhibitedTrading,
				)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()

				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = append(p.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
						ChainId: "some-chain",
						Aliases: []string{aliasProhibitedTrading},
					})
					return p
				})
			},
			wantErr:           false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, aliasProhibitedTrading, dymnstypes.TypeAlias))

				// refunded
				s.Equal(int64(500-minPrice), s.moduleBalance())
				s.Equal(int64(minPrice), s.balance(rollApp_2_byBuyer_asDst.owner))
			},
		},
		{
			name: "pass - process multiple - mixed SOs",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc, rollApp_4_byOwner_asSrc, rollApp_5_byOwner_asSrc,
			},
			sellOrders: []dymnstypes.SellOrder{
				// expired SO without bid
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					Build(),
				// not yet finished
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soNotExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by matching sell-price
				s.newAliasSellOrder(rollApp_4_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, 200, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by min price
				s.newAliasSellOrder(rollApp_5_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by min price, but prohibited trading because presents in module params
				s.newAliasSellOrder(aliasProhibitedTrading).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  rollApp_4_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_5_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  aliasProhibitedTrading,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 450,
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp_1_byOwner_asSrc.rollAppId, aliasProhibitedTrading)
				s.NoError(err)

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(
					rollApp_1_byOwner_asSrc.alias, aliasProhibitedTrading,
				)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasAlias(rollApp_4_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasAlias(rollApp_5_byOwner_asSrc.alias)

				s.updateModuleParams(func(p dymnstypes.Params) dymnstypes.Params {
					p.Chains.AliasesOfChainIds = append(p.Chains.AliasesOfChainIds, dymnstypes.AliasesOfChainId{
						ChainId: "some-chain",
						Aliases: []string{aliasProhibitedTrading},
					})
					return p
				})
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// SO for alias 1 is expired without any bid/winner
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				requireNoActiveSO(rollApp_1_byOwner_asSrc.alias)
				requireNoHistoricalSO(rollApp_1_byOwner_asSrc.alias)

				// SO of the prohibited alias should be removed
				s.Nil(s.dymNsKeeper.GetSellOrder(s.ctx, aliasProhibitedTrading, dymnstypes.TypeAlias))

				// SO for alias 3 not yet finished
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, rollApp_3_byOwner_asSrc.alias, dymnstypes.TypeAlias))
				requireNoHistoricalSO(rollApp_3_byOwner_asSrc.alias)

				// SO for alias 4 is completed with winner
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasNoAlias()
				requireNoActiveSO(rollApp_4_byOwner_asSrc.alias)
				requireNoHistoricalSO(rollApp_4_byOwner_asSrc.alias)

				// SO for alias 5 is completed with winner
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasNoAlias()
				requireNoActiveSO(rollApp_5_byOwner_asSrc.alias)
				requireNoHistoricalSO(rollApp_5_byOwner_asSrc.alias)

				// aliases moved to RollApps of the winner
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).
					HasAlias(rollApp_4_byOwner_asSrc.alias, rollApp_5_byOwner_asSrc.alias)

				s.Equal(int64(50), s.moduleBalance())
				s.Equal(int64(300), s.balance(creator_1_asOwner))  // price from 2 completed SO
				s.Equal(int64(100), s.balance(creator_2_asBidder)) // refunded from prohibited trading SO
			},
		},
		{
			name: "pass - should do nothing if invalid epoch identifier",
			rollApps: []rollapp{
				rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc, rollApp_4_byOwner_asSrc, rollApp_5_byOwner_asSrc,
			},
			sellOrders: []dymnstypes.SellOrder{
				// expired SO without bid
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					Build(),
				// not yet finished
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soNotExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by matching sell-price
				s.newAliasSellOrder(rollApp_4_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, 200, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
				// completed by min price
				s.newAliasSellOrder(rollApp_5_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(200).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(rollApp_2_byBuyer_asDst.owner, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
				{
					AssetId:  rollApp_4_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_5_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			customEpochIdentifier: "another",
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasAlias(rollApp_4_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasAlias(rollApp_5_byOwner_asSrc.alias)
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// deep unchanged but order changed due to sorting
				{
					AssetId:  rollApp_5_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_4_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged

				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_4_byOwner_asSrc.rollAppId).HasAlias(rollApp_4_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_5_byOwner_asSrc.rollAppId).HasAlias(rollApp_5_byOwner_asSrc.alias)
			},
		},
		{
			name:     "pass - should remove expiry reference to non-exists SO",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_3_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithSellPrice(300).
					WithExpiry(soExpiredEpoch).
					Build(),
				// no SO for alias of rollapp 3
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					// no SO for alias of RollApp 3 but still have reference
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
			},
			wantErr:           false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to alias of RollApp 1 because of processed
				// removed reference to alias of RollApp 2 because SO not exists
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_3_byOwner_asSrc.rollAppId).HasAlias(rollApp_3_byOwner_asSrc.alias)
			},
		},
		{
			name:     "pass - update expiry if in-correct",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst, rollApp_3_byOwner_asSrc},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					Build(),
				s.newAliasSellOrder(rollApp_3_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soNotExpiredEpoch). // SO not expired
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
				{
					// incorrect, SO not expired
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			beforeHookTestFunc: func(s *KeeperTestSuite) {
			},
			wantErr: false,
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				// removed reference to alias of RollApp 1 because of processed
				// reference to alias of RollApp 3 was kept because not expired
				{
					AssetId:  rollApp_3_byOwner_asSrc.alias,
					ExpireAt: soNotExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
			},
		},
		{
			name:     "fail - returns error when can not process complete order",
			rollApps: []rollapp{rollApp_1_byOwner_asSrc, rollApp_2_byBuyer_asDst},
			sellOrders: []dymnstypes.SellOrder{
				s.newAliasSellOrder(rollApp_1_byOwner_asSrc.alias).
					WithMinPrice(minPrice).
					WithExpiry(soExpiredEpoch).
					WithAliasBid(creator_2_asBidder, minPrice, rollApp_2_byBuyer_asDst.rollAppId).
					Build(),
			},
			expiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			preMintModuleBalance: 1, // not enough balance
			beforeHookTestFunc: func(s *KeeperTestSuite) {
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
			wantErr:         true,
			wantErrContains: "insufficient funds",
			wantExpiryByAlias: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  rollApp_1_byOwner_asSrc.alias,
					ExpireAt: soExpiredEpoch,
				},
			},
			afterHookTestFunc: func(s *KeeperTestSuite) {
				// unchanged
				s.requireRollApp(rollApp_1_byOwner_asSrc.rollAppId).HasAlias(rollApp_1_byOwner_asSrc.alias)
				s.requireRollApp(rollApp_2_byBuyer_asDst.rollAppId).HasNoAlias()
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			s.Require().NotNil(tt.beforeHookTestFunc, "mis-configured test case")
			s.Require().NotNil(tt.afterHookTestFunc, "mis-configured test case")

			if tt.preMintModuleBalance > 0 {
				s.mintToModuleAccount(tt.preMintModuleBalance)
			}

			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: tt.expiryByAlias,
			}, dymnstypes.TypeAlias)
			s.Require().NoError(err)

			for _, rollApp := range tt.rollApps {
				s.persistRollApp(rollApp)
			}

			for _, so := range tt.sellOrders {
				err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			}

			useEpochIdentifier := s.moduleParams().Misc.EndEpochHookIdentifier
			if tt.customEpochIdentifier != "" {
				useEpochIdentifier = tt.customEpochIdentifier
			}

			tt.beforeHookTestFunc(s)

			err = s.dymNsKeeper.GetEpochHooks().AfterEpochEnd(s.ctx, useEpochIdentifier, 1)

			defer func() {
				tt.afterHookTestFunc(s)

				aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)
				if len(tt.wantExpiryByAlias) == 0 {
					s.Empty(aSoe.Records)
				} else {
					s.Equal(tt.wantExpiryByAlias, aSoe.Records)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)

				return
			}

			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) Test_rollappHooks_RollappCreated() {
	const price1L = 9
	const price2L = 8
	const price3L = 7
	const price4L = 6
	const price5L = 5
	const price6L = 4
	const price7PL = 3

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	setupParams := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Price.AliasPriceSteps = []sdkmath.Int{
				sdk.NewInt(price1L).Mul(priceMultiplier),
				sdk.NewInt(price2L).Mul(priceMultiplier),
				sdk.NewInt(price3L).Mul(priceMultiplier),
				sdk.NewInt(price4L).Mul(priceMultiplier),
				sdk.NewInt(price5L).Mul(priceMultiplier),
				sdk.NewInt(price6L).Mul(priceMultiplier),
				sdk.NewInt(price7PL).Mul(priceMultiplier),
			}
			return moduleParams
		})
	}

	creatorAccAddr := sdk.AccAddress(testAddr(1).bytes())
	dymNameOwnerAcc := testAddr(2)
	anotherAcc := testAddr(3)

	tests := []struct {
		name                    string
		addRollApps             []string
		preRunSetup             func(s *KeeperTestSuite)
		originalCreatorBalance  int64
		originalModuleBalance   int64
		rollAppId               string
		alias                   string
		wantErr                 bool
		wantErrContains         string
		wantSuccess             bool
		wantLaterCreatorBalance int64
		postTest                func(s *KeeperTestSuite)
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
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.True(found)
				s.Equal("alias", alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_1-1", rollAppId)
			},
		},
		{
			name:                    "pass - if input alias is empty, do nothing",
			addRollApps:             []string{"rollapp_1-1"},
			originalCreatorBalance:  0,
			rollAppId:               "rollapp_1-1",
			alias:                   "",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 0,
			postTest: func(s *KeeperTestSuite) {
				// mapping should not be created

				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.False(found)
				s.Empty(rollAppId)
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
			postTest: func(s *KeeperTestSuite) {
				_, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "nad_0-0")
				s.False(found)

				_, found = s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.False(found)
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
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price5L + 2,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 false,
			wantSuccess:             true,
			wantLaterCreatorBalance: 2,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.True(found)
				s.Equal("alias", alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_1-1", rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "bridge",
							Aliases: []string{"b"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "bridge",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "bridge")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is presents as alias of a chain-id in params",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "dym",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "dym")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias is a RollApp-ID",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.pureSetRollApp(rollapptypes.Rollapp{
					RollappId: "rollapp",
					Owner:     creatorAccAddr.String(),
				})

				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "ra2")
				s.Require().NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "rollapp",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "rollapp")
				s.False(found)
				s.Empty(rollAppId)
			},
		},
		{
			name:        "fail - reject if alias used by another RollApp",
			addRollApps: []string{"rollapp_1-1", "rollapp_2-2"},
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "dymension_1100-1",
							Aliases: []string{"dym"},
						},
					}
					return moduleParams
				})

				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_2-2", "alias")
				s.Require().NoError(err)
			},
			originalCreatorBalance:  price1L,
			originalModuleBalance:   1,
			rollAppId:               "rollapp_1-1",
			alias:                   "alias",
			wantErr:                 true,
			wantErrContains:         "alias already in use or preserved",
			wantSuccess:             false,
			wantLaterCreatorBalance: price1L,
			postTest: func(s *KeeperTestSuite) {
				alias, found := s.dymNsKeeper.GetAliasByRollAppId(s.ctx, "rollapp_1-1")
				s.False(found)
				s.Empty(alias)

				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, "alias")
				s.True(found)
				s.Equal("rollapp_2-2", rollAppId)
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
			postTest: func(s *KeeperTestSuite) {
				dymName := dymnstypes.DymName{
					Name:       "my-name",
					Owner:      dymNameOwnerAcc.bech32(),
					Controller: dymNameOwnerAcc.bech32(),
					ExpireAt:   s.now.Unix() + 1,
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
				s.setDymNameWithFunctionsAfter(dymName)

				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "my-name@rollapp_1-1")
				s.Require().NoError(err)
				s.Equal(dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "my-name@alias")
				s.Require().NoError(err)
				s.Equal(dymNameOwnerAcc.bech32(), outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub.my-name@alias")
				s.Require().NoError(err)
				s.Equal(anotherAcc.bech32(), outputAddr)

				outputs, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, anotherAcc.bech32(), "rollapp_1-1")
				s.Require().NoError(err)
				s.Require().NotEmpty(outputs)
				s.Equal("sub.my-name@alias", outputs[0].String())
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Require().NotEqual(tt.wantSuccess, tt.wantErr, "mis-configured test case")

			s.SetupTest()

			setupParams(s)

			if tt.originalCreatorBalance > 0 {
				s.mintToAccount2(creatorAccAddr.String(), sdk.NewInt(tt.originalCreatorBalance).Mul(priceMultiplier))
			}

			if tt.originalModuleBalance > 0 {
				s.mintToModuleAccount2(sdk.NewInt(tt.originalModuleBalance).Mul(priceMultiplier))
			}

			for _, rollAppId := range tt.addRollApps {
				s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
					RollappId: rollAppId,
					Owner:     creatorAccAddr.String(),
				})
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			err := s.dymNsKeeper.GetRollAppHooks().RollappCreated(s.ctx, tt.rollAppId, tt.alias, creatorAccAddr)

			defer func() {
				if s.T().Failed() {
					return
				}

				laterModuleBalance := s.moduleBalance2()
				s.Equal(
					sdk.NewInt(tt.originalModuleBalance).Mul(priceMultiplier),
					laterModuleBalance,
					"module balance should not be changed regardless of success because of burn",
				)

				if tt.postTest != nil {
					tt.postTest(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)

			laterCreatorBalance := s.balance2(creatorAccAddr.String())
			s.Equal(
				sdk.NewInt(tt.wantLaterCreatorBalance).Mul(priceMultiplier),
				laterCreatorBalance,
				"creator balance mismatch",
			)

			// event should be fired
			func() {
				if tt.alias == "" {
					return
				}

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

				for _, event := range events {
					if event.Type == dymnstypes.EventTypeSell {
						return
					}
				}

				s.T().Errorf("event %s not found", dymnstypes.EventTypeSell)
			}()
		})
	}

	s.Run("if alias is empty, do nothing", func() {
		originalTxGas := s.ctx.GasMeter().GasConsumed()

		err := s.dymNsKeeper.GetRollAppHooks().RollappCreated(s.ctx, "rollapp_1-1", "", creatorAccAddr)
		s.Require().NoError(err)

		s.Equal(originalTxGas, s.ctx.GasMeter().GasConsumed(), "should not consume gas")
	})
}
