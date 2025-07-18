package keeper_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T())
	ctx := app.NewContext(false)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(*app.IROKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.msgServer = keeper.NewMsgServerImpl(*app.IROKeeper)
	suite.Ctx = ctx
	suite.queryClient = queryClient

	// fund alice, the default rollapp creator, so it would have enough balance for IRO creation fee
	rollappId := suite.CreateDefaultRollapp()
	rollapp := suite.App.RollappKeeper.MustGetRollapp(suite.Ctx, rollappId)
	funds := suite.App.IROKeeper.GetParams(suite.Ctx).CreationFee.Mul(math.NewInt(10)) // 10 times the creation fee
	suite.FundAcc(sdk.MustAccAddressFromBech32(rollapp.Owner), sdk.NewCoins(sdk.NewCoin("adym", funds)))

	// set txfees basedenom
	err := suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "adym")
	suite.Require().NoError(err)
}

// BuySomeTokens buys some tokens from the plan
func (suite *KeeperTestSuite) BuySomeTokens(planId string, buyer sdk.AccAddress, amt math.Int) {
	maxAmt := math.NewInt(1_000_000_000).MulRaw(1e18)
	suite.FundAcc(buyer, sdk.NewCoins(sdk.NewCoin("adym", amt.MulRaw(10)))) // 10 times the amount to buy, for buffer and fees
	err := suite.App.IROKeeper.Buy(suite.Ctx, planId, buyer, amt, maxAmt)
	suite.Require().NoError(err)
}
