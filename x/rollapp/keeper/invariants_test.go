package keeper_test

import (
	"github.com/cometbft/cometbft/libs/rand"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *RollappTestSuite) TestInvariants() {
	suite.SetupTest()
	initialheight := int64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight)

	numOfRollapps := 10
	numOfStates := 10

	// create rollapps
	seqPerRollapp := make(map[string]string)
	for i := 0; i < numOfRollapps; i++ {
		rollapp, seqaddr := suite.CreateDefaultRollappWithProposer()

		// skip one of the rollapps so it won't have any state updates
		if i == 0 {
			continue
		}
		seqPerRollapp[rollapp] = seqaddr
	}

	rollapp, seqaddr := suite.CreateRollappWithNameWithProposer("dym_1100-1")
	seqPerRollapp[rollapp] = seqaddr

	rollapp, seqaddr = suite.CreateRollappWithNameWithProposer("dym_1100")
	seqPerRollapp[rollapp] = seqaddr

	// send state updates
	var lastHeight uint64 = 0
	for j := 0; j < numOfStates; j++ {
		numOfBlocks := uint64(rand.Intn(10) + 1)
		for rollapp := range seqPerRollapp {
			_, err := suite.PostStateUpdate(suite.Ctx, rollapp, seqPerRollapp[rollapp], lastHeight+1, numOfBlocks)
			suite.Require().Nil(err)
		}

		suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
		lastHeight = lastHeight + numOfBlocks
	}

	// progress finalization queue
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight + 2)
	suite.App.RollappKeeper.FinalizeRollappStates(suite.Ctx)

	// check invariant
	msg, ok := keeper.AllInvariants(*suite.App.RollappKeeper)(suite.Ctx)
	suite.Require().False(ok, msg)
}

func (suite *RollappTestSuite) TestRollappFinalizedStateInvariant() {
	suite.SetupTest()
	ctx := suite.Ctx
	rollapp1, rollapp2, rollapp3 := "rollapp1", "rollapp2", "rollapp3"
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
			"rollapp1",
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
		suite.Run(tc.name, func() {
			// create rollapp
			suite.CreateRollappWithNameWithProposer(tc.rollappId)
			// update state infos
			if tc.stateInfo != nil {
				suite.App.RollappKeeper.SetStateInfo(ctx, *tc.stateInfo)
			}
			// update latest finalized state info
			suite.App.RollappKeeper.SetStateInfo(ctx, tc.latestFinalizedStateInfo)
			suite.App.RollappKeeper.SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{
				RollappId: tc.rollappId,
				Index:     tc.latestFinalizedStateInfo.GetIndex().Index,
			})
			// update latest state info index
			suite.App.RollappKeeper.SetStateInfo(ctx, tc.latestStateInfoIndex)
			suite.App.RollappKeeper.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
				RollappId: tc.rollappId,
				Index:     tc.latestStateInfoIndex.GetIndex().Index,
			})
			// check invariant
			_, isBroken := keeper.RollappFinalizedStateInvariant(*suite.App.RollappKeeper)(ctx)
			suite.Require().Equal(tc.expectedIsBroken, isBroken)
		})
	}
}
