package keeper_test

import (
	"fmt"
	"strconv"

	"testing"

	dymensionapp "github.com/dymensionxyz/dymension/app"
	sharedtypes "github.com/dymensionxyz/dymension/shared/types"
	"github.com/dymensionxyz/dymension/testutil/sample"
	"github.com/dymensionxyz/dymension/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/x/sequencer/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	transferEventCount            = 3 // As emitted by the bank
	createEventCount              = 8
	playEventCountFirst           = 8 // Extra "sender" attribute emitted by the bank
	playEventCountNext            = 7
	rejectEventCount              = 4
	rejectEventCountWithTransfer  = 5 // Extra "sender" attribute emitted by the bank
	forfeitEventCount             = 4
	forfeitEventCountWithTransfer = 5 // Extra "sender" attribute emitted by the bank
	alice                         = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bob                           = "cosmos1xyxs3skf3f4jfqeuv89yyaqvjc6lffavxqhc8g"
	carol                         = "cosmos1e0w5t53nrq7p66fye6c8p0ynyhf6y24l4yuxd7"
	balAlice                      = 50000000
	balBob                        = 20000000
	balCarol                      = 10000000
	foreignToken                  = "foreignToken"
	balTokenAlice                 = 5
	balTokenBob                   = 2
	balTokenCarol                 = 1
)

var (
	rollappModuleAddress string
)

type IntegrationTestSuite struct {
	suite.Suite

	app         *dymensionapp.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupTest(deployerWhitelist ...string) {
	app := dymensionapp.Setup(false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.RollappKeeper.SetParams(ctx, types.NewParams(2, deployerWhitelist))
	rollappModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.RollappKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *IntegrationTestSuite) createRollappFromWhitelist(expectedErr error, deployerWhitelist []string) {
	suite.SetupTest(deployerWhitelist...)
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// rollappsExpect is the expected result of query all
	rollappsExpect := []*types.Rollapp{}

	// test 10 rollap creations
	for i := 0; i < 10; i++ {
		// generate sequences address
		addresses := sample.GenerateAddresses(i)
		// rollapp is the rollapp to create
		rollapp := types.MsgCreateRollapp{
			Creator:               alice,
			RollappId:             fmt.Sprintf("%s%d", "rollapp", i),
			CodeStamp:             "",
			GenesisPath:           "",
			MaxWithholdingBlocks:  1,
			MaxSequencers:         1,
			PermissionedAddresses: sharedtypes.Sequencers{Addresses: addresses},
		}
		// rollappExpect is the expected result of creating rollapp
		rollappExpect := types.Rollapp{
			RollappId:             rollapp.GetRollappId(),
			Creator:               rollapp.GetCreator(),
			Version:               0,
			CodeStamp:             rollapp.GetCodeStamp(),
			GenesisPath:           rollapp.GetGenesisPath(),
			MaxWithholdingBlocks:  rollapp.GetMaxWithholdingBlocks(),
			MaxSequencers:         rollapp.GetMaxSequencers(),
			PermissionedAddresses: rollapp.GetPermissionedAddresses(),
		}
		// create rollapp
		createResponse, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
		if expectedErr != nil {
			suite.EqualError(err, expectedErr.Error())
			continue
		}
		suite.Require().Nil(err)
		suite.Require().EqualValues(types.MsgCreateRollappResponse{}, *createResponse)

		// query the specific rollapp
		queryResponse, err := suite.queryClient.Rollapp(goCtx, &types.QueryGetRollappRequest{
			RollappId: rollapp.GetRollappId(),
		})
		if queryResponse.Rollapp.PermissionedAddresses.Addresses == nil {
			queryResponse.Rollapp.PermissionedAddresses.Addresses = []string{}
		}
		suite.Require().Nil(err)
		suite.Require().EqualValues(&rollappExpect, &queryResponse.Rollapp)

		// add the rollapp to the list of get all expected list
		rollappsExpect = append(rollappsExpect, &rollappExpect)
		// verify that query all contains all the rollapps that were created
		rollappsRes, totalRes := getAll(suite)
		suite.Require().EqualValues(totalRes, i+1)
		vereifyAll(suite, rollappsExpect, rollappsRes)

	}

}

func (suite *IntegrationTestSuite) TestCreateRollapp() {
	suite.createRollappFromWhitelist(nil, nil)
}

func (suite *IntegrationTestSuite) TestCreateRollappFromWhitelist() {
	suite.createRollappFromWhitelist(nil, []string{alice})
}

func (suite *IntegrationTestSuite) TestCreateRollappUnauthorizedRollappCreator() {
	suite.createRollappFromWhitelist(types.ErrUnauthorizedRollappCreator, []string{bob})
}

func (suite *IntegrationTestSuite) TestCreateRollappAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// rollapp is the rollapp to create
	rollapp := types.MsgCreateRollapp{
		Creator:               alice,
		RollappId:             "rollapp1",
		CodeStamp:             "",
		GenesisPath:           "",
		MaxWithholdingBlocks:  1,
		MaxSequencers:         1,
		PermissionedAddresses: sharedtypes.Sequencers{},
	}
	_, err := suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateRollapp(goCtx, &rollapp)
	suite.EqualError(err, types.ErrRollappExists.Error())
}

func (suite *IntegrationTestSuite) TestFirstUpdateState() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        rollapp.GetRollappId(),
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// check no index exists
	expectedLatestStateInfoIndex, found := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(false, found)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	// check first index is 1
	expectedLatestStateInfoIndex, found = suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(true, found)
	suite.Require().EqualValues(expectedLatestStateInfoIndex.Index, 1)
}

func (suite *IntegrationTestSuite) TestUpdateState() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// parameters
	disputePeriodInBlocks := suite.app.RollappKeeper.DisputePeriodInBlocks(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        rollapp.GetRollappId(),
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// create new update
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	// update state
	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	// test 10 update state
	for i := 0; i < 10; i++ {
		// bump block height
		suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeader().Height + 1)
		goCtx = sdk.WrapSDKContext(suite.ctx)

		// calc new updateState
		expectedLatestStateInfoIndex, found := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
		suite.Require().EqualValues(true, found)
		// verify index
		suite.Require().EqualValues(i+1, expectedLatestStateInfoIndex.Index)
		// load last state info
		expectedStateInfo, found := suite.app.RollappKeeper.GetStateInfo(suite.ctx, rollapp.GetRollappId(), expectedLatestStateInfoIndex.GetIndex())
		suite.Require().EqualValues(true, found)

		// verify finalization queue
		expectedFinalization := expectedStateInfo.CreationHeight + disputePeriodInBlocks
		expectedFinalizationQueue, found := suite.app.RollappKeeper.GetBlockHeightToFinalizationQueue(suite.ctx, expectedFinalization)
		suite.Require().EqualValues(expectedFinalizationQueue, types.BlockHeightToFinalizationQueue{
			FinalizationHeight: expectedFinalization,
			FinalizationQueue:  []types.StateInfoIndex{expectedLatestStateInfoIndex},
		})

		// create new update
		updateState := types.MsgUpdateState{
			Creator:     bob,
			RollappId:   rollapp.GetRollappId(),
			StartHeight: expectedStateInfo.StartHeight + expectedStateInfo.NumBlocks,
			NumBlocks:   2,
			DAPath:      "",
			Version:     3,
			BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: expectedStateInfo.StartHeight}, {Height: expectedStateInfo.StartHeight + 1}}},
		}

		// update state
		_, err := suite.msgServer.UpdateState(goCtx, &updateState)
		suite.Require().Nil(err)

		// end block
		responseEndBlock := suite.app.EndBlocker(suite.ctx, abci.RequestEndBlock{Height: suite.ctx.BlockHeight()})

		// check finalization status change
		finalizationQueue, found := suite.app.RollappKeeper.GetBlockHeightToFinalizationQueue(suite.ctx, uint64(suite.ctx.BlockHeader().Height))
		if found {
			//fmt.Printf("finalizationQueue: %s\n", finalizationQueue.String())
			stateInfo, found := suite.app.RollappKeeper.GetStateInfo(suite.ctx, finalizationQueue.FinalizationQueue[0].RollappId, finalizationQueue.FinalizationQueue[0].Index)
			suite.Require().True(found)
			//fmt.Printf("stateInfo: %s\n", stateInfo.String())
			suite.Require().EqualValues(stateInfo.CreationHeight, uint64(suite.ctx.BlockHeader().Height)-disputePeriodInBlocks)
			suite.Require().EqualValues(stateInfo.Status, types.STATE_STATUS_FINALIZED)
			// use a boolean to ensure the event exists
			contains := false
			for _, event := range responseEndBlock.Events {
				if event.Type == types.EventTypeStatusChange {
					contains = true
					// there are 5 attributes in the event
					suite.Require().EqualValues(5, len(event.Attributes))
					for _, attr := range event.Attributes {
						switch string(attr.Key) {
						case types.AttributeKeyRollappId:
							suite.Require().EqualValues(string(attr.Value), rollapp.RollappId)
						case types.AttributeKeyStateInfoIndex:
							suite.Require().EqualValues(string(attr.Value), strconv.FormatUint(stateInfo.StateInfoIndex.Index, 10))
						case types.AttributeKeyStartHeight:
							suite.Require().EqualValues(string(attr.Value), strconv.FormatUint(stateInfo.StartHeight, 10))
						case types.AttributeKeyNumBlocks:
							suite.Require().EqualValues(string(attr.Value), strconv.FormatUint(stateInfo.NumBlocks, 10))
						case types.AttributeKeyStatus:
							suite.Require().EqualValues(string(attr.Value), stateInfo.Status.String())
						default:
							suite.Fail("unexpected attribute in event: %s", event.String())
						}
					}
				}

			}
			suite.Require().True(contains)
		} else {
			suite.Require().LessOrEqualf(uint64(suite.ctx.BlockHeader().Height), disputePeriodInBlocks,
				"no finalization for currHeight(%d), disputePeriodInBlocks(%d)", suite.ctx.BlockHeader().Height, disputePeriodInBlocks)
		}
	}
}

func (suite *IntegrationTestSuite) TestUpdateStateUnknownRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// update state of unknown rollapp
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   "rollapp1",
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     0,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.EqualError(err, types.ErrUnknownRollappId.Error())
}

func (suite *IntegrationTestSuite) TestUpdateStateVersionMismatch() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// update state of version diffrent than the rollapp
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     0,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrVersionMismatch)
}

func (suite *IntegrationTestSuite) TestUpdateStateUnknownSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrUnknownSequencer)
}

func (suite *IntegrationTestSuite) TestUpdateStateSequencerRollappMismatch() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp2",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrSequencerRollappMismatch)
}

func (suite *IntegrationTestSuite) TestUpdateStateErrLogicUnpermissioned() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{carol},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set unpermissioned sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sdkerrors.ErrLogic)
}

func (suite *IntegrationTestSuite) TestFirstUpdateStateErrWrongBlockHeightInitial() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 0,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 0}, {Height: 1}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrWrongBlockHeight)
}

func (suite *IntegrationTestSuite) TestFirstUpdateStateErrWrongBlockHeight() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 2,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrWrongBlockHeight)
}

func (suite *IntegrationTestSuite) TestUpdateStateErrWrongBlockHeight() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// set initial latestStateInfoIndex & StateInfo
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: "rollapp1",
		Index:     1,
	}
	stateInfo := types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: "rollapp1", Index: 1},
		Sequencer:      sequencer.SequencerAddress,
		StartHeight:    1,
		NumBlocks:      3,
		DAPath:         "",
		Version:        0,
		CreationHeight: 0,
		Status:         types.STATE_STATUS_RECEIVED,
		BDs:            types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}
	suite.app.RollappKeeper.SetLatestStateInfoIndex(suite.ctx, latestStateInfoIndex)
	suite.app.RollappKeeper.SetStateInfo(suite.ctx, stateInfo)

	// bump block height
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeader().Height + 1)
	goCtx = sdk.WrapSDKContext(suite.ctx)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 2,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}, {Height: 4}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrWrongBlockHeight)
}

func (suite *IntegrationTestSuite) TestUpdateStateErrLogicMissingStateInfo() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

	// set initial latestStateInfoIndex
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: "rollapp1",
		Index:     1,
	}
	suite.app.RollappKeeper.SetLatestStateInfoIndex(suite.ctx, latestStateInfoIndex)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sdkerrors.ErrLogic)
}

func (suite *IntegrationTestSuite) TestUpdateStateErrMultiUpdateStateInBlock() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// set initial latestStateInfoIndex & StateInfo
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: "rollapp1",
		Index:     1,
	}
	stateInfo := types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: "rollapp1", Index: 1},
		Sequencer:      sequencer.SequencerAddress,
		StartHeight:    1,
		NumBlocks:      3,
		DAPath:         "",
		Version:        0,
		CreationHeight: 0,
		Status:         types.STATE_STATUS_RECEIVED,
		BDs:            types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}
	suite.app.RollappKeeper.SetLatestStateInfoIndex(suite.ctx, latestStateInfoIndex)
	suite.app.RollappKeeper.SetStateInfo(suite.ctx, stateInfo)

	// we don't bump block height

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 3,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 3}, {Height: 4}, {Height: 5}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrMultiUpdateStateInBlock)
}

func (suite *IntegrationTestSuite) TestUpdateStateErrLogicNotRegisteredInScheduler() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)
	// skip register sequncer in sequencer as Proposer

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sdkerrors.ErrLogic)
}

func (suite *IntegrationTestSuite) TestUpdateStateErrNotActiveSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
		PermissionedAddresses: sharedtypes.Sequencers{
			Addresses: []string{},
		},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Creator:          alice,
		SequencerAddress: bob,
		RollappId:        "rollapp1",
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

	// register sequncer in sequencer as Inactive
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Inactive,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrNotActiveSequencer)
}

//---------------------------------------
// vereifyAll receives a list of expected results and a map of rollapId->rollapp
// the function verifies that the map contains all the rollapps that are in the list and only them
func vereifyAll(suite *IntegrationTestSuite, rollappsExpect []*types.Rollapp, rollappsRes map[string]*types.Rollapp) {
	// check number of items are equal
	suite.Require().EqualValues(len(rollappsExpect), len(rollappsRes))
	for i := 0; i < len(rollappsExpect); i++ {
		rollappExpect := rollappsExpect[i]
		rollappRes := rollappsRes[rollappExpect.GetRollappId()]
		// println("rollappId:", rollappExpect.GetRollappId(), "=>", "rollapp:", rollappExpect.String())
		suite.Require().EqualValues(&rollappExpect, &rollappRes)
	}
}

// getAll queries for all exsisting rollapps and returns a tuple of:
// map of rollappId->rollapp and the number of retrieved rollapps
func getAll(suite *IntegrationTestSuite) (map[string]*types.Rollapp, int) {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	totalChecked := 0
	totalRes := 0
	nextKey := []byte{}
	rollappsRes := make(map[string]*types.Rollapp)
	for {
		queryAllResponse, err := suite.queryClient.RollappAll(goCtx,
			&types.QueryAllRollappRequest{
				Pagination: &query.PageRequest{
					Key:        nextKey,
					Offset:     0,
					Limit:      0,
					CountTotal: true,
					Reverse:    false,
				}})
		suite.Require().Nil(err)

		if totalRes == 0 {
			totalRes = int(queryAllResponse.GetPagination().GetTotal())
		}

		for i := 0; i < len(queryAllResponse.GetRollapp()); i++ {
			rollappRes := queryAllResponse.GetRollapp()[i]
			if rollappRes.PermissionedAddresses.Addresses == nil {
				rollappRes.PermissionedAddresses.Addresses = []string{}
			}
			rollappsRes[rollappRes.GetRollappId()] = &rollappRes
		}
		totalChecked += len(queryAllResponse.GetRollapp())
		nextKey = queryAllResponse.GetPagination().GetNextKey()

		if nextKey == nil {
			break
		}
	}

	return rollappsRes, totalRes
}
