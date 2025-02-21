package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// create some lps and for a given order, find the compatible lps
func (suite *KeeperTestSuite) TestLPFindCompatible() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "1", // wrong rollup
		Denom:      "aaa",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "1", // wrong rollup
		Denom:      "bbb",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "aaa", // wrong denom
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(1), // max price too low
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(5),
		MinFee:     math.NewInt(8), // min fee too high
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	expect, err := k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(5), // valid
		MinFee:     math.NewInt(7), // valid
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	expect1, err := k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "2",
		Denom:      "bbb",
		MaxPrice:   math.NewInt(6), // also valid, but not first
		MinFee:     math.NewInt(5), // also valid, but not first
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "3", // wrong rollup
		Denom:      "aaa",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)

	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:    "3", // wrong rollup
		Denom:      "bbb",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	o := types.DemandOrder{
		Price:     sdk.NewCoins(sdk.NewCoin("bbb", math.NewInt(5))),
		Fee:       sdk.NewCoins(sdk.NewCoin("bbb", math.NewInt(7))),
		RollappId: "2",
	}
	lps, err := k.LPs.GetOrderCompatibleLPs(ctx, o)
	suite.Require().NoError(err)
	suite.Equal(2, len(lps))
	suite.Equal(expect, lps[0].Id)
	suite.Equal(expect1, lps[1].Id)
}

// test the order age compatibility
// not practical due to test in other test as get explosion of combinations
func (suite *KeeperTestSuite) TestLPCompatibilityHeightAge() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	_, err = k.LPs.Create(ctx, &types.OnDemandLP{
		Rollapp:           "1", // wrong rollup
		Denom:             "aaa",
		SpendLimit:        math.NewInt(100),
		MaxPrice:          math.NewInt(100),
		OrderMinAgeBlocks: 50, // practical value of 5 mins with 6 secs per block, although doesn't matter for test
	})
	suite.Require().NoError(err)
	o := types.DemandOrder{
		RollappId:      "1",
		Price:          sdk.NewCoins(sdk.NewCoin("aaa", math.NewInt(7))),
		Fee:            sdk.NewCoins(sdk.NewCoin("aaa", math.NewInt(7))),
		CreationHeight: 20,
	}
	for i := 68; i < 72; i++ {
		lps, err := k.LPs.GetOrderCompatibleLPs(ctx.WithBlockHeight(int64(i)), o)
		suite.Require().NoError(err)
		if i < 70 {
			suite.Empty(lps)
		} else {
			suite.NotEmpty(lps)
		}
	}
}

func (suite *KeeperTestSuite) TestLPQueriesByAddr() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	addrs := []string{
		"dym1ra6le06p8lle3q6gnsmwz769t2kqld9pmden5k",
		"dym10j59k4whfvtu5flc3lypsjmcyx3fn57ygw78du",
	}
	for i := range 6 {
		_, err = k.LPs.Create(ctx, &types.OnDemandLP{
			Rollapp:   "1",
			Denom:     "bbb",
			FundsAddr: addrs[i%2],
		})
		suite.Require().NoError(err)
	}
	for _, addr := range addrs {
		lps, err := k.LPs.GetByAddr(ctx, sdk.MustAccAddressFromBech32(addr))
		suite.Require().NoError(err)
		suite.Equal(3, len(lps))
		for _, lp := range lps {
			suite.Equal(addr, lp.Lp.FundsAddr)
		}
	}
}
