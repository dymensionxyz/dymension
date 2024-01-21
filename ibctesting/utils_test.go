package ibctesting_test

import (
	"encoding/json"
	"strings"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	app "github.com/dymensionxyz/dymension/app"
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
}

func ConvertToApp(chain *ibctesting.TestChain) *app.App {
	app, ok := chain.App.(*app.App)
	require.True(chain.T, ok)

	return app
}

// TODO: Change IBCTestUtilSuite to IBCUtilsTestSuite and wrap each test in a subsuite
// IBCTestUtilSuite is a testing suite to test keeper functions.
type IBCTestUtilSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	hubChain     *ibctesting.TestChain
	cosmosChain  *ibctesting.TestChain
	rollappChain *ibctesting.TestChain
}

// TestKeeperTestSuite runs all the tests within this package.
// func TestKeeperTestSuite(t *testing.T) {
// 	suite.Run(t, new(IBCTestUtilSuite))
// }

// SetupTest creates a coordinator with 2 test chains.
func (suite *IBCTestUtilSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)               // initializes 3 test chains
	suite.hubChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))     // convenience and readability
	suite.cosmosChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))  // convenience and readability
	suite.rollappChain = suite.coordinator.GetChain(ibctesting.GetChainID(3)) // convenience and readability
}

func (suite *IBCTestUtilSuite) CreateRollapp() {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		suite.hubChain.SenderAccount.GetAddress().String(),
		suite.rollappChain.ChainID,
		10,
		[]string{},
		nil,
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed

}

func (suite *IBCTestUtilSuite) CreateRollappWithMetadata(denom string) {
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
						Denom:    "big" + denom,
						Exponent: 6,
					},
				},
				Description: "stake as rollapp token",
				Display:     strings.ToUpper(denom),
				Name:        strings.ToUpper(denom),
				Symbol:      strings.ToUpper(denom),
			},
		},
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed
}

func (suite *IBCTestUtilSuite) UpdateRollappState(index uint64, startHeight uint64) {
	rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
	ctx := suite.hubChain.GetContext()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: suite.rollappChain.ChainID, Index: index}
	stateInfo := rollapptypes.StateInfo{
		StateInfoIndex: stateInfoIdx,
		StartHeight:    startHeight,
		NumBlocks:      1,
		Status:         rollapptypes.STATE_STATUS_RECEIVED,
	}

	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
}

func (suite *IBCTestUtilSuite) FinalizeRollappState(index uint64, endHeight uint64) error {
	rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
	ctx := suite.hubChain.GetContext()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: suite.rollappChain.ChainID, Index: index}
	stateInfo, found := rollappKeeper.GetStateInfo(ctx, suite.rollappChain.ChainID, stateInfoIdx.Index)
	suite.Require().True(found)
	stateInfo.NumBlocks = endHeight - stateInfo.StartHeight + 1
	stateInfo.Status = rollapptypes.STATE_STATUS_FINALIZED
	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
	// update the LatestStateInfoIndex of the rollapp
	rollappKeeper.SetLatestFinalizedStateIndex(ctx, stateInfoIdx)
	err := rollappKeeper.GetHooks().AfterStateFinalized(
		suite.hubChain.GetContext(),
		suite.rollappChain.ChainID,
		&stateInfo,
	)
	return err
}

func (suite *IBCTestUtilSuite) NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort

	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

func (suite *IBCTestUtilSuite) GetRollappToHubIBCDenomFromPacket(packet channeltypes.Packet) string {
	var data transfertypes.FungibleTokenPacketData
	err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data)
	suite.Require().NoError(err)
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + data.Denom
	// construct the denomination trace from the full raw denomination
	denomTrace := types.ParseDenomTrace(prefixedDenom)
	return denomTrace.IBCDenom()
}
