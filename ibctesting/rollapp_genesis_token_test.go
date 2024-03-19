package ibctesting_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/suite"
)

var (
	genesisAuthorizedAccount = apptesting.CreateRandomAccounts(1)[0]
	rollappDenom             = "arax"
)

type RollappGenesisTokenTestSuite struct {
	IBCTestUtilSuite

	msgServer types.MsgServer
	ctx       sdk.Context
}

func TestRollappGenesisTokenTestSuite(t *testing.T) {
	suite.Run(t, new(RollappGenesisTokenTestSuite))
}

func (suite *RollappGenesisTokenTestSuite) SetupTest() {
	suite.IBCTestUtilSuite.SetupTest()
	rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
	suite.msgServer = rollappkeeper.NewMsgServerImpl(rollappKeeper)
	suite.ctx = suite.hubChain.GetContext()
}

func (suite *RollappGenesisTokenTestSuite) TestTriggerGenesisEvent() {
	suite.CreateRollapp()
	// Create a primary path
	hubToRollappPath := suite.NewTransferPath(suite.hubChain, suite.rollappChain)
	suite.coordinator.Setup(hubToRollappPath)
	// Create a secondary path with a 3rd party chain
	hubToCosmosPath := suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
	suite.coordinator.Setup(hubToCosmosPath)

	cases := []struct {
		name           string
		gensisState    *types.RollappGenesisState
		msg            *types.MsgRollappGenesisEvent
		deployerParams []types.DeployerParams
		expErr         error
	}{
		{
			"successful rollapp genesis event",
			&types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			&types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			[]types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			nil,
		},
		{
			"invalid rollapp genesis event - genesis event already triggered",
			&types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: true,
			},
			&types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			[]types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			types.ErrGenesisEventAlreadyTriggered,
		},
		{
			"invalid rollapp genesis event - unauthorized address",
			&types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: true,
			},
			&types.MsgRollappGenesisEvent{
				Address:   apptesting.CreateRandomAccounts(1)[0].String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			[]types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			sdkerrors.ErrUnauthorized,
		},
		{
			"invalid rollapp genesis event - rollapp doesn't exist",
			&types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			&types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: "someRandomChainID",
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			[]types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			types.ErrUnknownRollappID,
		},
		{
			"invalid rollapp genesis event - channel doesn't exist",
			&types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			&types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: "SomeRandomChannelID",
			},
			[]types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			channeltypes.ErrChannelNotFound,
		},
		{
			"invalid rollapp genesis event - channel id doesn't match chain id",
			&types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			&types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToCosmosPath.EndpointA.ChannelID,
			},
			[]types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			types.ErrInvalidGenesisChannelId,
		},
	}
	for _, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// Setup the deployer whitelist
			rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
			rollappKeeper.SetParams(suite.ctx, types.NewParams(true, 2, tc.deployerParams))
			// Setup the rollapp genesis state
			rollapp, found := rollappKeeper.GetRollapp(suite.ctx, suite.rollappChain.ChainID)
			suite.Require().True(found)
			rollapp.GenesisState = tc.gensisState
			rollappKeeper.SetRollapp(suite.ctx, rollapp)
			// Send the genesis event
			_, err := suite.msgServer.TriggerGenesisEvent(suite.ctx, tc.msg)
			suite.hubChain.NextBlock()
			suite.Require().ErrorIs(err, tc.expErr)
			// Validate no tokens are in the module account
			accountKeeper := ConvertToApp(suite.hubChain).AccountKeeper
			bankKeeper := ConvertToApp(suite.hubChain).BankKeeper
			moduleAcc := accountKeeper.GetModuleAccount(suite.ctx, types.ModuleName)
			suite.Require().Equal(sdk.NewCoins(), bankKeeper.GetAllBalances(suite.ctx, moduleAcc.GetAddress()))
			// Validate the genesis accounts balances
			rollappIBCDenom := utils.GetForeignIBCDenom(hubToRollappPath.EndpointB.ChannelID, rollappDenom)
			for _, roallppGenesisAccount := range tc.gensisState.GenesisAccounts {
				balance := bankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(roallppGenesisAccount.Address), rollappIBCDenom)
				if tc.expErr != nil {
					suite.Require().Equal(sdk.NewCoin(rollappIBCDenom, sdk.NewInt(0)), balance)
				} else {
					suite.Require().Equal(roallppGenesisAccount.Amount.Amount, balance.Amount)
				}

				denomMetaData, found := bankKeeper.GetDenomMetaData(suite.ctx, rollappIBCDenom)
				suite.Require().True(found)
				suite.Require().Equal(rollappDenom, denomMetaData.Display)
			}
		})
	}
}
