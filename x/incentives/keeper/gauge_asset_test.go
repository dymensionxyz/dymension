package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// TestGetDistributeToBaseLocks tests Keeper.GetDistributeToBaseLocks for gauges with only duration, only lock age, and both.
func (suite *KeeperTestSuite) TestGetDistributeToBaseLocks() {
	suite.SetupTest()

	baseDenom := defaultLPDenom
	coins := sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 10))
	owner := suite.setupAddr(0, "", coins.MulInt(math.NewInt(3)))

	// Create three locks:
	// 1. duration = 10s, created 20s ago
	// 2. duration = 20s, created 10s ago
	// 3. duration = 30s, created 40s ago
	blockTime := time.Now()
	suite.Ctx = suite.Ctx.WithBlockTime(blockTime)

	locks := []struct {
		lockDuration     time.Duration
		lockAmount       sdk.Coins
		lockCreationTime time.Time
	}{
		{
			lockDuration:     10 * time.Second,
			lockAmount:       coins,
			lockCreationTime: blockTime.Add(-20 * time.Second),
		},
		{
			lockDuration:     20 * time.Second,
			lockAmount:       coins,
			lockCreationTime: blockTime.Add(-10 * time.Second),
		},
		{
			lockDuration:     30 * time.Second,
			lockAmount:       coins,
			lockCreationTime: blockTime.Add(-40 * time.Second),
		},
	}

	for _, lock := range locks {
		suite.Ctx = suite.Ctx.WithBlockTime(lock.lockCreationTime)
		suite.LockTokens(owner, lock.lockAmount, lock.lockDuration)
	}

	suite.Ctx = suite.Ctx.WithBlockTime(blockTime)

	testCases := []struct {
		name     string
		cond     lockuptypes.QueryCondition
		expected []uint64 // expected lock IDs
	}{
		{
			name: "only duration",
			cond: lockuptypes.QueryCondition{
				Denom:    baseDenom,
				Duration: 20 * time.Second,
			},
			expected: []uint64{2, 3},
		},
		{
			name: "only lock age",
			cond: lockuptypes.QueryCondition{
				Denom:   baseDenom,
				LockAge: 15 * time.Second,
			},
			expected: []uint64{1, 3},
		},
		{
			name: "duration and lock age",
			cond: lockuptypes.QueryCondition{
				Denom:    baseDenom,
				Duration: 20 * time.Second,
				LockAge:  15 * time.Second,
			},
			expected: []uint64{3},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			gauge := types.NewAssetGauge(1, true, tc.cond, coins, blockTime, 1)
			cache := make(types.DenomLocksCache)
			locks := suite.App.IncentivesKeeper.GetDistributeToBaseLocks(suite.Ctx, gauge, cache)
			var gotIDs []uint64
			for _, l := range locks {
				gotIDs = append(gotIDs, l.ID)
			}
			suite.ElementsMatch(tc.expected, gotIDs)
		})
	}
}
