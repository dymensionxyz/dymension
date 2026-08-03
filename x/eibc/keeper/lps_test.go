package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) fulfillWinner(orderSeq uint64, headerHash []byte, rng int64, addrs []sdk.AccAddress) uint64 {
	suite.SetupTest()
	suite.Ctx = suite.Ctx.WithBlockHeight(10).WithHeaderHash(headerHash)
	for _, addr := range addrs {
		apptesting.FundAccount(suite.App, suite.Ctx, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(1_000_000))))
	}

	k := suite.App.EIBCKeeper
	for _, addr := range addrs[1:] {
		_, err := k.LPs.Create(suite.Ctx, &types.OnDemandLP{
			FundsAddr:  addr.String(),
			Rollapp:    rollappPacket.RollappId,
			Denom:      sdk.DefaultBondDenom,
			MaxPrice:   math.NewInt(100),
			MinFee:     math.LegacyZeroDec(),
			SpendLimit: math.NewInt(1_000),
		})
		suite.Require().NoError(err)
	}

	orderID := suite.orderWithSeq(orderSeq, addrs[0].String(), math.NewInt(90), math.NewInt(10))
	_, err := suite.msgServer.TryFulfillOnDemand(suite.Ctx, &types.MsgTryFulfillOnDemand{
		Signer:  addrs[0].String(),
		OrderId: orderID,
		Rng:     rng,
	})
	suite.Require().NoError(err)

	lps, err := k.LPs.GetAll(suite.Ctx)
	suite.Require().NoError(err)
	for _, lp := range lps {
		if lp.Spent.IsPositive() {
			return lp.Id
		}
	}
	suite.FailNow("no LP recorded spend")
	return 0
}

func (suite *KeeperTestSuite) TestTryFulfillOnDemandIgnoresCallerRNG() {
	addrs := apptesting.CreateRandomAccounts(3)
	winnerFromZero := suite.fulfillWinner(101, []byte("fixed-header"), 0, addrs)
	winnerFromOne := suite.fulfillWinner(101, []byte("fixed-header"), 1, addrs)
	suite.Require().Equal(winnerFromZero, winnerFromOne)
}

func (suite *KeeperTestSuite) TestTryFulfillOnDemandSeedDeterminismAndOrderVariation() {
	addrs := apptesting.CreateRandomAccounts(3)
	headerHash := []byte("fixed-header")
	winner := suite.fulfillWinner(201, headerHash, 0, addrs)
	suite.Require().Equal(winner, suite.fulfillWinner(201, headerHash, 0, addrs))

	winners := map[uint64]struct{}{}
	for seq := uint64(201); seq < 217; seq++ {
		winners[suite.fulfillWinner(seq, headerHash, 0, addrs)] = struct{}{}
	}
	suite.Require().Greater(len(winners), 1)
}

func (suite *KeeperTestSuite) TestTryFulfillOnDemandEmptyHeaderHashDeterministic() {
	addrs := apptesting.CreateRandomAccounts(3)
	winner := suite.fulfillWinner(301, nil, -1, addrs)
	suite.Require().Equal(winner, suite.fulfillWinner(301, nil, 42, addrs))
}
