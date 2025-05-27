package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

var (
	now         = time.Now().UTC()
	acc1        = sdk.AccAddress([]byte("addr1---------------"))
	acc2        = sdk.AccAddress([]byte("addr2---------------"))
	testGenesis = types.GenesisState{
		LastLockId: 10,
		Locks: []types.PeriodLock{
			{
				ID:        1,
				Owner:     acc1.String(),
				Duration:  time.Second,
				EndTime:   time.Time{},
				Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 10000000)},
				UpdatedAt: now,
			},
			{
				ID:        2,
				Owner:     acc1.String(),
				Duration:  time.Hour,
				EndTime:   time.Time{},
				Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 15000000)},
				UpdatedAt: now,
			},
			{
				ID:        3,
				Owner:     acc2.String(),
				Duration:  time.Minute,
				EndTime:   time.Time{},
				Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
				UpdatedAt: now,
			},
		},
	}
)

func TestInitGenesis(t *testing.T) {
	app := apptesting.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.LockupKeeper.InitGenesis(ctx, genesis)

	coins := app.LockupKeeper.GetAccountLockedCoins(ctx, acc1)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 25000000).String())

	coins = app.LockupKeeper.GetAccountLockedCoins(ctx, acc2)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 5000000).String())

	lastLockId := app.LockupKeeper.GetLastLockID(ctx)
	require.Equal(t, lastLockId, uint64(10))

	acc := app.LockupKeeper.GetPeriodLocksAccumulation(ctx, types.QueryCondition{
		Denom:    "foo",
		Duration: time.Second,
	})
	require.Equal(t, math.NewInt(30000000), acc)
}

func TestExportGenesis(t *testing.T) {
	app := apptesting.Setup(t)
	ctx := app.BaseApp.NewContext(false)
	ctx = ctx.WithBlockTime(now.Add(time.Second))
	genesis := testGenesis
	app.LockupKeeper.InitGenesis(ctx, genesis)

	err := bankutil.FundAccount(ctx, app.BankKeeper, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)})
	require.NoError(t, err)

	_, err = app.LockupKeeper.CreateLock(ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)}, time.Second*5)
	require.NoError(t, err)

	coins := app.LockupKeeper.GetAccountLockedCoins(ctx, acc2)
	require.Equal(t, coins.String(), sdk.NewInt64Coin("foo", 10000000).String())

	genesisExported := app.LockupKeeper.ExportGenesis(ctx)
	require.Equal(t, genesisExported.LastLockId, uint64(11))

	expectedLocks := []types.PeriodLock{
		{
			ID:        1,
			Owner:     acc1.String(),
			Duration:  time.Second,
			EndTime:   time.Time{},
			Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 10000000)},
			UpdatedAt: now,
		},
		{
			ID:        11,
			Owner:     acc2.String(),
			Duration:  time.Second * 5,
			EndTime:   time.Time{},
			Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
			UpdatedAt: ctx.BlockTime(),
		},
		{
			ID:        3,
			Owner:     acc2.String(),
			Duration:  time.Minute,
			EndTime:   time.Time{},
			Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 5000000)},
			UpdatedAt: now,
		},
		{
			ID:        2,
			Owner:     acc1.String(),
			Duration:  time.Hour,
			EndTime:   time.Time{},
			Coins:     sdk.Coins{sdk.NewInt64Coin("foo", 15000000)},
			UpdatedAt: now,
		},
	}

	actualLocks := genesisExported.Locks
	require.Equal(t, len(expectedLocks), len(actualLocks), "number of locks should match")

	timeTolerance := 2 * time.Second // allow 2s tolerance due to block time changes
	for i, exp := range expectedLocks {
		act := actualLocks[i]
		assert.Equal(t, exp.ID, act.ID, "ID mismatch at index %d", i)
		assert.Equal(t, exp.Owner, act.Owner, "Owner mismatch at index %d", i)
		assert.Equal(t, exp.Duration, act.Duration, "Duration mismatch at index %d", i)
		assert.Equal(t, exp.EndTime, act.EndTime, "EndTime mismatch at index %d", i)
		assert.True(t, exp.Coins.Equal(act.Coins), "Coins mismatch at index %d: expected %s, got %s", i, exp.Coins, act.Coins)
		assert.WithinDuration(t, exp.UpdatedAt, act.UpdatedAt, timeTolerance, "UpdatedAt mismatch at index %d: expected %v, got %v", i, exp.UpdatedAt, act.UpdatedAt)
	}
}

func TestMarshalUnmarshalGenesis(t *testing.T) {
	app := apptesting.Setup(t)
	ctx := app.BaseApp.NewContext(false).WithBlockTime(now.Add(time.Second))

	keeper := app.LockupKeeper
	err := bankutil.FundAccount(ctx, app.BankKeeper, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)})
	require.NoError(t, err)
	_, err = app.LockupKeeper.CreateLock(ctx, acc2, sdk.Coins{sdk.NewInt64Coin("foo", 5000000)}, time.Second*5)
	require.NoError(t, err)

	genesisExported := keeper.ExportGenesis(ctx)
	assert.NotPanics(t, func() {
		app = apptesting.Setup(t)
		ctx = app.BaseApp.NewContext(false)
		ctx = ctx.WithBlockTime(now.Add(time.Second))
		keeper := app.LockupKeeper
		keeper.InitGenesis(ctx, *genesisExported)
	})
}
