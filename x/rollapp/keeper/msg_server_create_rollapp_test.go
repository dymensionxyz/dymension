package keeper_test

import (
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *RollappTestSuite) TestCreateRollapp() {
	suite.createRollappFromWhitelist(nil, nil)
}

func (suite *RollappTestSuite) TestCreateRollappFromWhitelist() {
	suite.createRollappFromWhitelist(nil, []types.DeployerParams{{Address: alice, MaxRollapps: 0}})
}

func (suite *RollappTestSuite) TestCreateRollappUnauthorizedRollappCreator() {
	suite.createRollappFromWhitelist(types.ErrUnauthorizedRollappCreator, []types.DeployerParams{{Address: bob, MaxRollapps: 0}})
}

func (suite *RollappTestSuite) TestCreateRollappAlreadyExists() {
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
		PermissionedAddresses: sharedtypes.Sequencers{},
	}
	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.EqualError(err, types.ErrRollappExists.Error())
}
