package ibctesting_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/suite"
)

var (
	genesisAuthorizedAccount = apptesting.CreateRandomAccounts(1)[0]
	rollappDenom             = "arax"
	rollappIBCDenom          = "ibc/9A1EACD53A6A197ADC81DF9A49F0C4A26F7FF685ACF415EE726D7D59796E71A7"
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
		name             string
		gensisState      *types.RollappGenesisState
		msg              *types.MsgRollappGenesisEvent
		deployerParams   []types.DeployerParams
		expectSavedDenom string
		expErr           string
	}{
		{
			name: "successful rollapp genesis event",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			deployerParams:   []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expectSavedDenom: rollappIBCDenom,
		},
		{
			name: "invalid rollapp genesis event - genesis event already triggered",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: true,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			deployerParams: []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expErr:         types.ErrGenesisEventAlreadyTriggered.Error(),
		},
		{
			name: "invalid rollapp genesis event - unauthorized address",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: true,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   apptesting.CreateRandomAccounts(1)[0].String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			deployerParams: []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expErr:         sdkerrors.ErrUnauthorized.Error(),
		},
		{
			name: "invalid rollapp genesis event - rollapp doesn't exist",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: "someRandomChainID",
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			deployerParams: []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expErr:         types.ErrUnknownRollappID.Error(),
		},
		{
			name: "invalid rollapp genesis event - channel doesn't exist",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: "SomeRandomChannelID",
			},
			deployerParams: []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expErr:         "port-id: transfer, channel-id: SomeRandomChannelID: channel not found",
		},
		{
			name: "invalid rollapp genesis event - channel id doesn't match chain id",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(350))},
					{Address: apptesting.CreateRandomAccounts(1)[0].String(), Amount: sdk.NewCoin(rollappDenom, sdk.NewInt(140))},
				},
				IsGenesisEvent: false,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToCosmosPath.EndpointA.ChannelID,
			},
			deployerParams: []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expErr:         "channel channel-1 is connected to chain ID evmos_9000-2, expected evmos_9000-3: invalid genesis channel id",
		},
		{
			name: "failed rollapp genesis event - error minting coins",
			gensisState: &types.RollappGenesisState{
				GenesisAccounts: []types.GenesisAccount{
					{Address: ""},
				},
				IsGenesisEvent: false,
			},
			msg: &types.MsgRollappGenesisEvent{
				Address:   genesisAuthorizedAccount.String(),
				RollappId: suite.rollappChain.ChainID,
				ChannelId: hubToRollappPath.EndpointA.ChannelID,
			},
			deployerParams:   []types.DeployerParams{{Address: genesisAuthorizedAccount.String()}},
			expectSavedDenom: rollappIBCDenom,
			expErr:           "failed to mint genesis tokens: failed to convert account address: empty address string is not allowed",
		},
	}
	for _, tc := range cases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			// Reset the test state
			defer func() {
				suite.SetupTest()
				suite.CreateRollapp()
				// Create a primary path
				hubToRollappPath = suite.NewTransferPath(suite.hubChain, suite.rollappChain)
				suite.coordinator.Setup(hubToRollappPath)
				// Create a secondary path with a 3rd party chain
				hubToCosmosPath = suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
				suite.coordinator.Setup(hubToCosmosPath)
			}()
			// Setup the deployer whitelist
			rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
			rollappKeeper.SetParams(suite.ctx, types.NewParams(true, 2, tc.deployerParams))
			// Setup the rollapp genesis state
			rollapp, found := rollappKeeper.GetRollapp(suite.ctx, suite.rollappChain.ChainID)
			suite.Require().True(found)
			rollapp.GenesisState = tc.gensisState
			rollappKeeper.SetRollapp(suite.ctx, rollapp)
			// Send the genesis event

			ctx := suite.ctx.WithProposer(suite.hubChain.NextVals.Proposer.Address.Bytes())
			_, err := suite.msgServer.TriggerGenesisEvent(ctx, tc.msg)
			suite.hubChain.NextBlock()
			if tc.expErr != "" {
				suite.Require().EqualError(err, tc.expErr)
			} else {
				suite.Require().NoError(err)
			}
			// Validate no tokens are in the module account
			accountKeeper := ConvertToApp(suite.hubChain).AccountKeeper
			bankKeeper := ConvertToApp(suite.hubChain).BankKeeper
			moduleAcc := accountKeeper.GetModuleAccount(suite.ctx, types.ModuleName)
			suite.Require().Equal(sdk.NewCoins(), bankKeeper.GetAllBalances(suite.ctx, moduleAcc.GetAddress()))
			// Validate the genesis accounts balances
			rollappIBCDenom := utils.GetForeignIBCDenom(hubToRollappPath.EndpointB.ChannelID, rollappDenom)
			for _, roallppGenesisAccount := range tc.gensisState.GenesisAccounts {
				if roallppGenesisAccount.Address != "" {
					balance := bankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(roallppGenesisAccount.Address), rollappIBCDenom)
					if tc.expErr != "" {
						suite.Require().Equal(sdk.NewCoin(rollappIBCDenom, sdk.NewInt(0)), balance)
					} else {
						suite.Require().Equal(roallppGenesisAccount.Amount.Amount, balance.Amount)
					}
				}

				_, found := bankKeeper.GetDenomMetaData(suite.ctx, rollappIBCDenom)
				suite.Require().Equal(tc.expectSavedDenom != "", found)
			}
		})
	}
}
