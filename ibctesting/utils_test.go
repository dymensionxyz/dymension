package ibctesting_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ChainIDPrefix defines the default chain ID prefix for Evmos test chains
var ChainIDPrefix = "evmos_9000-"

func init() {
	ibctesting.ChainIDPrefix = ChainIDPrefix
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		return apptesting.SetupTestingApp()
	}
}

func ConvertToApp(chain *ibctesting.TestChain) *app.App {
	app, ok := chain.App.(*app.App)
	require.True(chain.T, ok)

	return app
}

// IBCTestUtilSuite is a testing suite to test keeper functions.
type IBCTestUtilSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	hubChain     *ibctesting.TestChain
	cosmosChain  *ibctesting.TestChain
	rollappChain *ibctesting.TestChain
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *IBCTestUtilSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)              // initializes 3 test chains
	suite.hubChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))    // convenience and readability
	suite.cosmosChain = suite.coordinator.GetChain(ibctesting.GetChainID(2)) // convenience and readability

	var validators []*tmtypes.Validator
	var signersByAddress = make(map[string]tmtypes.PrivValidator, 1)
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()
	validators = append(validators, tmtypes.NewValidator(pubKey, 1))
	signersByAddress[pubKey.Address().String()] = privVal
	valSet := tmtypes.NewValidatorSet(validators)

	suite.rollappChain = ibctesting.NewTestChainWithValSet(suite.T(), suite.coordinator, ChainIDPrefix+"3", valSet, signersByAddress)
	//suite.rollappChain = NewTestChainWithValSet()
	fmt.Println(len(suite.rollappChain.Vals.Validators))
}

func (suite *IBCTestUtilSuite) CreateRollapp() {
	msgCreateRollapp := rollapptypes.NewMsgCreateRollapp(
		suite.hubChain.SenderAccount.GetAddress().String(),
		suite.rollappChain.ChainID,
		10,
		[]string{},
		nil,
		nil,
	)
	_, err := suite.hubChain.SendMsgs(msgCreateRollapp)
	suite.Require().NoError(err) // message committed

}

func (suite *IBCTestUtilSuite) RegisterSequencer() {

	bond := sequencertypes.DefaultParams().MinBond
	//fund account
	err := bankutil.FundAccount(ConvertToApp(suite.hubChain).BankKeeper, suite.hubChain.GetContext(), suite.hubChain.SenderAccount.GetAddress(), sdk.NewCoins(bond))
	suite.Require().Nil(err)

	msgCreateSequencer, err := sequencertypes.NewMsgCreateSequencer(
		suite.hubChain.SenderAccount.GetAddress().String(),
		suite.rollappChain.SenderAccount.GetPubKey(),
		suite.rollappChain.ChainID,
		&sequencertypes.Description{},
		bond,
	)

	suite.Require().NoError(err) // message committed
	_, err = suite.hubChain.SendMsgs(msgCreateSequencer)
	suite.Require().NoError(err) // message committed
}

func (suite *IBCTestUtilSuite) CreateRollappWithMetadata(denom string) {
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
		nil,
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
		Status:         common.Status_PENDING,
		Sequencer:      suite.hubChain.SenderAccount.GetAddress().String(),
	}

	// update the status of the stateInfo
	rollappKeeper.SetStateInfo(ctx, stateInfo)
	rollappKeeper.SetLatestStateInfoIndex(ctx, stateInfo.StateInfoIndex)
}

func (suite *IBCTestUtilSuite) FinalizeRollappState(index uint64, endHeight uint64) error {
	rollappKeeper := ConvertToApp(suite.hubChain).RollappKeeper
	ctx := suite.hubChain.GetContext()

	stateInfoIdx := rollapptypes.StateInfoIndex{RollappId: suite.rollappChain.ChainID, Index: index}
	stateInfo, found := rollappKeeper.GetStateInfo(ctx, suite.rollappChain.ChainID, stateInfoIdx.Index)
	suite.Require().True(found)
	stateInfo.NumBlocks = endHeight - stateInfo.StartHeight + 1
	stateInfo.Status = common.Status_FINALIZED
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
	return suite.GetIBCDenomForChannel(packet.GetDestChannel(), data.Denom)
}

func (suite *IBCTestUtilSuite) GetIBCDenomForChannel(channel string, denom string) string {
	// since SendPacket did not prefix the denomination, we must prefix denomination here
	sourcePrefix := types.GetDenomPrefix("transfer", channel)
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom
	// construct the denomination trace from the full raw denomination
	denomTrace := types.ParseDenomTrace(prefixedDenom)
	return denomTrace.IBCDenom()
}
