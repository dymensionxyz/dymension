package keeper_test

import (
	fmt "fmt"
	"math/rand"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *RollappTestSuite) createRollappAndVerify(numOfAddresses int, expectedErr error) types.RollappSummary {
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	// generate sequences address
	addresses := sample.GenerateAddresses(numOfAddresses)
	// rollapp is the rollapp to create
	rollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             fmt.Sprintf("%s%d", "rollapp", rand.Int63()), //nolint:gosec // this is for a test
		MaxSequencers:         uint64(numOfAddresses),
		PermissionedAddresses: addresses,
	}
	// rollappExpect is the expected result of creating rollapp
	rollappExpect := types.Rollapp{
		RollappId:             rollapp.GetRollappId(),
		Creator:               rollapp.GetCreator(),
		Version:               0,
		MaxSequencers:         rollapp.GetMaxSequencers(),
		PermissionedAddresses: rollapp.GetPermissionedAddresses(),
	}
	// create rollapp
	createResponse, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	if expectedErr != nil {
		suite.EqualError(err, expectedErr.Error())
		return types.RollappSummary{}
	}
	suite.Require().Nil(err)
	suite.Require().EqualValues(types.MsgCreateRollappResponse{}, *createResponse)

	// query the specific rollapp
	queryResponse, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
		RollappId: rollapp.GetRollappId(),
	})
	if queryResponse.Rollapp.PermissionedAddresses == nil {
		queryResponse.Rollapp.PermissionedAddresses = []string{}
	}
	suite.Require().Nil(err)
	suite.Require().EqualValues(&rollappExpect, &queryResponse.Rollapp)

	rollappSummaryExpect := types.RollappSummary{
		RollappId: rollappExpect.RollappId,
	}
	return rollappSummaryExpect
}

func (suite *RollappTestSuite) createRollappFromWhitelist(expectedErr error, deployerWhitelist []types.DeployerParams) {
	suite.SetupTest(deployerWhitelist...)

	// rollappsExpect is the expected result of query all
	var rollappsExpect []*types.RollappSummary

	// test 10 rollap creations
	for i := 0; i < 10; i++ {
		res := suite.createRollappAndVerify(i, expectedErr)
		rollappsExpect = append(rollappsExpect, &res)
	}

	// verify that query all contains all the rollapps that were created
	rollappsRes, totalRes := getAll(suite)
	if expectedErr != nil {
		suite.Require().EqualValues(totalRes, 0)
		return
	} else {
		suite.Require().EqualValues(totalRes, 10)
		verifyAll(suite, rollappsExpect, rollappsRes)
	}
}

func (suite *RollappTestSuite) TestCreateRollapp() {
	suite.createRollappFromWhitelist(nil, nil)
}

func (suite *RollappTestSuite) TestCreateRollappFromWhitelist() {
	suite.createRollappFromWhitelist(nil, []types.DeployerParams{{Address: alice}})
}

func (suite *RollappTestSuite) TestCreateRollappUnauthorizedRollappCreator() {
	suite.createRollappFromWhitelist(types.ErrUnauthorizedRollappCreator, []types.DeployerParams{{Address: bob}})
}

func (suite *RollappTestSuite) TestCreateRollappAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// rollapp is the rollapp to create
	rollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             "rollapp1",
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.EqualError(err, types.ErrRollappExists.Error())
}

func (suite *RollappTestSuite) TestCreateRollappWhenDisabled() {
	suite.SetupTest()

	suite.createRollappAndVerify(1, nil)
	params := suite.App.RollappKeeper.GetParams(suite.Ctx)
	params.RollappsEnabled = false

	suite.App.RollappKeeper.SetParams(suite.Ctx, params)
	suite.createRollappAndVerify(1, types.ErrRollappsDisabled)
}

func (suite *RollappTestSuite) TestOverwriteEIP155Key() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	rollappId := "rollapp_1234-1"     // without whitespace
	badrollappId := "rollapp_1234-1 " // with whitespace
	// create rollapp with normal ID
	rollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             rollappId,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().NoError(err)

	// get eip155 key
	id, err := types.NewChainID(rollappId)

	suite.Require().NoError(err)
	suite.Require().NotNil(id.EIP155ID)
	eip155key := id.EIP155ID.Uint64()
	// eip155 key registers to correct roll app
	rollAppfromEip1155, found := suite.App.RollappKeeper.GetRollappByEIP155(suite.Ctx, eip155key)
	suite.Require().True(found)
	suite.Require().Equal(rollAppfromEip1155.RollappId, rollapp.RollappId)
	// create bad rollapp
	badrollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             badrollappId,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	_, err = suite.msgServer.CreateRollapp(goCtx, &badrollapp)
	// it should not be possible to register rollapp name with extra space
	suite.Require().ErrorIs(err, types.ErrRollappExists)
}
