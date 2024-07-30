package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *RollappTestSuite) createRollappAndVerify(numOfAddresses int, expectedErr error) types.RollappSummary {
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	// generate sequences address
	addresses := sample.GenerateAddresses(numOfAddresses)
	// rollapp is the rollapp to create
	rollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             apptesting.GenerateRollappID(),
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
	suite.Require().Nil(err)
	if queryResponse.Rollapp.PermissionedAddresses == nil {
		queryResponse.Rollapp.PermissionedAddresses = []string{}
	}
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

	rollapp := types.MsgCreateRollapp{
		Creator:   alice,
		RollappId: "rollapp_1234-1",
	}

	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().NoError(err)

	tests := []struct {
		name      string
		rollappId string
		expErr    error
		malleate  func()
	}{
		{
			name:      "same ID",
			rollappId: "rollapp_1234-1",
			expErr:    types.ErrRollappExists,
		}, {
			name:      "same EIP155, different name",
			rollappId: "trollapp_1234-1",
			expErr:    types.ErrRollappExists,
		}, {
			name:      "same name, different EIP155",
			rollappId: "rollapp_2345-1",
			expErr:    types.ErrRollappExists,
		}, {
			name:      "same ID, forked",
			rollappId: "rollapp_1234-2",
			malleate: func() {
				r := rollapp.GetRollapp()
				r.Frozen = true
				suite.App.RollappKeeper.SetRollapp(suite.Ctx, r)
			},
			expErr: nil,
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			newRollapp := types.MsgCreateRollapp{
				Creator:   alice,
				RollappId: test.rollappId,
			}

			if test.malleate != nil {
				test.malleate()
			}

			_, err := suite.msgServer.CreateRollapp(goCtx, &newRollapp)
			suite.Require().ErrorIs(err, test.expErr)
		})
	}
}

func (suite *RollappTestSuite) TestCreateRollappWhenDisabled() {
	suite.SetupTest()

	suite.createRollappAndVerify(1, nil)
	params := suite.App.RollappKeeper.GetParams(suite.Ctx)
	params.RollappsEnabled = false

	suite.App.RollappKeeper.SetParams(suite.Ctx, params)
	suite.createRollappAndVerify(1, types.ErrRollappsDisabled)
}

func (suite *RollappTestSuite) TestCreateRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	tests := []struct {
		name      string
		rollappId string
		expErr    error
	}{
		{
			name:      "default is valid",
			rollappId: "rollapp_1234-1",
			expErr:    nil,
		},
		{
			name:      "too long id",
			rollappId: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "wrong EIP155",
			rollappId: "rollapp_ea2413-1",
			expErr:    types.ErrRollappIDNotEIP155,
		},
		{
			name:      "no EIP155 with revision",
			rollappId: "rollapp-1",
			expErr:    types.ErrRollappIDNotEIP155,
		},
		{
			name:      "starts with dash",
			rollappId: "-1234",
			expErr:    types.ErrRollappIDNotEIP155,
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			rollapp := types.MsgCreateRollapp{
				Creator:               alice,
				RollappId:             test.rollappId,
				MaxSequencers:         1,
				PermissionedAddresses: []string{},
			}

			_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
			if test.expErr == nil {
				suite.Require().NoError(err)
				_, err = types.NewChainID(test.rollappId)
				suite.Require().NoError(err)
			} else {
				suite.Require().ErrorIs(err, test.expErr)
			}
		})
	}
}

func (suite *RollappTestSuite) TestCreateRollappIdRevisionNumber() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	tests := []struct {
		name      string
		rollappId string
		revision  uint64
		valid     bool
	}{
		{
			name:      "revision set with eip155",
			rollappId: "rollapp_1234-1",
			revision:  1,
			valid:     true,
		},
		{
			name:      "revision set without eip155",
			rollappId: "rollapp-3",
			revision:  3,
			valid:     false,
		},
		{
			name:      "revision not set",
			rollappId: "rollapp",
			revision:  0,
			valid:     false,
		},
		{
			name:      "invalid revision",
			rollappId: "rollapp-1-1",
			revision:  0,
			valid:     false,
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			rollapp := types.MsgCreateRollapp{
				Creator:               alice,
				RollappId:             test.rollappId,
				MaxSequencers:         1,
				PermissionedAddresses: []string{},
			}

			_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)

			if test.valid {
				suite.Require().NoError(err)
				id, err := types.NewChainID(test.rollappId)
				suite.Require().NoError(err)
				suite.Require().Equal(test.revision, id.GetRevisionNumber())

			} else {
				suite.Require().ErrorIs(err, types.ErrRollappIDNotEIP155)
			}
		})
	}
}

func (suite *RollappTestSuite) TestForkChainId() {
	tests := []struct {
		name         string
		rollappId    string
		newRollappId string
		valid        bool
	}{
		{
			name:         "valid eip155 id",
			rollappId:    "rollapp_1234-1",
			newRollappId: "rollapp_1234-2",
			valid:        true,
		},
		{
			name:         "non-valid eip155 id",
			rollappId:    "rollapp_1234-1",
			newRollappId: "rollapp_1234-5",
			valid:        false,
		},
		{
			name:         "same eip155 but different name",
			rollappId:    "rollapp_1234-1",
			newRollappId: "rollapy_1234-2",
			valid:        false,
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollappMsg := types.MsgCreateRollapp{
				Creator:               alice,
				RollappId:             test.rollappId,
				MaxSequencers:         1,
				PermissionedAddresses: []string{},
			}

			_, err := suite.msgServer.CreateRollapp(goCtx, &rollappMsg)
			suite.Require().NoError(err)
			rollapp, found := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappMsg.RollappId)
			suite.Require().True(found)
			rollapp.Frozen = true
			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			rollappMsg2 := types.MsgCreateRollapp{
				Creator:               alice,
				RollappId:             test.newRollappId,
				MaxSequencers:         1,
				PermissionedAddresses: []string{},
			}
			_, err = suite.msgServer.CreateRollapp(goCtx, &rollappMsg2)
			if test.valid {
				suite.Require().NoError(err)
				_, found = suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappMsg2.RollappId)
				suite.Require().True(found)
			} else {
				suite.Require().ErrorIs(err, types.ErrInvalidRollappID)
			}
		})
	}
}

func (suite *RollappTestSuite) TestOverwriteEIP155Key() {
	tests := []struct {
		name         string
		rollappId    string
		badRollappId string
	}{
		{
			name:         "extra whitespace id",
			rollappId:    "rollapp_1234-1",
			badRollappId: "rollapp_1234-1  ",
		},
		{
			name:         "same EIP ID",
			rollappId:    "rollapp_1234-1",
			badRollappId: "dummy_1234-1",
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()
			goCtx := sdk.WrapSDKContext(suite.Ctx)
			rollapp := types.MsgCreateRollapp{
				Creator:               alice,
				RollappId:             test.rollappId,
				MaxSequencers:         1,
				PermissionedAddresses: []string{},
			}
			_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
			suite.Require().NoError(err)

			id, err := types.NewChainID(test.rollappId)
			suite.Require().NoError(err)
			suite.Require().NotEqual(0, id.GetEIP155ID())
			// eip155 key registers to correct roll app
			gotRollapp, found := suite.App.RollappKeeper.GetRollapp(suite.Ctx, test.rollappId)
			suite.Require().True(found)
			suite.Require().Equal(gotRollapp.RollappId, rollapp.RollappId)
			// create bad rollapp
			badrollapp := types.MsgCreateRollapp{
				Creator:               alice,
				RollappId:             test.badRollappId,
				MaxSequencers:         1,
				PermissionedAddresses: []string{},
			}
			_, err = suite.msgServer.CreateRollapp(goCtx, &badrollapp)
			// it should not be possible to register rollapp name with extra space
			suite.Require().ErrorIs(err, types.ErrRollappExists)
		})
	}
}
