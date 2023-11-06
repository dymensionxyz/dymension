package ibctesting_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	app "github.com/dymensionxyz/dymension/app"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ChainIDPrefix defines the default chain ID prefix for Evmos test chains
var ChainIDPrefix = "evmos_9000-"

func init() {
	ibctesting.ChainIDPrefix = ChainIDPrefix
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		return app.SetupTestingApp()
	}

	// sdk.DefaultPowerReduction = sdk.NewIntFromUint64(1000000)
}

func ConvertToApp(chain *ibctesting.TestChain) *app.App {
	app, ok := chain.App.(*app.App)
	require.True(chain.T, ok)

	return app
}

// KeeperTestSuite is a testing suite to test keeper functions.
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	hubChain     *ibctesting.TestChain
	cosmosChain  *ibctesting.TestChain
	rollappChain *ibctesting.TestChain

	pathHubCosmos   *ibctesting.Path
	pathCosmosEvmos *ibctesting.Path
	pathHubRollapp  *ibctesting.Path
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)               // initializes 2 test chains
	suite.hubChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))     // convenience and readability
	suite.cosmosChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))  // convenience and readability
	suite.rollappChain = suite.coordinator.GetChain(ibctesting.GetChainID(3)) // convenience and readability
}

func (suite *KeeperTestSuite) CreateRollapp() {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		suite.hubChain.SenderAccount.GetAddress().String(),
		suite.rollappChain.ChainID,
		10,
		&sharedtypes.Sequencers{},
		nil,
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed
}

func (suite *KeeperTestSuite) CreateRollappWithMetadata() {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		suite.hubChain.SenderAccount.GetAddress().String(),
		suite.rollappChain.ChainID,
		10,
		&sharedtypes.Sequencers{},
		[]rollapptypes.TokenMetadata{
			{
				Description: "test",
				DenomUnits: []*rollapptypes.DenomUnit{
					{
						Denom:    "utest",
						Exponent: 0,
					},
					{
						Denom:    "utest",
						Exponent: 0,
					},
				},
				Base:    "utest",
				Display: "TST",
				Name:    "test",
				Symbol:  "TST",
			},
		},
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed
}

func (suite *KeeperTestSuite) NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}
