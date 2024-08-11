package keeper_test

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/sdk-utils/utils/urand"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *RollappTestSuite) TestFirstUpdateState() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	/*
			rollappId := suite.CreateDefaultRollapp()
		proposer := suite.CreateDefaultSequencer(suite.Ctx, rollappId)
	*/

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:        urand.RollappID(),
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		GenesisChecksum:  "checksum",
		Bech32Prefix:     "rol",
		Metadata: &types.RollappMetadata{
			Website:          "https://dymension.xyz",
			Description:      "Sample description",
			LogoDataUri:      "data:image/png;base64,c2lzZQ==",
			TokenLogoDataUri: "data:image/png;base64,ZHVwZQ==",
			Telegram:         "https://t.me/rolly",
			X:                "https://x.dymension.xyz",
		},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: rollapp.GetRollappId(),
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// check no index exists
	_, found := suite.App.RollappKeeper.GetLatestStateInfoIndex(suite.Ctx, rollappId)
	suite.Require().False(found)

	/*
		_, err := suite.PostStateUpdate(suite.Ctx, rollappId, proposer, 1, uint64(3))
		suite.Require().NoError(err)
	*/

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	// check first index is 1
	expectedLatestStateInfoIndex, found := suite.App.RollappKeeper.GetLatestStateInfoIndex(suite.Ctx, rollappId)
	suite.Require().True(found)
	suite.Require().Equal(expectedLatestStateInfoIndex.Index, uint64(1))
}

func (suite *RollappTestSuite) TestUpdateState() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// parameters
	disputePeriodInBlocks := suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:        urand.RollappID(),
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: rollapp.GetRollappId(),
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// create new update
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	// update state
	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
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

		// calc new updateState
		latestStateInfoIndex, found := suite.App.RollappKeeper.GetLatestStateInfoIndex(suite.Ctx, rollappId)
		suite.Require().EqualValues(true, found)
		// verify index
		suite.Require().EqualValues(i+1, latestStateInfoIndex.Index)
		// load last state info
		expectedStateInfo, found := suite.App.RollappKeeper.GetStateInfo(suite.Ctx, rollappId, latestStateInfoIndex.GetIndex())
		suite.Require().EqualValues(true, found)

		// verify finalization queue
		expectedFinalizationQueue, _ := suite.App.RollappKeeper.GetBlockHeightToFinalizationQueue(suite.Ctx, expectedStateInfo.CreationHeight)
		suite.Require().EqualValues(expectedFinalizationQueue, types.BlockHeightToFinalizationQueue{
			CreationHeight:    expectedStateInfo.CreationHeight,
			FinalizationQueue: []types.StateInfoIndex{latestStateInfoIndex},
		}, "finalization queue", "i", i)

		// create new update
		updateState := types.MsgUpdateState{
			Creator:     bob,
			RollappId:   rollapp.GetRollappId(),
			StartHeight: expectedStateInfo.StartHeight + expectedStateInfo.NumBlocks,
			NumBlocks:   2,
			DAPath:      "",
			BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: expectedStateInfo.StartHeight}, {Height: expectedStateInfo.StartHeight + 1}}},
		}

		// update state
		_, err := suite.PostStateUpdate(suite.Ctx, rollappId, proposer, expectedStateInfo.StartHeight+expectedStateInfo.NumBlocks, uint64(2))
		suite.Require().Nil(err)

		// end block
		suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})

		if uint64(suite.Ctx.BlockHeight()) > disputePeriodInBlocks {
			for i := uint64(1); i <= latestStateInfoIndex.Index; i++ {
				expectedStateInfo, _ := suite.App.RollappKeeper.GetStateInfo(suite.Ctx, rollappId, i)
				if expectedStateInfo.CreationHeight < uint64(suite.Ctx.BlockHeight())-disputePeriodInBlocks {
					suite.Require().EqualValues(expectedStateInfo.Status, common.Status_FINALIZED)
				}
			}
		}

		// check finalization status change
		pendingQueues := suite.App.RollappKeeper.GetAllFinalizationQueueUntilHeightInclusive(suite.Ctx, uint64(suite.Ctx.BlockHeader().Height))
		for _, finalizationQueue := range pendingQueues {
			stateInfo, found := suite.App.RollappKeeper.GetStateInfo(suite.Ctx, finalizationQueue.FinalizationQueue[0].RollappId, finalizationQueue.FinalizationQueue[0].Index)
			suite.Require().True(found)
			suite.Require().EqualValues(stateInfo.Status, common.Status_PENDING)
		}
	}
}

func (suite *RollappTestSuite) TestUpdateStateUnknownRollappId() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// update state of unknown rollapp
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   urand.RollappID(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.EqualError(err, types.ErrUnknownRollappID.Error())
}

// FIXME: need to add sequncer to rollapp to test this scenario
func (suite *RollappTestSuite) TestUpdateStateUnknownSequencer() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:        urand.RollappID(),
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrUnknownSequencer)
}

func (suite *RollappTestSuite) TestUpdateStateSequencerRollappMismatch() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:        urand.RollappID(),
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: urand.RollappID(),
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrSequencerRollappMismatch)
}

func (suite *RollappTestSuite) TestUpdateStateErrLogicUnpermissioned() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	rollappID := urand.RollappID()

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:        rollappID,
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set unpermissioned sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   rollapp.InitialSequencer,
		RollappId: rollappID,
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)
	suite.App.SequencerKeeper.SetProposer(suite.Ctx, "rollapp1", sequencer.SequencerAddress)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   "rollapp1",
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.ErrorIs(err, sequencertypes.ErrUnknownSequencer)
}

func (suite *RollappTestSuite) TestFirstUpdateStateGensisHeightGreaterThanZero() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollappID := urand.RollappID()
	rollapp := types.Rollapp{
		RollappId:        rollappID,
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: rollappID,
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 2,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.NoError(err)
}

func (suite *RollappTestSuite) TestUpdateStateErrWrongBlockHeight() {
	_ = sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollappID := urand.RollappID()
	rollapp := types.Rollapp{
		RollappId:        rollappID,
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: rollappID,
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// set initial latestStateInfoIndex & StateInfo
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: rollappID,
		Index:     1,
	}
	stateInfo := types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappID, Index: 1},
		Sequencer:      sequencer.Address,
		StartHeight:    1,
		NumBlocks:      3,
		DAPath:         "",
		CreationHeight: 0,
		Status:         common.Status_PENDING,
		BDs:            types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}
	suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, latestStateInfoIndex)
	suite.App.RollappKeeper.SetStateInfo(suite.Ctx, stateInfo)

	// bump block height
	suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     proposer,
		RollappId:   rollappId,
		StartHeight: 2,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}, {Height: 4}}},
	}

	_, err := suite.msgServer.UpdateState(suite.Ctx, &updateState)
	suite.ErrorIs(err, types.ErrWrongBlockHeight)
}

func (suite *RollappTestSuite) TestUpdateStateErrLogicMissingStateInfo() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollappID := urand.RollappID()
	rollapp := types.Rollapp{
		RollappId:        rollappID,
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: rollappID,
		Status:    sequencertypes.Bonded,
		Proposer:  true,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// set initial latestStateInfoIndex
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: rollappID,
		Index:     1,
	}
	suite.App.RollappKeeper.SetLatestStateInfoIndex(suite.Ctx, latestStateInfoIndex)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     proposer,
		RollappId:   rollappId,
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := suite.msgServer.UpdateState(suite.Ctx, &updateState)
	suite.ErrorIs(err, types.ErrLogic)
}

func (suite *RollappTestSuite) TestUpdateStateErrNotActiveSequencer() {
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// set rollapp
	rollappID := urand.RollappID()
	rollapp := types.Rollapp{
		RollappId:        rollappID,
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		Bech32Prefix:     "rol",
		GenesisChecksum:  "checksum",
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	// set sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   bob,
		RollappId: rollappID,
		Status:    sequencertypes.Bonded,
	}
	suite.App.SequencerKeeper.SetSequencer(suite.Ctx, sequencer)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
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
	goCtx := sdk.WrapSDKContext(suite.Ctx)
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
				},
			})
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
