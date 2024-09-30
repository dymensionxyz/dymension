package keeper_test

import (
	"github.com/cometbft/cometbft/libs/rand"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	damodule "github.com/dymensionxyz/dymension/v3/x/delayedack"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (suite *DelayedAckTestSuite) TestInvariants() {
	suite.SetupTest()

	transferStack := damodule.NewIBCMiddleware(
		damodule.WithIBCModule(ibctransfer.NewIBCModule(suite.App.TransferKeeper)),
		damodule.WithKeeper(suite.App.DelayedAckKeeper),
		damodule.WithRollappKeeper(suite.App.RollappKeeper),
	)

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
					Packet:      getNewTestPacket(sequence),
					Status:      commontypes.Status_PENDING,
					ProofHeight: proofHeight,
				}
				suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, rollappPacket)

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
	for rollapp := range seqPerRollapp {
		packetsNum, err := suite.App.DelayedAckKeeper.FinalizeRollappPacketsUntilHeight(suite.Ctx, transferStack.NextIBCMiddleware(), rollapp, rollappBlocks[rollapp])
		suite.Require().NoError(err)
		suite.Require().Equal(rollappBlocks[rollapp], uint64(packetsNum))
	}

	// test fraud
	for rollapp := range seqPerRollapp {
		err := suite.App.DelayedAckKeeper.HandleFraud(suite.Ctx, rollapp, transferStack)
		suite.Require().NoError(err)
		break
	}

	// check invariant
	msg, fails := suite.App.DelayedAckKeeper.PacketsFinalizationCorrespondsToFinalizationHeight(suite.Ctx)
	suite.Require().False(fails, msg)
	msg, fails = suite.App.DelayedAckKeeper.PacketsFromRevertedHeightsAreReverted(suite.Ctx)
	suite.Require().False(fails, msg)
}

func (suite *DelayedAckTestSuite) TestRollappPacketsCasesInvariant() {
	rollapp := "rollapp_1234-1"

	cases := []struct {
		name             string
		frozenRollapp    bool
		allFinalized     bool
		nothingFinalized bool
		packet           commontypes.RollappPacket
		packet2          commontypes.RollappPacket
		expectedIsBroken bool
	}{
		{
			"successful invariant check - packets are finalized only for finalized heights",
			false,
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
			"successful revert check - packets are reverted for non-finalized states",
			true,
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
				Status:      commontypes.Status_REVERTED,
				ProofHeight: 15,
				Packet:      getNewTestPacket(2),
			},
			false,
		},
		{
			"successful non-finalized state invariant check - packets without finalization state are not finalized",
			false,
			false,
			true,
			commontypes.RollappPacket{
				RollappId:   rollapp,
				Status:      commontypes.Status_PENDING,
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
			"wrong invariant revert check - packets for frozen rollapps in non-finalized heights are not reverted",
			true,
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
			true,
		},
		{
			"wrong finalized packet check - packets are finalized in non-finalized heights",
			false,
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
			"wrong finalized packet check - packets for non-finalized heights are finalized",
			false,
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
				Status:      commontypes.Status_FINALIZED,
				ProofHeight: 15,
				Packet:      getNewTestPacket(2),
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
				Status:      commontypes.Status_FINALIZED,
				Sequencer:   proposer,
			}.WithNumBlocks(10)
			stateInfo2 := rollapptypes.StateInfo{
				StateInfoIndex: rollapptypes.StateInfoIndex{
					RollappId: rollapp,
					Index:     2,
				},
				StartHeight: 11,
				Status:      commontypes.Status_PENDING,
				Sequencer:   proposer,
			}.WithNumBlocks(10)

			// if nothingFinalized true, all the state infos submitted should be pending
			if tc.nothingFinalized {
				stateInfo.Status = commontypes.Status_PENDING
			} else {
				suite.App.RollappKeeper.SetLatestFinalizedStateIndex(ctx, types.StateInfoIndex{
					RollappId: rollapp,
					Index:     stateInfo.GetIndex().Index,
				})
			}

			suite.App.RollappKeeper.SetStateInfo(ctx, stateInfo)

			// if allFinalized true, all the state infos submitted should be finalized
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

			// if frozenRollapp true, we should freeze the rollapp and revert pending states
			if tc.frozenRollapp {
				err := suite.App.RollappKeeper.HandleFraud(ctx, rollapp, "", 11, proposer)
				suite.Require().NoError(err)
			}

			// add rollapp packets
			suite.App.DelayedAckKeeper.SetRollappPacket(ctx, tc.packet)
			suite.App.DelayedAckKeeper.SetRollappPacket(ctx, tc.packet2)

			// check invariant
			_, failsFinalize := suite.App.DelayedAckKeeper.PacketsFinalizationCorrespondsToFinalizationHeight(suite.Ctx)
			_, failsRevert := suite.App.DelayedAckKeeper.PacketsFromRevertedHeightsAreReverted(suite.Ctx)

			isBroken := failsFinalize || failsRevert
			suite.Require().Equal(tc.expectedIsBroken, isBroken)
		})
	}
}

func getNewTestPacket(sequence uint64) *channeltypes.Packet {
	return &channeltypes.Packet{
		SourcePort:         "testSourcePort",
		SourceChannel:      testSourceChannel,
		DestinationPort:    "testDestinationPort",
		DestinationChannel: "testDestinationChannel",
		Data:               []byte("testData"),
		Sequence:           sequence,
	}
}
