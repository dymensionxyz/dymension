package ibctesting_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	app "github.com/dymensionxyz/dymension/v3/app"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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
		[]string{},
		nil,
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed

	rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
	ctx := suite.hubChain.GetContext()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: suite.rollappChain.ChainID, Index: 1}
	stateInfo := rollapptypes.StateInfo{
		StateInfoIndex: stateInfoIdx,
		StartHeight:    0,
		NumBlocks:      uint64(ctx.BlockHeader().Height - 1),
		Status:         rollapptypes.STATE_STATUS_FINALIZED,
	}

	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
	// uppdate the LatestStateInfoIndex of the rollapp
	rollappKeeper.SetLatestFinalizedStateIndex(ctx, stateInfoIdx)
}

func (suite *KeeperTestSuite) CreateRollappWithMetadata(denom string) {
	displayDenom := "big" + denom
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		suite.hubChain.SenderAccount.GetAddress().String(),
		suite.rollappChain.ChainID,
		10,
		[]string{},
		[]rollapptypes.TokenMetadata{
			{
				Base: denom,
				DenomUnits: []*rollapptypes.DenomUnit{
					{
						Denom:    denom,
						Exponent: 0,
					},
					{
						Denom:    displayDenom,
						Exponent: 6,
					},
				},
				Description: "stake as rollapp token",
				Display:     displayDenom,
				Name:        displayDenom,
				Symbol:      strings.ToUpper(displayDenom),
			},
		},
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed
}

func (suite *KeeperTestSuite) FinalizeRollapp() error {
	rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
	ctx := suite.hubChain.GetContext()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: suite.rollappChain.ChainID, Index: 2}
	stateInfo := rollapptypes.StateInfo{
		StateInfoIndex: stateInfoIdx,
		StartHeight:    uint64(ctx.BlockHeader().Height),
		NumBlocks:      10,
		Status:         rollapptypes.STATE_STATUS_FINALIZED,
	}

	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
	// uppdate the LatestStateInfoIndex of the rollapp
	rollappKeeper.SetLatestFinalizedStateIndex(ctx, stateInfoIdx)

	err := rollappKeeper.GetHooks().AfterStateFinalized(
		suite.hubChain.GetContext(),
		suite.rollappChain.ChainID,
		&stateInfo,
	)
	return err
}

func (suite *KeeperTestSuite) NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}
