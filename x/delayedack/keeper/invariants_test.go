package keeper_test

import (
	"github.com/cometbft/cometbft/libs/rand"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *DelayedAckTestSuite) TestInvariants() {
	initialHeight := int64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(initialHeight)

	numOfRollapps := 10
	numOfStates := 10
	// create rollapps
	seqPerRollapp := make(map[string]string)
	rollappBlocks := make(map[string]uint64)

	for i := 0; i < numOfRollapps; i++ {
		rollapp, seqaddr := suite.CreateDefaultRollappAndProposer()

		// skip one of the rollapps so it won't have any state updates
		if i == 0 {
			continue
		}
		seqPerRollapp[rollapp] = seqaddr
		rollappBlocks[rollapp] = 0
	}

	sequence := uint64(0)
	for j := 0; j < numOfStates; j++ {
		numOfBlocks := uint64(rand.Intn(10) + 1)
		for rollapp, sequencer := range seqPerRollapp {
			_, err := suite.PostStateUpdate(suite.Ctx, rollapp, sequencer, rollappBlocks[rollapp]+uint64(1), numOfBlocks)
			suite.Require().NoError(err)

			for k := uint64(1); k <= numOfBlocks; k++ {
				// calculating a different proof height incrementing a block height for each new packet
				proofHeight := rollappBlocks[rollapp] + k
				rollappPacket := commontypes.RollappPacket{
					RollappId:   rollapp,
					Packet:      apptesting.GenerateTestPacket(suite.T(), sequence),
					Status:      commontypes.Status_PENDING,
					ProofHeight: proofHeight,
				}
				suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, rollappPacket)
				err = suite.App.DelayedAckKeeper.SetPendingPacketByAddress(suite.Ctx, apptesting.TestPacketReceiver, rollappPacket.RollappPacketKey())
				suite.Require().NoError(err)

				sequence++
			}

			rollappBlocks[rollapp] = rollappBlocks[rollapp] + numOfBlocks
		}
		suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
	}

	// skip a dispute period
	disputePeriod := int64(suite.App.RollappKeeper.DisputePeriodInBlocks(suite.Ctx))
	suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + disputePeriod)

	// progress finalization queue
	suite.App.RollappKeeper.FinalizeRollappStates(suite.Ctx)

	// manually finalize packets for all rollapps
	finalizedNum := suite.FinalizeAllPendingPackets(apptesting.TestPacketReceiver)
	// check the total number of packets
	var total int
	for rollapp := range seqPerRollapp {
		total += int(rollappBlocks[rollapp])
	}
	suite.Require().Equal(total, finalizedNum)

	// test fraud
	for rollapp := range seqPerRollapp {
		err := suite.App.DelayedAckKeeper.OnHardFork(suite.Ctx, rollapp, uint64(suite.Ctx.BlockHeight()))
		suite.Require().NoError(err)
	}

	msg, fails := keeper.AllInvariants(suite.App.DelayedAckKeeper)(suite.Ctx)
	suite.Require().False(fails, msg)
}

// TestRollappPacketsCasesInvariant tests the invariant that checks if the packets are finalized only for finalized heights
// by default, we have:
// - state1 with blocks 1-10 which is finalized
// - state2 with blocks 11-20 which is pending
func (suite *DelayedAckTestSuite) TestRollappPacketsCasesInvariant() {
	suite.T().Skip("skipping TestRollappPacketsCasesInvariant as it's not supported with lazy finalization feature")

	rollapp := "rollapp_1234-1"
	cases := []struct {
		name             string
		nothingFinalized bool
		packet           commontypes.RollappPacket
		packet2          commontypes.RollappPacket
		expectedIsBroken bool
	}{
		// successful checks
		{
			"successful invariant check - packets are finalized only for finalized heights",
			false,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 2),
			},
			false,
		},
		{
			"successful invariant check - packets are not yet finalized for finalized heights",
			false,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 5,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 2),
			},
			false,
		},
		{
			"successful non-finalized state invariant check - packets without finalization state are not finalized",
			true,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 5,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 2),
			},
			false,
		},
		// failed checks
		{
			"wrong finalized packet check - packets are finalized in non-finalized heights",
			true,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
				ProofHeight: 15,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 2),
			},
			true,
		},
		{
			"wrong finalized packet check - packets for non-finalized heights are finalized",
			false,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 5,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 1),
			},
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 15,
				Packet:      apptesting.GenerateTestPacket(suite.T(), 2),
			},
			true,
		},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			// create rollapp
			suite.SetupTest()
			ctx := suite.Ctx
			suite.CreateRollappByName(rollapp)
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
			if tc.nothingFinalized {
				stateInfo.Status = commontypes.Status_PENDING
			}
			suite.App.RollappKeeper.SetStateInfo(ctx, stateInfo)

			stateInfo2 := rollapptypes.StateInfo{
				StateInfoIndex: rollapptypes.StateInfoIndex{
					RollappId: rollapp,
					Index:     2,
				},
				StartHeight: 11,
				NumBlocks:   10,
				Status:      commontypes.Status_PENDING,
				Sequencer:   proposer,
			}
			suite.App.RollappKeeper.SetStateInfo(ctx, stateInfo2)
			suite.App.RollappKeeper.SetLatestStateInfoIndex(ctx, stateInfo2.StateInfoIndex)
			if !tc.nothingFinalized {
				suite.App.RollappKeeper.SetLatestFinalizedStateIndex(ctx, stateInfo.StateInfoIndex)
			}

			// add rollapp packets
			suite.App.DelayedAckKeeper.SetRollappPacket(ctx, tc.packet)
			suite.App.DelayedAckKeeper.SetRollappPacket(ctx, tc.packet2)

			_, fails := keeper.AllInvariants(suite.App.DelayedAckKeeper)(suite.Ctx)
			suite.Require().Equal(tc.expectedIsBroken, fails)
		})
	}
}
