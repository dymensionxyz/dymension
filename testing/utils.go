package irctesting

import (
	"encoding/json"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	"github.com/cosmos/ibc-go/v3/testing/simapp"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dymensionapp "github.com/dymensionxyz/dymension/app"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/x/sequencer/types"
)

var (
	timeoutHeight = clienttypes.NewHeight(1, 100)
)

type IrcTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	chainC *ibctesting.TestChain

	// consensus setup
	chainAConsensusType string
	chainBConsensusType string
	chainCConsensusType string

	// sequencers account
	rollappToSeqAccount map[string]ibctesting.SenderAccount
}

// SetupDymenstionTestingApp initializes a new SetupTestApp implements the interface for ibctesting.DefaultTestingAppInit
func SetupDymenstionTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
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
		return SetupDymenstionTestingApp()
	}
	// build dymint chain
	return ibctesting.SetupTestingApp(chainConsensusType)
}

func (suite *IrcTestSuite) SetupTest() {
	// setup the testing app creation callback
	ibctesting.DefaultTestingAppInit = SetupTestingApp
	// setup endpoints

	suite.coordinator = ibctesting.NewCoordinatorWithConsensusType(suite.T(), []string{suite.chainAConsensusType,
		suite.chainBConsensusType,
		suite.chainCConsensusType})

	// dymension hub chain
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	// fill missing bds
	bds := rollapptypes.BlockDescriptors{}
	for height := uint64(1); height <= uint64(suite.chainB.GetContext().BlockHeader().Height); {
		bd := rollapptypes.BlockDescriptor{
			Height:                 height,
			StateRoot:              hash32,
			IntermediateStatesRoot: hash32,
		}
		bds.BD = append(bds.BD, bd)
		height += 1
	}

	// rollapp1 dymint chain
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	// replace the TestChainClient to be DymintTestChainClient
	suite.chainA.TestChainClient = &DymintTestChainClient{
		baseTestChainClient: suite.chainA.TestChainClient,
		baseTestChain:       suite.chainA,
		bds:                 bds,
	}

	// rollapp2 dymint chain
	suite.chainC = suite.coordinator.GetChain(ibctesting.GetChainID(3))
	// replace the TestChainClient to be DymintTestChainClient
	suite.chainC.TestChainClient = &DymintTestChainClient{
		baseTestChainClient: suite.chainC.TestChainClient,
		baseTestChain:       suite.chainC,
		bds:                 bds,
	}

	suite.rollappToSeqAccount = make(map[string]ibctesting.SenderAccount)
	// allocate sequencer acounts on the hub
	suite.rollappToSeqAccount[suite.chainA.ChainID] = suite.chainB.SenderAccounts[1]
	suite.rollappToSeqAccount[suite.chainC.ChainID] = suite.chainB.SenderAccounts[2]

	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)
	suite.coordinator.CommitNBlocks(suite.chainC, 2)
}

// CreateRollapp creates a rollapp on the hub
func (suite *IrcTestSuite) CreateRollapp(rollapId string) (err error) {

	msg := rollapptypes.NewMsgCreateRollapp(
		suite.chainB.SenderAccount.GetAddress().String(),
		rollapId,
		"argCodeStamp",
		"argGenesisPath",
		7,
		1,
		&sharedtypes.Sequencers{
			Addresses: []string{},
		},
	)

	_, err = suite.chainB.SendMsgs(msg)

	return err
}

// CreateSequencer creates a sequencer on the hub
func (suite *IrcTestSuite) CreateSequencer(rollapId string) error {

	seqAccount := suite.rollappToSeqAccount[rollapId]
	msg, err := sequencertypes.NewMsgCreateSequencer(
		seqAccount.SenderAccount.GetAddress().String(),
		seqAccount.SenderAccount.GetPubKey(),
		rollapId,
		new(sequencertypes.Description),
	)

	if err != nil {
		return err
	}

	_, err = SendMsgs(suite.chainB, seqAccount, msg)

	return err
}

// UpdateState sends state update of all the the new blocks that were collected in the rollapp chain since the last update
func (suite *IrcTestSuite) UpdateState(rollapId string, bds *rollapptypes.BlockDescriptors) (err error) {

	seqAccount := suite.rollappToSeqAccount[rollapId]

	msg := rollapptypes.NewMsgUpdateState(
		seqAccount.SenderAccount.GetAddress().String(),
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

	_, err = SendMsgs(suite.chainB, seqAccount, msg)

	// reset bds array
	bds.BD = nil

	return err
}

// FinalizeState advance the hub chain in DisputePeriodInBlocks blocks
// All rollapp updates will become finalized
func (suite *IrcTestSuite) FinalizeState() (err error) {
	disputePeriodInBlocks := suite.chainB.App.(*dymensionapp.App).RollappKeeper.DisputePeriodInBlocks(suite.chainB.GetContext())
	suite.coordinator.CommitNBlocks(suite.chainB, disputePeriodInBlocks)
	return err
}

// UpdateRollappState sends state update of all the the new blocks that were collected in the rollapp chain since the last update
// and finalize the state
func (suite *IrcTestSuite) UpdateRollappState(rollapId string, bds *rollapptypes.BlockDescriptors) (err error) {
	// update rollapp state
	err = suite.UpdateState(rollapId, bds)
	suite.Require().Nil(err)
	err = suite.FinalizeState()
	return err
}

// GetHubSimApp returns the SimApp to allow usage ofnon-interface fields.
// CONTRACT: This function should not be called by third parties implementing
// their own SimApp.
func GetHubSimApp(chain *ibctesting.TestChain) *dymensionapp.App {
	app, ok := chain.App.(*dymensionapp.App)
	require.True(chain.T, ok)

	return app
}

// GetRollappSimApp returns the SimApp to allow usage ofnon-interface fields.
// CONTRACT: This function should not be called by third parties implementing
// their own SimApp.
func GetRollappSimApp(chain *ibctesting.TestChain) *simapp.SimApp {
	return chain.GetSimApp()
}

// SendMsgs delivers a transaction through the application. It updates the senders sequence
// number and updates the TestChain's headers. It returns the result and error if one
// occurred.
func SendMsgs(dymHubChain *ibctesting.TestChain, sequencerAccount ibctesting.SenderAccount, msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	dymHubChain.Coordinator.UpdateTimeForChain(dymHubChain)

	_, r, err := simapp.SignAndDeliver(
		dymHubChain.T,
		dymHubChain.TxConfig,
		dymHubChain.App.GetBaseApp(),
		dymHubChain.GetContext().BlockHeader(),
		msgs,
		dymHubChain.ChainID,
		[]uint64{sequencerAccount.SenderAccount.GetAccountNumber()},
		[]uint64{sequencerAccount.SenderAccount.GetSequence()},
		true, true, sequencerAccount.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	dymHubChain.NextBlock()

	// increment sequence for successful transaction execution
	err = sequencerAccount.SenderAccount.SetSequence(sequencerAccount.SenderAccount.GetSequence() + 1)

	dymHubChain.Coordinator.IncrementTime()

	return r, err
}
