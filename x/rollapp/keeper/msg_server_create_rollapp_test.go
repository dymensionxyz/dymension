package keeper_test

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cometbft/cometbft/libs/rand"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *RollappTestSuite) TestCreateRollapp() {
	s.SetupTest()
	s.createRollapp(nil)
}

func (s *RollappTestSuite) TestCreateRollappUnauthorizedRollappCreator() {
	s.SetupTest()
	s.createRollappWithCreatorAndVerify(sdkerrors.ErrInsufficientFunds, bob, false) // bob is broke
}

func (s *RollappTestSuite) TestCreateRollappWithBechGenesisSum() {
	s.SetupTest()
	goCtx := sdk.WrapSDKContext(s.Ctx)

	rollapp := types.MsgCreateRollapp{
		Creator:          alice,
		RollappId:        "rollapp_1234-1",
		InitialSequencer: sample.AccAddress(),
		MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,

		Alias:       "rollapp",
		VmType:      types.Rollapp_EVM,
		GenesisInfo: mockGenesisInfo,
	}
	_, err := s.msgServer.CreateRollapp(goCtx, &rollapp)
	s.Require().NoError(err)
}

func (s *RollappTestSuite) TestCreateRollappAlreadyExists() {
	s.SetupTest()
	goCtx := sdk.WrapSDKContext(s.Ctx)

	rollapp := types.MsgCreateRollapp{
		Creator:          alice,
		RollappId:        "rollapp_1234-1",
		InitialSequencer: sample.AccAddress(),
		MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,

		Alias:  "rollapp",
		VmType: types.Rollapp_EVM,
	}

	_, err := s.msgServer.CreateRollapp(goCtx, &rollapp)
	s.Require().NoError(err)

	tests := []struct {
		name      string
		rollappId string
		expErr    error
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
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			newRollapp := types.MsgCreateRollapp{
				Creator:     alice,
				RollappId:   test.rollappId,
				VmType:      types.Rollapp_EVM,
				Alias:       strings.ToLower(rand.Str(7)),
				GenesisInfo: mockGenesisInfo,
			}

			s.FundForAliasRegistration(newRollapp)

			_, err := s.msgServer.CreateRollapp(goCtx, &newRollapp)
			s.Require().ErrorIs(err, test.expErr)
		})
	}
}

func (s *RollappTestSuite) TestOverwriteEIP155Key() {
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
		s.Run(test.name, func() {
			s.SetupTest()
			goCtx := sdk.WrapSDKContext(s.Ctx)
			rollapp := types.MsgCreateRollapp{
				Creator:          alice,
				RollappId:        test.rollappId,
				InitialSequencer: sample.AccAddress(),
				MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,

				Alias:       "alias",
				VmType:      types.Rollapp_EVM,
				GenesisInfo: mockGenesisInfo,
			}
			s.FundForAliasRegistration(rollapp)
			_, err := s.msgServer.CreateRollapp(goCtx, &rollapp)
			s.Require().NoError(err)

			id, err := types.NewChainID(test.rollappId)
			s.Require().NoError(err)
			s.Require().NotEqual(0, id.GetEIP155ID())

			// eip155 key registers to correct rollapp
			rollappFromEip1155, found := s.k().GetRollappByEIP155(s.Ctx, id.GetEIP155ID())
			s.Require().True(found)
			s.Require().Equal(rollappFromEip1155.RollappId, rollapp.RollappId)

			// create bad rollapp
			badRollapp := types.MsgCreateRollapp{
				Creator:          alice,
				RollappId:        test.badRollappId,
				InitialSequencer: sample.AccAddress(),
				Alias:            "alias",
				VmType:           types.Rollapp_EVM,
				GenesisInfo:      mockGenesisInfo,
			}
			s.FundForAliasRegistration(rollapp)
			_, err = s.msgServer.CreateRollapp(goCtx, &badRollapp)
			// it should not be possible to register rollapp name with extra space
			s.Require().ErrorIs(err, types.ErrRollappExists)
		})
	}
}

func (s *RollappTestSuite) createRollapp(expectedErr error) {
	// rollappsExpect is the expected result of query all
	var rollappsExpect []*types.RollappSummary

	// test 10 rollapp creations
	for i := 0; i < 10; i++ {
		res := s.createRollappAndVerify(expectedErr)
		rollappsExpect = append(rollappsExpect, &res)
	}

	// verify that query all contains all the rollapps that were created
	rollappsRes, totalRes := getAll(s)
	if expectedErr != nil {
		s.Require().EqualValues(totalRes, 0)
		return
	} else {
		s.Require().EqualValues(totalRes, 10)
		verifyAll(s, rollappsExpect, rollappsRes)
	}
}

func (s *RollappTestSuite) createRollappAndVerify(expectedErr error) types.RollappSummary {
	return s.createRollappWithCreatorAndVerify(expectedErr, alice, true)
}

func (s *RollappTestSuite) createRollappWithCreatorAndVerify(
	expectedErr error, creator string, fundAccount bool,
) types.RollappSummary {
	goCtx := sdk.WrapSDKContext(s.Ctx)
	// generate sequencer address
	address := sample.AccAddress()
	// rollapp is the rollapp to create
	rollapp := types.MsgCreateRollapp{
		Creator:          creator,
		RollappId:        urand.RollappID(),
		InitialSequencer: address,
		MinSequencerBond: types.DefaultMinSequencerBondGlobalCoin,
		Alias:            strings.ToLower(rand.Str(7)),
		VmType:           types.Rollapp_EVM,
		Metadata:         &mockRollappMetadata,
		GenesisInfo:      mockGenesisInfo,
	}
	if fundAccount {
		s.FundForAliasRegistration(rollapp)
	}
	// rollappExpect is the expected result of creating rollapp
	rollappExpect := types.Rollapp{
		RollappId:        rollapp.GetRollappId(),
		Owner:            rollapp.GetCreator(),
		InitialSequencer: rollapp.GetInitialSequencer(),
		MinSequencerBond: sdk.NewCoins(types.DefaultMinSequencerBondGlobalCoin),
		VmType:           types.Rollapp_EVM,
		Metadata:         rollapp.GetMetadata(),
		GenesisInfo:      *rollapp.GetGenesisInfo(),
		Revisions: []types.Revision{{
			Number:      0,
			StartHeight: 0,
		}},
	}

	// create rollapp
	createResponse, err := s.msgServer.CreateRollapp(goCtx, &rollapp)
	if expectedErr != nil {
		s.ErrorIs(err, expectedErr)
		return types.RollappSummary{}
	}
	s.Require().Nil(err)
	s.Require().EqualValues(types.MsgCreateRollappResponse{}, *createResponse)

	// query the specific rollapp
	queryResponse, err := s.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
		RollappId: rollapp.GetRollappId(),
	})
	s.Require().Nil(err)
	s.Require().EqualValues(&rollappExpect, &queryResponse.Rollapp)

	rollappSummaryExpect := types.RollappSummary{
		RollappId: rollappExpect.RollappId,
	}
	return rollappSummaryExpect
}

var mockRollappMetadata = types.RollappMetadata{
	Website:     "https://dymension.xyz",
	Description: "Sample description",
	LogoUrl:     "https://dymension.xyz/logo.png",
	Telegram:    "https://t.me/rolly",
	X:           "https://x.dymension.xyz",
	Tags:        []string{"AI", "DeFi", "NFT"},
}

var mockGenesisInfo = &types.GenesisInfo{
	Bech32Prefix:    "rol",
	GenesisChecksum: "checksum",
	NativeDenom: types.DenomMetadata{
		Display:  "DEN",
		Base:     "aden",
		Exponent: 18,
	},
	InitialSupply: sdk.NewInt(100000000),
}
