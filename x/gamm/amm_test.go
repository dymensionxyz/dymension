package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	"github.com/stretchr/testify/suite"

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
	app := apptesting.Setup(s.T())
	ctx := app.NewContext(false)

	// set txfees basedenom
	err := app.TxFeesKeeper.SetBaseDenom(ctx, "adym")
	s.Require().NoError(err)

	s.App = app
	s.Ctx = ctx
}

func (s *KeeperTestSuite) TestSwapsRevenue() {
	// Create a pool with 100_000 DYM and 100_000 FOO
	fooDenom := "ibc/A88EE35932B15B981676EFA6700342EDEF63C41C9EE1265EA5BEDAE0A6518CEA"
	poolCoins := sdk.NewCoins(
		sdk.NewCoin("adym", apptesting.EXP.Mul(math.NewInt(100_000))),
		sdk.NewCoin(fooDenom, apptesting.EXP.Mul(math.NewInt(100_000))),
	)

	testCases := []struct {
		name       string
		swapFee    math.LegacyDec
		takerFee   math.LegacyDec
		expRevenue bool
	}{
		{
			name:       "1% swap fee, 0.9% taker fee",
			swapFee:    math.LegacyNewDecWithPrec(1, 2), // 1%
			takerFee:   math.LegacyNewDecWithPrec(9, 3), // 0.9%
			expRevenue: true,
		},
		{
			name:       "1% swap fee, no taker fee",
			swapFee:    math.LegacyNewDecWithPrec(1, 2), // 1%
			takerFee:   math.LegacyZeroDec(),            // 0%
			expRevenue: true,
		},
		{
			name:       "0% swap fee, 1% taker fee",
			swapFee:    math.LegacyZeroDec(),            // 0%
			takerFee:   math.LegacyNewDecWithPrec(1, 2), // 1%
			expRevenue: false,
		},
		{
			name:       "0% swap fee, no taker fee",
			swapFee:    math.LegacyZeroDec(), // 0%
			takerFee:   math.LegacyZeroDec(), // 0%
			expRevenue: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			params := s.App.GAMMKeeper.GetParams(s.Ctx)
			params.TakerFee = tc.takerFee
			s.App.GAMMKeeper.SetParams(s.Ctx, params)

			s.FundAcc(sdk.MustAccAddressFromBech32(apptesting.Alice), apptesting.DefaultAcctFunds.Add(sdk.NewCoin(fooDenom, apptesting.EXP.Mul(math.NewInt(1_000_000)))))
			poolId := s.PrepareCustomPoolFromCoins(poolCoins, balancer.PoolParams{
				SwapFee: tc.swapFee,
				ExitFee: math.LegacyZeroDec(),
			})

			// join pool
			addr := sample.Acc()
			s.FundAcc(addr, apptesting.DefaultAcctFunds.Add(sdk.NewCoin(fooDenom, apptesting.EXP.Mul(math.NewInt(1_000_000)))))
			shares, _ := s.RunBasicJoin(poolId, addr.String())

			// check position
			p, _ := s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			pool := p.(*balancer.Pool) // nolint: errcheck
			position, err := pool.CalcExitPoolCoinsFromShares(s.Ctx, shares, math.LegacyZeroDec())
			s.Require().NoError(err)
			liquidity := pool.GetTotalPoolLiquidity(s.Ctx)
			spot, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, fooDenom, "adym")
			s.Require().NoError(err)
			s.T().Logf("positionBefore: %s, liquidity: %s, spot: %s", position, liquidity, spot)

			// swap tokens (swap 5 DYM for FOO) and vice versa
			s.RunBasicSwap(poolId, addr.String(), sdk.NewCoin("adym", apptesting.EXP.Mul(math.NewInt(5))), fooDenom)
			s.RunBasicSwap(poolId, addr.String(), sdk.NewCoin(fooDenom, apptesting.EXP.Mul(math.NewInt(5))), "adym")

			// check position
			p, _ = s.App.GAMMKeeper.GetPool(s.Ctx, poolId)
			pool = p.(*balancer.Pool) // nolint: errcheck
			liquidity = pool.GetTotalPoolLiquidity(s.Ctx)
			positionAfter, err := pool.CalcExitPoolCoinsFromShares(s.Ctx, shares, math.LegacyZeroDec())
			s.Require().NoError(err)
			spot, err = s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, fooDenom, "adym")
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
