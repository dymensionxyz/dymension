package irctesting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"

	dymensionapp "github.com/dymensionxyz/dymension/app"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/x/sequencer/types"
)

type IrcTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	// consensus setup
	chainAConsensusType string
	chainBConsensusType string

	// track rollapp BDs for update
	bds rollapptypes.BlockDescriptors
}

// SetupDymenstionTestingApp initializes a new SetupTestApp implements the interface for ibctesting.DefaultTestingAppInit
func SetupDymenstionTestingApp(chainConsensusType string) (ibctesting.TestingApp, map[string]json.RawMessage) {
	checkTx := false
	encCdc := simappparams.MakeTestEncodingConfig()
	app := dymensionapp.Setup(checkTx)
	return app, dymensionapp.NewDefaultGenesisState(encCdc.Marshaler)
}

// SetupTestingApp is a callback for the ibctesting to create a chain
// the test creates a dymension chain and dymint chain
func SetupTestingApp(chainConsensusType string) (ibctesting.TestingApp, map[string]json.RawMessage) {
	// build dymension chain
	if chainConsensusType == exported.Tendermint {
		return SetupDymenstionTestingApp(chainConsensusType)
	}
	// build dymint chain
	return ibctesting.SetupTestingApp(chainConsensusType)
}

func (suite *IrcTestSuite) SetupTest() {
	// setup the testing app creation callback
	ibctesting.DefaultTestingAppInit = SetupTestingApp
	// setup endpoints
	suite.coordinator = ibctesting.NewCoordinatorWithConsensusType(suite.T(), []string{suite.chainAConsensusType, suite.chainBConsensusType})
	// rollapp dymint chain
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	// replace the TestChainClient to be DymintTestChainClient
	suite.chainA.TestChainClient = &DymintTestChainClient{suite.chainA.TestChainClient, suite.chainA, &suite.bds}
	// dymension hub chain
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	// fill missing bds
	for height := uint64(1); height <= uint64(suite.chainB.GetContext().BlockHeader().Height); {
		bd := rollapptypes.BlockDescriptor{
			Height:                 height,
			StateRoot:              hash32,
			IntermediateStatesRoot: hash32,
		}
		suite.bds.BD = append(suite.bds.BD, bd)
		height += 1
	}

	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)
}

// TestSuite initialize dymint rollapp and dymension hub chain
func TestSuite(t *testing.T) {
	suite.Run(t, &IrcTestSuite{
		chainAConsensusType: exported.Dymint,     // rollapp dymint chain
		chainBConsensusType: exported.Tendermint, // dymension hub chain
	})
}

// TestConection create connection between dymint rollapp and dymension hub chain
func (suite *IrcTestSuite) TestConection() {
	// create chains, chainA is dymint and chainB is dymension hub
	path := ibctesting.NewPath(suite.chainA, suite.chainB)
	// dymension hub chain
	dymensionEndpoint := path.EndpointB
	// rollapp dymint chain
	rollappEndpoint := path.EndpointA
	// create rollapp
	err := CreateRollapp(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID)
	suite.Require().Nil(err)
	// create sequencer
	err = CreateSequencer(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID)
	suite.Require().Nil(err)

	//--
	// Create Clients
	//--

	// update rollapp state
	err = UpdateState(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID, &suite.bds)
	suite.Require().Nil(err)
	FinalizeState(suite.coordinator, dymensionEndpoint.Chain)
	// create the client of dymension on the rollapp chain
	err = rollappEndpoint.CreateClient()
	suite.Require().Nil(err)
	// update rollapp state
	err = UpdateState(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID, &suite.bds)
	suite.Require().Nil(err)
	FinalizeState(suite.coordinator, dymensionEndpoint.Chain)
	// create the client of the rollapp on dymension
	err = dymensionEndpoint.CreateClient()
	suite.Require().Nil(err)

	//--
	// Create Connection
	//--

	// init a connection on the rollapp
	err = rollappEndpoint.ConnOpenInit()
	suite.Require().Nil(err)
	// update rollapp state
	err = UpdateState(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID, &suite.bds)
	suite.Require().Nil(err)
	FinalizeState(suite.coordinator, dymensionEndpoint.Chain)
	// try to open connection on the hub
	err = dymensionEndpoint.ConnOpenTry()
	suite.Require().Nil(err)
	// send ack to to rollapp
	err = rollappEndpoint.ConnOpenAck()
	suite.Require().Nil(err)
	// update rollapp state
	err = UpdateState(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID, &suite.bds)
	suite.Require().Nil(err)
	FinalizeState(suite.coordinator, dymensionEndpoint.Chain)
	// send confirmation to the hub
	err = dymensionEndpoint.ConnOpenConfirm()
	suite.Require().Nil(err)
	// ensure rollapp is up to date
	err = rollappEndpoint.UpdateClient()
	suite.Require().Nil(err)

	//--
	// Create Channel
	//--

	// set configuration for transfer channel
	rollappEndpoint.ChannelConfig.PortID = ibctransfertypes.PortID
	rollappEndpoint.ChannelConfig.Version = ibctransfertypes.Version
	dymensionEndpoint.ChannelConfig.PortID = ibctransfertypes.PortID
	dymensionEndpoint.ChannelConfig.Version = ibctransfertypes.Version
	// init a channel on the rollapp
	err = rollappEndpoint.ChanOpenInit()
	suite.Require().Nil(err)
	// update rollapp state
	err = UpdateState(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID, &suite.bds)
	suite.Require().Nil(err)
	FinalizeState(suite.coordinator, dymensionEndpoint.Chain)
	// try to open channel on the hub
	err = dymensionEndpoint.ChanOpenTry()
	suite.Require().Nil(err)
	// send ack to to rollapp
	err = rollappEndpoint.ChanOpenAck()
	suite.Require().Nil(err)
	// update rollapp state
	err = UpdateState(dymensionEndpoint.Chain, rollappEndpoint.Chain.ChainID, &suite.bds)
	suite.Require().Nil(err)
	FinalizeState(suite.coordinator, dymensionEndpoint.Chain)
	// send confirmation to the hub
	err = dymensionEndpoint.ChanOpenConfirm()
	suite.Require().Nil(err)
	// ensure counterparty is up to date
	err = rollappEndpoint.UpdateClient()
	suite.Require().Nil(err)
}

// CreateRollapp creates a rollapp on the hub
func CreateRollapp(dymHubChain *ibctesting.TestChain, rollapId string) (err error) {

	msg := rollapptypes.NewMsgCreateRollapp(
		dymHubChain.SenderAccount.GetAddress().String(),
		rollapId,
		"argCodeStamp",
		"argGenesisPath",
		7,
		1,
		&sharedtypes.Sequencers{
			Addresses: []string{},
		},
	)

	_, err = dymHubChain.SendMsgs(msg)

	return err
}

// CreateSequencer creates a sequencer on the hub
func CreateSequencer(dymHubChain *ibctesting.TestChain, rollapId string) (err error) {

	msg, err := sequencertypes.NewMsgCreateSequencer(
		dymHubChain.SenderAccount.GetAddress().String(),
		dymHubChain.SenderAccount.GetAddress().String(),
		dymHubChain.SenderAccount.GetPubKey(),
		rollapId,
		new(sequencertypes.Description),
	)

	if err != nil {
		return err
	}

	_, err = dymHubChain.SendMsgs(msg)

	return err
}

// UpdateState sends state update of all the the new blocks that were collected in the rollapp chain since the last update
func UpdateState(dymHubChain *ibctesting.TestChain, rollapId string, bds *rollapptypes.BlockDescriptors) (err error) {

	msg := rollapptypes.NewMsgUpdateState(
		dymHubChain.SenderAccount.GetAddress().String(),
		rollapId,
		bds.BD[0].Height,
		uint64(len(bds.BD)),
		"argDAPath",
		0,
		bds,
	)

	if err != nil {
		return err
	}

	_, err = dymHubChain.SendMsgs(msg)

	// reset bds array
	bds.BD = nil

	return err
}

// FinalizeState advance the hub chain in DisputePeriodInBlocks blocks
// All rollapp updates will become finalized
func FinalizeState(coord *ibctesting.Coordinator, dymHubChain *ibctesting.TestChain) (err error) {
	disputePeriodInBlocks := dymHubChain.App.(*dymensionapp.App).RollappKeeper.DisputePeriodInBlocks(dymHubChain.GetContext())
	coord.CommitNBlocks(dymHubChain, disputePeriodInBlocks)
	return err
}
