package keeper_test

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
)

func (s *RollappTestSuite) TestFirstUpdateState() {
	rollappId, proposer := s.CreateDefaultRollappAndProposer()

	// check no index exists
	_, found := s.k().GetLatestStateInfoIndex(s.Ctx, rollappId)
	s.Require().False(found)

	_, err := s.PostStateUpdate(s.Ctx, rollappId, proposer, 1, uint64(3))
	s.Require().NoError(err)

	// check first index is 1
	expectedLatestStateInfoIndex, found := s.k().GetLatestStateInfoIndex(s.Ctx, rollappId)
	s.Require().True(found)
	s.Require().Equal(expectedLatestStateInfoIndex.Index, uint64(1))
}

func (s *RollappTestSuite) TestUpdateState() {
	// parameters
	disputePeriodInBlocks := s.k().DisputePeriodInBlocks(s.Ctx)

	// set rollapp
	rollappId, proposer := s.CreateDefaultRollappAndProposer()

	// create new update
	_, err := s.PostStateUpdate(s.Ctx, rollappId, proposer, 1, uint64(3))
	s.Require().Nil(err)

	// test 10 update state
	for i := 0; i < 10; i++ {
		// bump block height

		if i == 3 {
			disputePeriodInBlocks += 2
		}

		if i == 6 {
			disputePeriodInBlocks -= 3
		}

		s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeader().Height + 1)

		// calc new updateState
		latestStateInfoIndex, found := s.k().GetLatestStateInfoIndex(s.Ctx, rollappId)
		s.Require().EqualValues(true, found)
		// verify index
		s.Require().EqualValues(i+1, latestStateInfoIndex.Index)
		// load last state info
		expectedStateInfo, found := s.k().GetStateInfo(s.Ctx, rollappId, latestStateInfoIndex.GetIndex())
		s.Require().EqualValues(true, found)

		// verify finalization queue
		expectedFinalizationQueue, _ := s.k().GetFinalizationQueue(s.Ctx, expectedStateInfo.CreationHeight, rollappId)
		s.Require().EqualValues(expectedFinalizationQueue, types.BlockHeightToFinalizationQueue{
			CreationHeight:    expectedStateInfo.CreationHeight,
			FinalizationQueue: []types.StateInfoIndex{latestStateInfoIndex},
			RollappId:         rollappId,
		}, "finalization queue", "i", i)

		// update state
		_, err := s.PostStateUpdate(s.Ctx, rollappId, proposer, expectedStateInfo.StartHeight+expectedStateInfo.NumBlocks, uint64(2))
		s.Require().Nil(err)

		// end block
		s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})

		if uint64(s.Ctx.BlockHeight()) > disputePeriodInBlocks {
			for i := uint64(1); i <= latestStateInfoIndex.Index; i++ {
				expectedStateInfo, _ := s.k().GetStateInfo(s.Ctx, rollappId, i)
				if expectedStateInfo.CreationHeight < uint64(s.Ctx.BlockHeight())-disputePeriodInBlocks {
					s.Require().EqualValues(expectedStateInfo.Status, common.Status_FINALIZED)
				}
			}
		}

		// check finalization status change
		pendingQueues, err := s.k().GetFinalizationQueueUntilHeightInclusive(s.Ctx, uint64(s.Ctx.BlockHeader().Height))
		s.Require().NoError(err)
		for _, finalizationQueue := range pendingQueues {
			stateInfo, found := s.k().GetStateInfo(s.Ctx, finalizationQueue.FinalizationQueue[0].RollappId, finalizationQueue.FinalizationQueue[0].Index)
			s.Require().True(found)
			s.Require().EqualValues(stateInfo.Status, common.Status_PENDING)
		}

		s.checkLiveness(rollappId, true, true)
	}
}

func (s *RollappTestSuite) TestUpdateStateObsoleteRollapp() {
	const (
		raName             = "rollapptest_1-1"
		nonObsoleteVersion = 2
		obsoleteVersion    = 1
	)

	// create a rollapp
	s.CreateRollappByName(raName)
	// create a sequencer
	proposer := s.CreateDefaultSequencer(s.Ctx, raName)

	// create the initial state update with non-obsolete version
	expectedNextHeight, err := s.PostStateUpdateWithDRSVersion(s.Ctx, raName, proposer, 1, uint64(3), nonObsoleteVersion)
	s.Require().Nil(err)

	// check the rollapp's last height
	actualLastHeight := s.GetRollappLastHeight(raName)
	s.Require().Equal(expectedNextHeight-1, actualLastHeight)

	// mark a DRS version as obsolete
	err = s.k().SetObsoleteDRSVersion(s.Ctx, obsoleteVersion)
	s.Require().NoError(err)

	// create a new update using the obsolete version
	_, err = s.PostStateUpdateWithDRSVersion(s.Ctx, raName, proposer, expectedNextHeight, uint64(3), obsoleteVersion)
	s.Require().Error(err)
}

func (s *RollappTestSuite) TestUpdateStateUnknownRollappId() {
	_, err := s.PostStateUpdate(s.Ctx, "unknown_rollapp", alice, 1, uint64(3))
	s.EqualError(err, types.ErrUnknownRollappID.Error())
}

func (s *RollappTestSuite) TestUpdateStateUnknownSequencer() {
	rollappId, _ := s.CreateDefaultRollappAndProposer()

	// update state
	_, err := s.PostStateUpdate(s.Ctx, rollappId, bob, 1, uint64(3))
	s.ErrorIs(err, sequencertypes.ErrNotProposer)
}

func (s *RollappTestSuite) TestUpdateStateSequencerRollappMismatch() {
	s.SetupTest()

	rollappId, _ := s.CreateDefaultRollappAndProposer()
	_, seq_2 := s.CreateDefaultRollappAndProposer()

	// update state from proposer of rollapp2
	_, err := s.PostStateUpdate(s.Ctx, rollappId, seq_2, 1, uint64(3))
	s.ErrorIs(err, sequencertypes.ErrNotProposer)
}

func (s *RollappTestSuite) TestUpdateStateErrLogicUnpermissioned() {
	goCtx := sdk.WrapSDKContext(s.Ctx)

	rollappID := urand.RollappID()

	// set rollapp
	rollapp := types.Rollapp{
		RollappId:        rollappID,
		Owner:            alice,
		InitialSequencer: sample.AccAddress(),
		GenesisInfo:      *mockGenesisInfo,
	}
	s.k().SetRollapp(s.Ctx, rollapp)

	// set unpermissioned sequencer
	sequencer := sequencertypes.Sequencer{
		Address:   rollapp.InitialSequencer,
		RollappId: rollappID,
		Status:    sequencertypes.Bonded,
	}
	s.App.SequencerKeeper.SetSequencer(s.Ctx, sequencer)
	s.App.SequencerKeeper.SetProposer(s.Ctx, rollappID, sequencer.Address)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollappID,
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := s.msgServer.UpdateState(goCtx, &updateState)
	s.ErrorIs(err, sequencertypes.ErrNotProposer)
}

func (s *RollappTestSuite) TestFirstUpdateStateGenesisHeightGreaterThanZero() {
	rollappId, proposer := s.CreateDefaultRollappAndProposer()

	_, err := s.PostStateUpdate(s.Ctx, rollappId, proposer, 3, uint64(3))
	s.NoError(err)
}

func (s *RollappTestSuite) TestUpdateStateErrWrongBlockHeight() {
	rollappId, proposer := s.CreateDefaultRollappAndProposer()

	// set initial latestStateInfoIndex & StateInfo
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: rollappId,
		Index:     1,
	}
	stateInfo := types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappId, Index: 1},
		Sequencer:      proposer,
		StartHeight:    1,
		NumBlocks:      3,
		Status:         common.Status_PENDING,
		BDs:            types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}
	s.k().SetLatestStateInfoIndex(s.Ctx, latestStateInfoIndex)
	s.k().SetStateInfo(s.Ctx, stateInfo)

	// bump block height
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeader().Height + 1)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     proposer,
		RollappId:   rollappId,
		StartHeight: 2,
		NumBlocks:   3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2}, {Height: 3}, {Height: 4}}},
	}

	_, err := s.msgServer.UpdateState(s.Ctx, &updateState)
	s.ErrorIs(err, types.ErrWrongBlockHeight)
}

func (s *RollappTestSuite) TestUpdateStateErrLogicMissingStateInfo() {
	rollappId, proposer := s.CreateDefaultRollappAndProposer()

	// set initial latestStateInfoIndex
	latestStateInfoIndex := types.StateInfoIndex{
		RollappId: rollappId,
		Index:     1,
	}
	s.k().SetLatestStateInfoIndex(s.Ctx, latestStateInfoIndex)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     proposer,
		RollappId:   rollappId,
		StartHeight: 1,
		NumBlocks:   3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 2}, {Height: 3}}},
	}

	_, err := s.msgServer.UpdateState(s.Ctx, &updateState)
	s.ErrorIs(err, types.ErrLogic)
}

func (s *RollappTestSuite) TestUpdateStateErrNotActiveSequencer() {
	rollappId, _ := s.CreateDefaultRollappAndProposer()
	addr2 := s.CreateDefaultSequencer(s.Ctx, rollappId) // non-proposer

	// update state from bob
	_, err := s.PostStateUpdate(s.Ctx, rollappId, addr2, 1, uint64(3))
	s.ErrorIs(err, sequencertypes.ErrNotProposer)
}

func (s *RollappTestSuite) TestUpdateStateDowngradeTimestamp() {
	rollappId, proposer := s.CreateDefaultRollappAndProposer()
	// update state without timestamp
	stateInfo := types.StateInfo{
		StateInfoIndex: types.StateInfoIndex{RollappId: rollappId, Index: 1},
		Sequencer:      proposer,
		StartHeight:    1,
		NumBlocks:      1,
		DAPath:         "",
		BDs:            types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}}},
	}
	s.k().SetLatestStateInfoIndex(s.Ctx, stateInfo.StateInfoIndex)
	s.k().SetStateInfo(s.Ctx, stateInfo)

	// update state with timestamp - this "upgrades" the rollapp such that all new state updates must have timestamp in BD
	updateState := types.MsgUpdateState{
		Creator:     proposer,
		RollappId:   rollappId,
		StartHeight: 2,
		NumBlocks:   1,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 2, Timestamp: time.Now().UTC()}}},
	}
	_, err := s.msgServer.UpdateState(s.Ctx, &updateState)
	s.NoError(err)

	// update state without timestamp
	updateState = types.MsgUpdateState{
		Creator:     proposer,
		RollappId:   rollappId,
		StartHeight: 3,
		NumBlocks:   1,
		DAPath:      "",
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 3}}},
	}
	_, err = s.msgServer.UpdateState(s.Ctx, &updateState)
	s.ErrorIs(err, types.ErrInvalidBlockDescriptorTimestamp)
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
			rollappsRes[rollappRes.Summary.GetRollappId()] = &rollappRes.Summary
		}
		totalChecked += len(queryAllResponse.GetRollapp())
		nextKey = queryAllResponse.GetPagination().GetNextKey()

		if nextKey == nil {
			break
		}
	}

	return rollappsRes, totalRes
}
