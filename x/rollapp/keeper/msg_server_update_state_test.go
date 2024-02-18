package keeper_test

import (
	"strconv"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	abci "github.com/tendermint/tendermint/abci/types"
)

//TODO: refactor the tests to use test-cases
//TODO: wrap the create rollapp and sequencer into a helper function

func (suite *RollappTestSuite) TestFirstUpdateState() {
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
		SequencerAddress: bob,
		RollappId:        rollapp.GetRollappId(),
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

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

func (suite *RollappTestSuite) TestUpdateState() {
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
		SequencerAddress: bob,
		RollappId:        rollapp.GetRollappId(),
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

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

func (suite *RollappTestSuite) TestUpdateStateUnknownRollappId() {
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
	suite.EqualError(err, types.ErrUnknownRollappID.Error())
}

// FIXME: need to add sequncer to rollapp to test this scenario
func (suite *RollappTestSuite) TestUpdateStateVersionMismatch() {
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

	// update state of version different than the rollapp
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

// FIXME: need to add sequncer to rollapp to test this scenario
func (suite *RollappTestSuite) TestUpdateStateUnknownSequencer() {
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

func (suite *RollappTestSuite) TestUpdateStateSequencerRollappMismatch() {
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
		SequencerAddress: bob,
		RollappId:        "rollapp2",
		Status:           sequencertypes.Proposer,
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

func (suite *RollappTestSuite) TestUpdateStateErrLogicUnpermissioned() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{carol},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set unpermissioned sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Proposer,
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

func (suite *RollappTestSuite) TestFirstUpdateStateErrWrongBlockHeightInitial() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

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

func (suite *RollappTestSuite) TestFirstUpdateStateErrWrongBlockHeight() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

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

func (suite *RollappTestSuite) TestUpdateStateErrWrongBlockHeight() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

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

func (suite *RollappTestSuite) TestUpdateStateErrLogicMissingStateInfo() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Proposer,
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

func (suite *RollappTestSuite) TestUpdateStateErrMultiUpdateStateInBlock() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)

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

// TODO: should test all status other than Proposer
func (suite *RollappTestSuite) TestUpdateStateErrNotActiveSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Bonded,
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
	suite.ErrorIs(err, sequencertypes.ErrNotActiveSequencer)
}

// ---------------------------------------
// verifyAll receives a list of expected results and a map of rollapId->rollapp
// the function verifies that the map contains all the rollapps that are in the list and only them
func verifyAll(suite *RollappTestSuite, rollappsExpect []*types.RollappSummary, rollappsRes map[string]*types.RollappSummary) {
	// check number of items are equal
	suite.Require().EqualValues(len(rollappsExpect), len(rollappsRes))
	for i := 0; i < len(rollappsExpect); i++ {
		rollappExpect := rollappsExpect[i]
		rollappRes := rollappsRes[rollappExpect.GetRollappId()]
		// println("rollappId:", rollappExpect.GetRollappId(), "=>", "rollapp:", rollappExpect.String())
		suite.Require().EqualValues(&rollappExpect, &rollappRes)
	}
}

// getAll queries for all existing rollapps and returns a tuple of:
// map of rollappId->rollapp and the number of retrieved rollapps
func getAll(suite *RollappTestSuite) (map[string]*types.RollappSummary, int) {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	totalChecked := 0
	totalRes := 0
	nextKey := []byte{}
	rollappsRes := make(map[string]*types.RollappSummary)
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
