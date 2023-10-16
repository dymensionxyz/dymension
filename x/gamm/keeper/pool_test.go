package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/dymensionxyz/dymension/testutil/tests/mocks"
	"github.com/dymensionxyz/dymension/x/gamm/keeper"
	"github.com/dymensionxyz/dymension/x/gamm/pool-models/balancer"
	balancertypes "github.com/dymensionxyz/dymension/x/gamm/pool-models/balancer"
	"github.com/dymensionxyz/dymension/x/gamm/pool-models/stableswap"
	"github.com/dymensionxyz/dymension/x/gamm/types"
	poolmanagertypes "github.com/dymensionxyz/dymension/x/poolmanager/types"
)

var (
	defaultPoolAssetsStableSwap = sdk.Coins{
		sdk.NewCoin("atom", sdk.NewInt(100)),
		sdk.NewCoin("osmo", sdk.NewInt(100)),
	}
	defaultPoolParamsStableSwap = stableswap.PoolParams{
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(1, 2),
	}
	defaultPoolId                        = uint64(1)
	defaultAcctFundsStableSwap sdk.Coins = sdk.NewCoins(
		sdk.NewCoin("udym", sdk.NewInt(10000000000)),
		sdk.NewCoin("atom", sdk.NewInt(100)),
		sdk.NewCoin("osmo", sdk.NewInt(100)),
	)
)

// TestGetPoolAndPoke tests that the right pools is returned from GetPoolAndPoke.
// For the pools implementing the weighted extension, asserts that PokePool is called.
func (suite *KeeperTestSuite) TestGetPoolAndPoke() {
	const (
		startTime = 1000
		blockTime = startTime + 100
	)

	// N.B.: We make a copy because SmoothWeightChangeParams get mutated.
	// We would like to avoid mutating global pool assets that are used in other tests.
	defaultPoolAssetsCopy := make([]balancertypes.PoolAsset, 2)
	copy(defaultPoolAssetsCopy, defaultPoolAssets)

	startPoolWeightAssets := []balancertypes.PoolAsset{
		{
			Weight: defaultPoolAssets[0].Weight.Quo(sdk.NewInt(2)),
			Token:  defaultPoolAssets[0].Token,
		},
		{
			Weight: defaultPoolAssets[1].Weight.Mul(sdk.NewInt(3)),
			Token:  defaultPoolAssets[1].Token,
		},
	}

	tests := map[string]struct {
		isPokePool bool
		poolId     uint64
	}{
		"weighted pool - change weights": {
			isPokePool: true,
			poolId: suite.prepareCustomBalancerPool(defaultAcctFunds, startPoolWeightAssets, balancer.PoolParams{
				SwapFee: defaultSwapFee,
				ExitFee: defaultExitFee,
				SmoothWeightChangeParams: &balancer.SmoothWeightChangeParams{
					StartTime:          time.Unix(startTime, 0), // start time is before block time so the weights should change
					Duration:           time.Hour,
					InitialPoolWeights: startPoolWeightAssets,
					TargetPoolWeights:  defaultPoolAssetsCopy,
				},
			}),
		},
		"non weighted pool": {
			poolId: suite.prepareCustomStableswapPool(
				defaultAcctFunds,
				stableswap.PoolParams{
					SwapFee: defaultSwapFee,
					ExitFee: defaultExitFee,
				},
				sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[0].Denom, defaultAcctFunds[0].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2))),
				[]uint64{1, 1},
			),
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			k := suite.App.GAMMKeeper
			ctx := suite.Ctx.WithBlockTime(time.Unix(blockTime, 0))

			pool, err := k.GetPoolAndPoke(ctx, tc.poolId)

			suite.Require().NoError(err)
			suite.Require().Equal(tc.poolId, pool.GetId())

			if tc.isPokePool {
				pokePool, ok := pool.(types.WeightedPoolExtension)
				suite.Require().True(ok)

				poolAssetWeight0, err := pokePool.GetTokenWeight(startPoolWeightAssets[0].Token.Denom)
				suite.Require().NoError(err)

				poolAssetWeight1, err := pokePool.GetTokenWeight(startPoolWeightAssets[1].Token.Denom)
				suite.Require().NoError(err)

				suite.Require().NotEqual(startPoolWeightAssets[0].Weight, poolAssetWeight0)
				suite.Require().NotEqual(startPoolWeightAssets[1].Weight, poolAssetWeight1)
				return
			}

			_, ok := pool.(types.WeightedPoolExtension)
			suite.Require().False(ok)
		})
	}
}

func (suite *KeeperTestSuite) TestConvertToCFMMPool() {
	ctrl := gomock.NewController(suite.T())

	tests := map[string]struct {
		pool        poolmanagertypes.PoolI
		expectError bool
	}{
		"cfmm pool": {
			pool: mocks.NewMockCFMMPoolI(ctrl),
		},
		"non cfmm pool": {
			pool:        mocks.NewMockConcentratedPoolExtension(ctrl),
			expectError: true,
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {
			suite.SetupTest()

			pool, err := keeper.ConvertToCFMMPool(tc.pool)

			if tc.expectError {
				suite.Require().Error(err)
				suite.Require().Nil(pool)
				return
			}

			suite.Require().NoError(err)
			suite.Require().NotNil(pool)
			suite.Require().Equal(tc.pool, pool)
		})
	}
}

// TestMarshalUnmarshalPool tests that by changing the interfaces
// that we marshal to and unmarshal from store, we do not
// change the underlying bytes. This shows that migrations are
// not necessary.
func (suite *KeeperTestSuite) TestMarshalUnmarshalPool() {

	suite.SetupTest()
	k := suite.App.GAMMKeeper

	balancerPoolId := suite.PrepareBalancerPool()
	balancerPool, err := k.GetPoolAndPoke(suite.Ctx, balancerPoolId)
	suite.Require().NoError(err)

	stableswapPoolId := suite.PrepareBasicStableswapPool()
	stableswapPool, err := k.GetPoolAndPoke(suite.Ctx, stableswapPoolId)
	suite.Require().NoError(err)

	tests := []struct {
		name string
		pool types.CFMMPoolI
	}{
		{
			name: "balancer",
			pool: balancerPool,
		},
		{
			name: "stableswap",
			pool: stableswapPool,
		},
	}

	for _, tc := range tests {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			var poolI poolmanagertypes.PoolI = tc.pool
			var cfmmPoolI types.CFMMPoolI = tc.pool

			// Marshal poolI as PoolI
			bzPoolI, err := k.MarshalPool(poolI)
			suite.Require().NoError(err)

			// Marshal cfmmPoolI as PoolI
			bzCfmmPoolI, err := k.MarshalPool(cfmmPoolI)
			suite.Require().NoError(err)

			suite.Require().Equal(bzPoolI, bzCfmmPoolI)

			// Unmarshal bzPoolI as CFMMPoolI
			unmarshalBzPoolIAsCfmmPoolI, err := k.UnmarshalPool(bzPoolI)
			suite.Require().NoError(err)

			// Unmarshal bzPoolI as PoolI
			unmarshalBzPoolIAsPoolI, err := k.UnmarshalPoolLegacy(bzPoolI)
			suite.Require().NoError(err)

			suite.Require().Equal(unmarshalBzPoolIAsCfmmPoolI, unmarshalBzPoolIAsPoolI)

			// Unmarshal bzCfmmPoolI as CFMMPoolI
			unmarshalBzCfmmPoolIAsCfmmPoolI, err := k.UnmarshalPool(bzCfmmPoolI)
			suite.Require().NoError(err)

			// Unmarshal bzCfmmPoolI as PoolI
			unmarshalBzCfmmPoolIAsPoolI, err := k.UnmarshalPoolLegacy(bzCfmmPoolI)
			suite.Require().NoError(err)

			// bzCfmmPoolI as CFMMPoolI equals bzCfmmPoolI as PoolI
			suite.Require().Equal(unmarshalBzCfmmPoolIAsCfmmPoolI, unmarshalBzCfmmPoolIAsPoolI)

			// All unmarshalled combinations are equal.
			suite.Require().Equal(unmarshalBzPoolIAsCfmmPoolI, unmarshalBzCfmmPoolIAsCfmmPoolI)
		})
	}
}

func (suite *KeeperTestSuite) TestSetStableSwapScalingFactors() {
	controllerAddr := suite.TestAccs[0]
	failAddr := suite.TestAccs[1]

	testcases := []struct {
		name             string
		poolId           uint64
		scalingFactors   []uint64
		sender           sdk.AccAddress
		expError         error
		isStableSwapPool bool
	}{
		{
			name:             "Error: Pool does not exist",
			poolId:           2,
			scalingFactors:   []uint64{1, 1},
			sender:           controllerAddr,
			expError:         types.PoolDoesNotExistError{PoolId: defaultPoolId + 1},
			isStableSwapPool: false,
		},
		{
			name:             "Error: Pool id is not of type stableswap pool",
			poolId:           1,
			scalingFactors:   []uint64{1, 1},
			sender:           controllerAddr,
			expError:         fmt.Errorf("pool id 1 is not of type stableswap pool"),
			isStableSwapPool: false,
		},
		{
			name:             "Error: Can not set scaling factors",
			poolId:           1,
			scalingFactors:   []uint64{1, 1},
			sender:           failAddr,
			expError:         types.ErrNotScalingFactorGovernor,
			isStableSwapPool: true,
		},
		{
			name:             "Valid case",
			poolId:           1,
			scalingFactors:   []uint64{1, 1},
			sender:           controllerAddr,
			isStableSwapPool: true,
		},
	}
	for _, tc := range testcases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.isStableSwapPool == true {
				poolId := suite.prepareCustomStableswapPool(
					defaultAcctFunds,
					stableswap.PoolParams{
						SwapFee: defaultSwapFee,
						ExitFee: defaultExitFee,
					},
					sdk.NewCoins(sdk.NewCoin(defaultAcctFunds[0].Denom, defaultAcctFunds[0].Amount.QuoRaw(2)), sdk.NewCoin(defaultAcctFunds[1].Denom, defaultAcctFunds[1].Amount.QuoRaw(2))),
					tc.scalingFactors,
				)
				pool, _ := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
				stableswapPool, _ := pool.(*stableswap.Pool)
				stableswapPool.ScalingFactorController = controllerAddr.String()
				suite.App.GAMMKeeper.SetPool(suite.Ctx, stableswapPool)
			} else {
				suite.prepareCustomBalancerPool(
					defaultAcctFunds,
					defaultPoolAssets,
					defaultPoolParams)
			}
			err := suite.App.GAMMKeeper.SetStableSwapScalingFactors(suite.Ctx, tc.poolId, tc.scalingFactors, tc.sender.String())
			if tc.expError != nil {
				suite.Require().Error(err)
				suite.Require().EqualError(err, tc.expError.Error())
			} else {
				suite.Require().NoError(err)
			}
		})
	}

}
