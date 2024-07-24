package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	v2types "github.com/dymensionxyz/dymension/v3/x/rollapp/types/v2"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func (suite *RollappTestSuite) TestFirstUpdateStateV2() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        rollapp.GetRollappId(),
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// check no index exists
	_, found := suite.App.RollappKeeper.GetLatestStateInfoIndex(suite.Ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(false, found)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	// check first index is 1
	expectedLatestStateInfoIndex, found := suite.App.RollappKeeper.GetLatestStateInfoIndex(suite.Ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(true, found)
	suite.Require().EqualValues(expectedLatestStateInfoIndex.Index, 1)
}

func (suite *RollappTestSuite) TestUpdateStateV2() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// parameters
	disputePeriodInBlocks := suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        rollapp.GetRollappId(),
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// create new update
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	// update state
	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	// test 10 update state
	for i := 0; i < 10; i++ {
		// bump block height

		if i == 3 {
			disputePeriodInBlocks += 2
		}

		if i == 6 {
			disputePeriodInBlocks -= 3
		}

		suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
		goCtx = sdk.WrapSDKContext(suite.Ctx)

		// calc new updateState
		latestStateInfoIndex, found := suite.App.RollappKeeper.GetLatestStateInfoIndex(suite.Ctx, rollapp.GetRollappId())
		suite.Require().EqualValues(true, found)
		// verify index
		suite.Require().EqualValues(i+1, latestStateInfoIndex.Index)
		// load last state info
		expectedStateInfo, found := suite.App.RollappKeeper.GetStateInfo(suite.Ctx, rollapp.GetRollappId(), latestStateInfoIndex.GetIndex())
		suite.Require().EqualValues(true, found)

		// verify finalization queue
		expectedFinalizationQueue, _ := suite.App.RollappKeeper.GetBlockHeightToFinalizationQueue(suite.Ctx, expectedStateInfo.CreationHeight)
		suite.Require().EqualValues(expectedFinalizationQueue, types.BlockHeightToFinalizationQueue{
			CreationHeight:    expectedStateInfo.CreationHeight,
			FinalizationQueue: []types.StateInfoIndex{latestStateInfoIndex},
		}, "finalization queue", "i", i)

		// create new update
		updateState := v2types.MsgUpdateState{
			Creator:     bob,
			RollappId:   rollapp.GetRollappId(),
			StartHeight: expectedStateInfo.StartHeight + expectedStateInfo.NumBlocks,
			NumBlocks:   2,
			DAPath:      &types.DAPath{DaType: "interchain"},
			Version:     3,
			BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: expectedStateInfo.StartHeight}, {Height: expectedStateInfo.StartHeight + 1}}},
		}

		// update state
		_, err = suite.msgServerV2.UpdateState(goCtx, &updateState)
		suite.Require().Nil(err)

		// end block
		suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})

		if uint64(suite.Ctx.BlockHeight()) > disputePeriodInBlocks {
			for i := uint64(1); i <= latestStateInfoIndex.Index; i++ {
				expectedStateInfo, _ := suite.App.RollappKeeper.GetStateInfo(suite.Ctx, rollapp.GetRollappId(), i)
				if expectedStateInfo.CreationHeight < uint64(suite.Ctx.BlockHeight())-disputePeriodInBlocks {
					suite.Require().EqualValues(expectedStateInfo.Status, common.Status_FINALIZED)
				}
			}
		}

		// check finalization status change
		pendingQueues := suite.App.RollappKeeper.GetAllFinalizationQueueUntilHeightInclusive(suite.Ctx, uint64(suite.Ctx.BlockHeader().Height))

		for _, finalizationQueue := range pendingQueues {

			// fmt.Printf("finalizationQueue: %s %d\n", finalizationQueue.String())
			stateInfo, found := suite.App.RollappKeeper.GetStateInfo(suite.Ctx, finalizationQueue.FinalizationQueue[0].RollappId, finalizationQueue.FinalizationQueue[0].Index)
			suite.Require().True(found)
			// fmt.Printf("stateInfo: %s\n", stateInfo.String())

			suite.Require().EqualValues(stateInfo.Status, common.Status_PENDING)

		}
	}
}

func (suite *RollappTestSuite) TestUpdateStateV2_UnknownRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// update state of unknown rollapp
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   "rollapp1",
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     0,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.EqualError(err, types.ErrUnknownRollappID.Error())
}

// FIXME: need to add sequncer to rollapp to test this scenario
func (suite *RollappTestSuite) TestUpdateStateV2_VersionMismatch() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// update state of version different than the rollapp
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     0,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrVersionMismatch)
}

// FIXME: need to add sequncer to rollapp to test this scenario
func (suite *RollappTestSuite) TestUpdateStateV2_UnknownSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrUnknownSequencer)
}

func (suite *RollappTestSuite) TestUpdateStateV2_SequencerRollappMismatch() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:     "rollapp1",
		Creator:       alice,
		Version:       3,
		MaxSequencers: 1,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp2",
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrSequencerRollappMismatch)
}

func (suite *RollappTestSuite) TestUpdateStateV2_ErrLogicUnpermissioned() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{carol},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set unpermissioned sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrLogic)
}

func (suite *RollappTestSuite) TestFirstUpdateStateV2_GensisHightGreaterThanZero() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 2,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.NoError(err)
}

func (suite *RollappTestSuite) TestUpdateStateV2_ErrWrongBlockHeight() {
	suite.SetupTest()
	_ = sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

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
		Status:         common.Status_PENDING,
		BDs:            types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}
	suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, latestStateInfoIndex)
	suite.App.RollappKeeper.SetStateInfo(suite.Ctx, stateInfo)

	// bump block height
	suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 2,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}, {Height: 4}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrWrongBlockHeight)
}

func (suite *RollappTestSuite) TestUpdateStateV2_ErrLogicMissingStateInfo() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Bonded,
		Proposer:         true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// set initial latestStateInfoIndex
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: "rollapp1",
		Index:     1,
	}
	suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, latestStateInfoIndex)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, types.ErrLogic)
}

// TODO: should test all status other than Proposer
func (suite *RollappTestSuite) TestUpdateStateV2_ErrNotActiveSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               3,
		MaxSequencers:         1,
		PermissionedAddresses: []string{},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		SequencerAddress: bob,
		RollappId:        "rollapp1",
		Status:           sequencertypes.Bonded,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := v2types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      &types.DAPath{DaType: "interchain"},
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServerV2.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrNotActiveSequencer)
}