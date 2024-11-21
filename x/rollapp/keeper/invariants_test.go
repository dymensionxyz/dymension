package keeper_test

import (
	"github.com/cometbft/cometbft/libs/rand"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *RollappTestSuite) TestInvariants() {
	initialheight := int64(10)
	s.Ctx = s.Ctx.WithBlockHeight(initialheight)

	numOfRollapps := 10
	numOfStates := 10

	// create rollapps
	seqPerRollapp := make(map[string]string)
	for i := 0; i < numOfRollapps; i++ {
		rollapp, seqaddr := s.CreateDefaultRollappAndProposer()

		// skip one of the rollapps so it won't have any state updates
		if i == 0 {
			continue
		}
		seqPerRollapp[rollapp] = seqaddr
	}

	rollapp, seqaddr := s.CreateDefaultRollappAndProposer()
	seqPerRollapp[rollapp] = seqaddr

	rollapp, seqaddr = s.CreateDefaultRollappAndProposer()
	seqPerRollapp[rollapp] = seqaddr

	// send state updates
	var lastHeight uint64 = 0
	for j := 0; j < numOfStates; j++ {
		numOfBlocks := uint64(rand.Intn(10) + 1)
		for rollapp := range seqPerRollapp {
			_, err := s.PostStateUpdate(s.Ctx, rollapp, seqPerRollapp[rollapp], lastHeight+1, numOfBlocks)
			s.Require().Nil(err)
		}

		s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeader().Height + 1)
		lastHeight = lastHeight + numOfBlocks
	}

	// progress finalization queue
	s.Ctx = s.Ctx.WithBlockHeight(initialheight + 2)
	s.k().FinalizeRollappStates(s.Ctx)

	// check invariant
	msg, ok := keeper.AllInvariants(*s.k())(s.Ctx)
	s.Require().False(ok, msg)
}

func (s *RollappTestSuite) TestRollappFinalizedStateInvariant() {
	ctx := s.Ctx
	rollapp1, rollapp2, rollapp3 := "rollapp_1234-1", "unrollapp_2345-1", "trollapp_3456-1"
	cases := []struct {
		name                     string
		rollappId                string
		stateInfo                *types.StateInfo
		latestFinalizedStateInfo types.StateInfo
		latestStateInfoIndex     types.StateInfo
		expectedIsBroken         bool
	}{
		{
			"successful invariant check",
			rollapp1,
			&types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp1,
					Index:     1,
				},
				Status: commontypes.Status_FINALIZED,
			},
			types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp1,
					Index:     2,
				},
				Status: commontypes.Status_FINALIZED,
			},
			types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp1,
					Index:     3,
				},
				Status: commontypes.Status_PENDING,
			},
			false,
		},
		{
			"failed invariant check - state not found",
			rollapp2,
			nil,
			types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp2,
					Index:     2,
				},
				Status: commontypes.Status_FINALIZED,
			},
			types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp2,
					Index:     3,
				},
				Status: commontypes.Status_PENDING,
			},
			true,
		},
		{
			"failed invariant check - state not finalized",
			rollapp3,
			&types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp3,
					Index:     1,
				},
				Status: commontypes.Status_PENDING,
			},
			types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp3,
					Index:     2,
				},
				Status: commontypes.Status_FINALIZED,
			},
			types.StateInfo{
				StateInfoIndex: types.StateInfoIndex{
					RollappId: rollapp3,
					Index:     3,
				},
				Status: commontypes.Status_PENDING,
			},
			true,
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			// create rollapp
			s.CreateRollappByName(tc.rollappId)
			// create sequencer
			s.CreateDefaultSequencer(ctx, tc.rollappId)
			// update state infos
			if tc.stateInfo != nil {
				s.k().SetStateInfo(ctx, *tc.stateInfo)
			}
			// update latest finalized state info
			s.k().SetStateInfo(ctx, tc.latestFinalizedStateInfo)
			s.k().SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{
				RollappId: tc.rollappId,
				Index:     tc.latestFinalizedStateInfo.GetIndex().Index,
			})
			// update latest state info index
			s.k().SetStateInfo(ctx, tc.latestStateInfoIndex)
			s.k().SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
				RollappId: tc.rollappId,
				Index:     tc.latestStateInfoIndex.GetIndex().Index,
			})
			// check invariant
			_, isBroken := keeper.RollappFinalizedStateInvariant(*s.k())(ctx)
			s.Require().Equal(tc.expectedIsBroken, isBroken)
		})
	}
}
