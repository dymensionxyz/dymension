package keeper_test

import (
	fmt "fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymensionapp "github.com/dymensionxyz/dymension/app"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

const (
	transferEventCount            = 3 // As emitted by the bank
	createEventCount              = 8
	playEventCountFirst           = 8 // Extra "sender" attribute emitted by the bank
	playEventCountNext            = 7
	rejectEventCount              = 4
	rejectEventCountWithTransfer  = 5 // Extra "sender" attribute emitted by the bank
	forfeitEventCount             = 4
	forfeitEventCountWithTransfer = 5 // Extra "sender" attribute emitted by the bank
	alice                         = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bob                           = "cosmos1xyxs3skf3f4jfqeuv89yyaqvjc6lffavxqhc8g"
	carol                         = "cosmos1e0w5t53nrq7p66fye6c8p0ynyhf6y24l4yuxd7"
	balAlice                      = 50000000
	balBob                        = 20000000
	balCarol                      = 10000000
	foreignToken                  = "foreignToken"
	balTokenAlice                 = 5
	balTokenBob                   = 2
	balTokenCarol                 = 1
)

var (
	rollappModuleAddress string
)

type RollappTestSuite struct {
	suite.Suite

	app         *dymensionapp.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *RollappTestSuite) SetupTest(deployerWhitelist ...types.DeployerParams) {
	app := dymensionapp.Setup(false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.RollappKeeper.SetParams(ctx, types.NewParams(2, deployerWhitelist))
	rollappModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.RollappKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *RollappTestSuite) createRollappFromWhitelist(expectedErr error, deployerWhitelist []types.DeployerParams) {
	suite.SetupTest(deployerWhitelist...)
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// rollappsExpect is the expected result of query all
	rollappsExpect := []*types.Rollapp{}

	// test 10 rollap creations
	for i := 0; i < 10; i++ {
		// generate sequences address
		addresses := sample.GenerateAddresses(i)
		// rollapp is the rollapp to create
		rollapp := types.MsgCreateRollapp{
			Creator:               alice,
			RollappId:             fmt.Sprintf("%s%d", "rollapp", i),
			CodeStamp:             "",
			GenesisPath:           "",
			MaxWithholdingBlocks:  1,
			MaxSequencers:         1,
			PermissionedAddresses: sharedtypes.Sequencers{Addresses: addresses},
		}
		// rollappExpect is the expected result of creating rollapp
		rollappExpect := types.Rollapp{
			RollappId:             rollapp.GetRollappId(),
			Creator:               rollapp.GetCreator(),
			Version:               0,
			CodeStamp:             rollapp.GetCodeStamp(),
			GenesisPath:           rollapp.GetGenesisPath(),
			MaxWithholdingBlocks:  rollapp.GetMaxWithholdingBlocks(),
			MaxSequencers:         rollapp.GetMaxSequencers(),
			PermissionedAddresses: rollapp.GetPermissionedAddresses(),
		}
		// create rollapp
		createResponse, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
		if expectedErr != nil {
			suite.EqualError(err, expectedErr.Error())
			continue
		}
		suite.Require().Nil(err)
		suite.Require().EqualValues(types.MsgCreateRollappResponse{}, *createResponse)

		// query the specific rollapp
		queryResponse, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
			RollappId: rollapp.GetRollappId(),
		})
		if queryResponse.Rollapp.PermissionedAddresses.Addresses == nil {
			queryResponse.Rollapp.PermissionedAddresses.Addresses = []string{}
		}
		suite.Require().Nil(err)
		suite.Require().EqualValues(&rollappExpect, &queryResponse.Rollapp)

		// add the rollapp to the list of get all expected list
		rollappsExpect = append(rollappsExpect, &rollappExpect)
		// verify that query all contains all the rollapps that were created
		rollappsRes, totalRes := getAll(suite)
		suite.Require().EqualValues(totalRes, i+1)
		vereifyAll(suite, rollappsExpect, rollappsRes)

	}

}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RollappTestSuite))
}
