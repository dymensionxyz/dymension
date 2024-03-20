package keeper_test

import (
	"github.com/tendermint/tendermint/libs/rand"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *DelayedAckTestSuite) TestInvariants() {
	suite.SetupTest()
	initialheight := int64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight)

	numOfRollapps := 10
	numOfStates := 10
	// create rollapps
	seqPerRollapp := make(map[string]string)
	rollappBlocks := make(map[string]int)

	for i := 0; i < numOfRollapps; i++ {
		rollapp := suite.CreateDefaultRollapp()
		seqaddr := suite.CreateDefaultSequencer(suite.Ctx, rollapp)

		// skip one of the rollapps so it won't have any state updates
		if i == 0 {
			continue
		}
		seqPerRollapp[rollapp] = seqaddr
		rollappBlocks[rollapp] = 0

	}

	// send state updates
	var lastHeight uint64 = 0

	sequence := uint64(0)
	for j := 0; j < numOfStates; j++ {
		numOfBlocks := uint64(rand.Intn(10) + 1)
		for rollapp := range seqPerRollapp {
			_, err := suite.PostStateUpdate(suite.Ctx, rollapp, seqPerRollapp[rollapp], lastHeight+1, numOfBlocks)
			suite.Require().Nil(err)
			for k := 1; k <= int(numOfBlocks); k++ {
				rollappPacket := &commontypes.RollappPacket{
					RollappId:   rollapp,
					Packet:      getNewTestPacket(sequence),
					Status:      commontypes.Status_PENDING,
					ProofHeight: uint64(rollappBlocks[rollapp] + k),
				}
				err := suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
				suite.Require().NoError(err)

				sequence++
			}
			rollappBlocks[rollapp] = rollappBlocks[rollapp] + int(numOfBlocks)

		}

		suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
		lastHeight = lastHeight + numOfBlocks
	}

	// progress finalization queue
	err := suite.App.RollappKeeper.FinalizeQueue(suite.Ctx)
	suite.Require().NoError(err)

	// test fraud
	for rollapp := range seqPerRollapp {
		err := suite.App.DelayedAckKeeper.HandleFraud(suite.Ctx, rollapp)
		suite.Require().Nil(err)
		break
	}

	// check invariant
	msg, bool := keeper.AllInvariants(suite.App.DelayedAckKeeper)(suite.Ctx)
	suite.Require().False(bool, msg)
}

func (suite *DelayedAckTestSuite) TestRollappPacketsCasesInvariant() {
	suite.SetupTest()
	ctx := suite.Ctx
	rollapp := "rollapp1"

	cases := []struct {
		name             string
		frozenRollapp    bool
		allFinalized     bool
		packet           commontypes.RollappPacket
		packet2          commontypes.RollappPacket
		expectedIsBroken bool
	}{
		{
			"successful invariant check",
			false,
			false,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      getNewTestPacket(1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      getNewTestPacket(2),
			},
			false,
		},
		{
			"successful revert check",
			true,
			false,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      getNewTestPacket(1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_REVERTED,
				ProofHeight: 15,
				Packet:      getNewTestPacket(2),
			},
			false,
		},
		{
			"error non-finalized packet",
			false,
			true,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      getNewTestPacket(1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      getNewTestPacket(2),
			},
			true,
		},
		{
			"wrong invariant revert check",
			true,
			false,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      getNewTestPacket(1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      getNewTestPacket(2),
			},
			true,
		},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {

			// create rollapp
			suite.CreateRollappWithName(rollapp)
			proposer := suite.CreateDefaultSequencer(ctx, rollapp)

			// update state infos
			stateInfo := rollapptypes.StateInfo{
				StateInfoIndex: rollapptypes.StateInfoIndex{
					RollappId: rollapp,
					Index:     1,
				},
				StartHeight: 1,
				NumBlocks:   10,
				Status:      commontypes.Status_FINALIZED,
				Sequencer:   proposer,
			}
			stateInfo2 := rollapptypes.StateInfo{
				StateInfoIndex: rollapptypes.StateInfoIndex{
					RollappId: rollapp,
					Index:     2,
				},
				StartHeight: 10,
				NumBlocks:   10,
				Status:      commontypes.Status_PENDING,
				Sequencer:   proposer,
			}

			suite.App.RollappKeeper.SetStateInfo(ctx, stateInfo)
			suite.App.RollappKeeper.SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{
				RollappId: rollapp,
				Index:     stateInfo.GetIndex().Index,
			})
			if tc.allFinalized {
				stateInfo2.Status = commontypes.Status_FINALIZED
			}
			suite.App.RollappKeeper.SetStateInfo(ctx, stateInfo2)
			if stateInfo2.Status == commontypes.Status_FINALIZED {
				suite.App.RollappKeeper.SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{
					RollappId: rollapp,
					Index:     stateInfo2.GetIndex().Index,
				})
			}
			suite.App.RollappKeeper.SetLatestStateInfoIndex(ctx, types.StateInfoIndex{
				RollappId: rollapp,
				Index:     stateInfo2.GetIndex().Index,
			})

			if tc.frozenRollapp {
				err := suite.App.RollappKeeper.HandleFraud(ctx, rollapp, "", 11, proposer)
				suite.Require().NoError(err)
			}

			// add rollapp packets
			err := suite.App.DelayedAckKeeper.SetRollappPacket(ctx, tc.packet)
			suite.Require().NoError(err)
			err = suite.App.DelayedAckKeeper.SetRollappPacket(ctx, tc.packet2)
			suite.Require().NoError(err)

			// check invariant
			_, isBroken := keeper.AllInvariants(suite.App.DelayedAckKeeper)(suite.Ctx)
			suite.Require().Equal(tc.expectedIsBroken, isBroken)
		})
	}
}

func getNewTestPacket(sequence uint64) *channeltypes.Packet {
	return &channeltypes.Packet{
		SourcePort:         "testSourcePort",
		SourceChannel:      "testSourceChannel",
		DestinationPort:    "testDestinationPort",
		DestinationChannel: "testDestinationChannel",
		Data:               []byte("testData"),
		Sequence:           sequence,
	}
}
