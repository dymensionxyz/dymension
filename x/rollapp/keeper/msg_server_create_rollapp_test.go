package keeper_test

import (
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *RollappTestSuite) TestCreateRollapp() {
	suite.SetupTest()
	suite.createRollapp(nil)
}

func (suite *RollappTestSuite) TestCreateRollappUnauthorizedRollappCreator() {
	suite.SetupTest()
	suite.createRollappWithCreatorAndVerify(types.ErrFeePayment, bob) // bob is broke
}

func (suite *RollappTestSuite) TestCreateRollappAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	rollapp := types.MsgCreateRollapp{
		Creator:          alice,
		RollappId:        "rollapp_1234-1",
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
		Alias:            "Rollapp",
		VmType:           types.Rollapp_EVM,
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
				Creator:      alice,
				RollappId:    test.rollappId,
				Bech32Prefix: "rol",
				VmType:       types.Rollapp_EVM,
				Alias:        strings.ToLower(rand.Str(3)),
			}

			if test.malleate != nil {
				test.malleate()
			}

			_, err := suite.msgServer.CreateRollapp(goCtx, &newRollapp)
			suite.Require().ErrorIs(err, test.expErr)
		})
	}
}

func (suite *RollappTestSuite) TestCreateRollappAliasAlreadyExists() {
	suite.T().Skip() // TODO: delete test
	suite.SetupTest()

	goCtx := sdk.WrapSDKContext(suite.Ctx)
	alias := "rollapp"

	rollapp := types.MsgCreateRollapp{
		Creator:          alice,
		RollappId:        urand.RollappID(),
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
		Alias:            alias,
		VmType:           types.Rollapp_EVM,
	}
	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().Nil(err)

	rollapp = types.MsgCreateRollapp{
		Creator:          alice,
		RollappId:        urand.RollappID(),
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
		Alias:            alias,
		VmType:           types.Rollapp_EVM,
	}
	_, err = suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.ErrorIs(err, types.ErrRollappAliasExists)
}

func (suite *RollappTestSuite) TestCreateRollappId() {
	suite.SetupTest()

	tests := []struct {
		name      string
		rollappId string
		revision  uint64
		expErr    error
	}{
		{
			name:      "default is valid",
			rollappId: "rollapp_1234-1",
			revision:  1,
			expErr:    nil,
		},
		{
			name:      "too long id",
			rollappId: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz_1234-1",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "wrong EIP155",
			rollappId: "rollapp_ea2413-1",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "no EIP155 with revision",
			rollappId: "rollapp-1",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "starts with dash",
			rollappId: "-1234",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "revision set without eip155",
			rollappId: "rollapp-3",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "revision not set",
			rollappId: "rollapp",
			expErr:    types.ErrInvalidRollappID,
		},
		{
			name:      "invalid revision",
			rollappId: "rollapp-1-1",
			expErr:    types.ErrInvalidRollappID,
		},
	}
	for _, test := range tests {
		suite.Run(test.name, func() {
			id, err := types.NewChainID(test.rollappId)
			suite.Require().ErrorIs(err, test.expErr)
			suite.Require().Equal(test.revision, id.GetRevisionNumber())
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
				Creator:          alice,
				RollappId:        test.rollappId,
				InitialSequencer: sample.AccAddress(),
				Bech32Prefix:     "rol",
				GenesisChecksum:  "checksum",
				Alias:            "Rollapp1",
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
			}

			_, err := suite.msgServer.CreateRollapp(goCtx, &rollappMsg)
			suite.Require().NoError(err)
			rollapp, found := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappMsg.RollappId)
			suite.Require().True(found)
			rollapp.Frozen = true
			suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

			rollappMsg2 := types.MsgCreateRollapp{
				Creator:          alice,
				RollappId:        test.newRollappId,
				InitialSequencer: sample.AccAddress(),
				Bech32Prefix:     "rol",
				GenesisChecksum:  "checksum1",
				Alias:            "Rollapp2",
				VmType:           types.Rollapp_EVM,
				Metadata:         &mockRollappMetadata,
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
			alias := strings.NewReplacer("_", "", "-", "").Replace(test.rollappId) // reuse rollapp ID to avoid alias conflicts
			rollapp := types.MsgCreateRollapp{
				Creator:          alice,
				RollappId:        test.rollappId,
				InitialSequencer: sample.AccAddress(),
				Bech32Prefix:     "rol",
				GenesisChecksum:  "checksum",
				Alias:            alias,
				VmType:           types.Rollapp_EVM,
			}
			_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
			suite.Require().NoError(err)

			id, err := types.NewChainID(test.rollappId)
			suite.Require().NoError(err)
			suite.Require().NotEqual(0, id.GetEIP155ID())

			// eip155 key registers to correct rollapp
			rollappFromEip1155, found := suite.App.RollappKeeper.GetRollappByEIP155(suite.Ctx, id.GetEIP155ID())
			suite.Require().True(found)
			suite.Require().Equal(rollappFromEip1155.RollappId, rollapp.RollappId)

			// create bad rollapp
			badRollapp := types.MsgCreateRollapp{
				Creator:          alice,
				RollappId:        test.badRollappId,
				InitialSequencer: sample.AccAddress(),
				Bech32Prefix:     "rol",
				GenesisChecksum:  "checksum",
				Alias:            "alias",
				VmType:           types.Rollapp_EVM,
			}
			_, err = suite.msgServer.CreateRollapp(goCtx, &badRollapp)
			// it should not be possible to register rollapp name with extra space
			suite.Require().ErrorIs(err, types.ErrRollappExists)
		})
	}
}

func (suite *RollappTestSuite) createRollapp(expectedErr error) {
	// rollappsExpect is the expected result of query all
	var rollappsExpect []*types.RollappSummary

	// test 10 rollapp creations
	for i := 0; i < 10; i++ {
		res := suite.createRollappAndVerify(expectedErr)
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

func (suite *RollappTestSuite) createRollappAndVerify(expectedErr error) types.RollappSummary {
	return suite.createRollappWithCreatorAndVerify(expectedErr, alice)
}

func (suite *RollappTestSuite) createRollappWithCreatorAndVerify(expectedErr error, creator string) types.RollappSummary {
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	// generate sequencer address
	address := sample.AccAddress()
	// rollapp is the rollapp to create
	rollappID := fmt.Sprintf("%s%d", "rollapp", rand.Int63())         //nolint:gosec // this is for a test
	alias := strings.NewReplacer("_", "", "-", "").Replace(rollappID) // reuse rollapp ID to avoid alias conflicts

	rollapp := types.MsgCreateRollapp{
		Creator:          creator,
		RollappId:        urand.RollappID(),
		InitialSequencer: address,
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
		Alias:            alias,
		VmType:           types.Rollapp_EVM,
		Metadata:         &mockRollappMetadata,
	}
	// rollappExpect is the expected result of creating rollapp
	rollappExpect := types.Rollapp{
		RollappId:        rollapp.GetRollappId(),
		Creator:          rollapp.GetCreator(),
		InitialSequencer: rollapp.GetInitialSequencer(),
		GenesisChecksum:  rollapp.GetGenesisChecksum(),
		Bech32Prefix:     rollapp.GetBech32Prefix(),
		Alias:            rollapp.GetAlias(),
		VmType:           types.Rollapp_EVM,
		Metadata:         rollapp.GetMetadata(),
	}
	// create rollapp
	createResponse, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	if expectedErr != nil {
		suite.ErrorIs(err, expectedErr)
		return types.RollappSummary{}
	}
	suite.Require().Nil(err)
	suite.Require().EqualValues(types.MsgCreateRollappResponse{}, *createResponse)

	// query the specific rollapp
	queryResponse, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
		RollappId: rollapp.GetRollappId(),
	})
	suite.Require().Nil(err)
	suite.Require().EqualValues(&rollappExpect, &queryResponse.Rollapp)

	// query the specific rollapp by alias
	queryResponse, err = suite.queryClient.RollappByAlias(goCtx, &types.QueryGetRollappByAliasRequest{
		Alias: rollapp.GetAlias(),
	})
	suite.Require().Nil(err)
	suite.Require().EqualValues(&rollappExpect, &queryResponse.Rollapp)

	rollappSummaryExpect := types.RollappSummary{
		RollappId: rollappExpect.RollappId,
	}
	return rollappSummaryExpect
}

var mockRollappMetadata = types.RollappMetadata{
	Website:          "https://dymension.xyz",
	Description:      "Sample description",
	LogoDataUri:      "data:image/png;base64,c2lzZQ==",
	TokenLogoDataUri: "data:image/png;base64,ZHVwZQ==",
	Telegram:         "rolly",
	X:                "rolly",
}
