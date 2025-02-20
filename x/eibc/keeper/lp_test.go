package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestLPs() {
	var err error
	k := suite.App.EIBCKeeper
	ctx := suite.Ctx
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "1", // wrong rollup
		Denom:      "1",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "1", // wrong rollup
		Denom:      "2",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "2",
		Denom:      "1", // wrong denom
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "2",
		Denom:      "2",
		MaxPrice:   math.NewInt(1), // max price too low
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "2",
		Denom:      "2",
		MaxPrice:   math.NewInt(5),
		MinFee:     math.NewInt(8), // min fee too high
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	expect, err := k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "2",
		Denom:      "2",
		MaxPrice:   math.NewInt(5), // valid
		MinFee:     math.NewInt(7), // valid
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "2",
		Denom:      "2",
		MaxPrice:   math.NewInt(6), // also valid, but not first
		MinFee:     math.NewInt(5), // also valid, but not first
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "3", // wrong rollup
		Denom:      "1",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)

	_, err = k.LPs.UpsertLP(ctx, &types.OnDemandLiquidity{
		Rollapp:    "3", // wrong rollup
		Denom:      "2",
		MaxPrice:   math.NewInt(1),
		MinFee:     math.NewInt(1),
		SpendLimit: math.NewInt(100),
	})
	suite.Require().NoError(err)
	o := types.DemandOrder{
		Price:     sdk.NewCoins(sdk.NewCoin("2", sdk.NewInt(5))),
		Fee:       sdk.NewCoins(sdk.NewCoin("2", sdk.NewInt(7))),
		RollappId: "2",
	}
	lp, err := k.LPs.FindLP(ctx, k, &o)
	suite.Require().NoError(err)
	suite.Require().NotNil(lp)
	suite.Equal(expect, lp.Id)
}
