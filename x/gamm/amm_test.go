package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/stretchr/testify/suite"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(s.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	// set txfees basedenom
	err := app.TxFeesKeeper.SetBaseDenom(ctx, "adym")
	s.Require().NoError(err)

	s.App = app
	s.Ctx = ctx
}

func (s *KeeperTestSuite) TestSwapsRevenue() {
	// Create a pool with 100_000 DYM and 100_000 FOO
	poolCoins := sdk.NewCoins(
		sdk.NewCoin("adym", apptesting.EXP.Mul(sdk.NewInt(100_000))),
		sdk.NewCoin("foo", apptesting.EXP.Mul(sdk.NewInt(100_000))),
	)

	testCases := []struct {
		name       string
		swapFee    sdk.Dec
		takerFee   sdk.Dec
		expRevenue bool
	}{
		{
			name:       "1% swap fee, 1% taker fee",
			swapFee:    sdk.NewDecWithPrec(1, 2), // 1%
			takerFee:   sdk.NewDecWithPrec(1, 2), // 1%
			expRevenue: true,
		},
		{
			name:       "1% swap fee, no taker fee",
			swapFee:    sdk.NewDecWithPrec(1, 2), // 1%
			takerFee:   sdk.ZeroDec(),            // 0%
			expRevenue: true,
		},
		{
			name:       "0% swap fee, 1% taker fee",
			swapFee:    sdk.ZeroDec(),            // 0%
			takerFee:   sdk.NewDecWithPrec(1, 2), // 1%
			expRevenue: false,
		},
		{
			name:       "0% swap fee, no taker fee",
			swapFee:    sdk.ZeroDec(), // 0%
			takerFee:   sdk.ZeroDec(), // 0%
			expRevenue: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			params := s.App.GAMMKeeper.GetParams(s.Ctx)
			params.TakerFee = tc.takerFee
			s.App.GAMMKeeper.SetParams(s.Ctx, params)

			poolId := s.PrepareCustomPoolFromCoins(poolCoins, balancer.PoolParams{
				SwapFee: tc.swapFee,
				ExitFee: sdk.ZeroDec(),
			})

			// join pool
			addr := sample.Acc()
			s.FundAcc(addr, apptesting.DefaultAcctFunds)
			shares, _ := s.RunBasicJoin(poolId, addr.String())

			// check position
			p, _ := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			pool := p.(*balancer.Pool) // nolint: errcheck
			position, err := pool.CalcExitPoolCoinsFromShares(s.Ctx, shares, sdk.ZeroDec())
			s.Require().NoError(err)
			liquidity := pool.GetTotalPoolLiquidity(s.Ctx)
			spot, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, "foo", "adym")
			s.Require().NoError(err)
			s.T().Logf("positionBefore: %s, liquidity: %s, spot: %s", position, liquidity, spot)

			// swap tokens (swap 5 DYM for FOO) and vice versa
			s.RunBasicSwap(poolId, addr.String(), sdk.NewCoin("adym", apptesting.EXP.Mul(sdk.NewInt(5))), "foo")
			s.RunBasicSwap(poolId, addr.String(), sdk.NewCoin("foo", apptesting.EXP.Mul(sdk.NewInt(5))), "adym")

			// check position
			p, _ = s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			pool = p.(*balancer.Pool) // nolint: errcheck
			liquidity = pool.GetTotalPoolLiquidity(s.Ctx)
			positionAfter, err := pool.CalcExitPoolCoinsFromShares(s.Ctx, shares, sdk.ZeroDec())
			s.Require().NoError(err)
			spot, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, "foo", "adym")
			s.Require().NoError(err)
			s.T().Logf("positionAfterSwap: %s, liquidity: %s, spot: %s", positionAfter, liquidity, spot)

			// assert
			if tc.expRevenue {
				s.True(positionAfter.IsAllGT(position), fmt.Sprintf("positionBefore: %s, positionAfter: %s", position, positionAfter))
			} else {
				s.True(positionAfter.IsAnyGT(position), fmt.Sprintf("positionBefore: %s, positionAfter: %s", position, positionAfter))
			}
		})
	}
}
