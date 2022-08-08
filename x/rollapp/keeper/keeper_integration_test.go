package keeper_test

import (
	"fmt"

	"testing"

	dymensionapp "github.com/dymensionxyz/dymension/app"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/x/sequencer/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

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

type IntegrationTestSuite struct {
	suite.Suite

	app         *dymensionapp.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := dymensionapp.Setup(false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	rollappModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.RollappKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *IntegrationTestSuite) TestCreateRollapp() {
	suite.SetupTest()
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
			PermissionedAddresses: sequencertypes.Sequencers{Addresses: addresses},
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

func (suite *IntegrationTestSuite) TestCreateRollappAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// rollapp is the rollapp to create
	rollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             "rollapp1",
		CodeStamp:             "",
		GenesisPath:           "",
		MaxWithholdingBlocks:  1,
		MaxSequencers:         1,
		PermissionedAddresses: sequencertypes.Sequencers{},
	}
	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.EqualError(err, types.ErrRollappExists.Error())
}

//-------------------------------------------------------------------------------------------------------------------------------

// vereifyAll receives a list of expected results and a map of rollapId->rollapp
// the function verifies that the map contains all the rollapps that are in the list and only them
func vereifyAll(suite *IntegrationTestSuite, rollappsExpect []*types.Rollapp, rollappsRes map[string]*types.Rollapp) {
	// check number of items are equal
	suite.Require().EqualValues(len(rollappsExpect), len(rollappsRes))
	for i := 0; i < len(rollappsExpect); i++ {
		rollappExpect := rollappsExpect[i]
		rollappRes := rollappsRes[rollappExpect.GetRollappId()]
		// println("rollappId:", rollappExpect.GetRollappId(), "=>", "rollapp:", rollappExpect.String())
		suite.Require().EqualValues(&rollappExpect, &rollappRes)
	}
}

// getAll queries for all exsisting rollapps and returns a tuple of:
// map of rollappId->rollapp and the number of retrieved rollapps
func getAll(suite *IntegrationTestSuite) (map[string]*types.Rollapp, int) {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	totalChecked := 0
	totalRes := 0
	nextKey := []byte{}
	rollappsRes := make(map[string]*types.Rollapp)
	for {
		queryAllResponse, err := suite.queryClient.RollappAll(goCtx,
			&types.QueryAllRollappRequest{
				Pagination: &query.PageRequest{
					Key:        nextKey,
					Offset:     0,
					Limit:      0,
					CountTotal: true,
					Reverse:    false,
				}})
		suite.Require().Nil(err)

		if totalRes == 0 {
			totalRes = int(queryAllResponse.GetPagination().GetTotal())
		}

		for i := 0; i < len(queryAllResponse.GetRollapp()); i++ {
			rollappRes := queryAllResponse.GetRollapp()[i]
			if rollappRes.PermissionedAddresses.Addresses == nil {
				rollappRes.PermissionedAddresses.Addresses = []string{}
			}
			rollappsRes[rollappRes.GetRollappId()] = &rollappRes
		}
		totalChecked += len(queryAllResponse.GetRollapp())
		nextKey = queryAllResponse.GetPagination().GetNextKey()

		if nextKey == nil {
			break
		}
	}

	return rollappsRes, totalRes
}
