package keeper_test

import (
	"math/rand"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
)

func (suite *DelayedAckTestSuite) TestInvariants() {
	suite.SetupTest()
	ctx := suite.Ctx
	initialheight := int64(10)
	suite.Ctx = suite.Ctx.WithBlockHeight(initialheight)

	numOfRollapps := 10
	numOfStates := 10
	//create rollapps
	seqPerRollapp := make(map[string]string)
	rollappBlocks := make(map[string]int)

	for i := 0; i < numOfRollapps; i++ {
		rollapp := suite.CreateDefaultRollapp()
		seqaddr := suite.CreateDefaultSequencer(ctx, rollapp)

		//skip one of the rollapps so it won't have any state updates
		if i == 0 {
			continue
		}
		seqPerRollapp[rollapp] = seqaddr
		rollappBlocks[rollapp] = 0

	}

	//send state updates
	var lastHeight uint64 = 0

	sequence := 0
	for j := 0; j < numOfStates; j++ {
		numOfBlocks := uint64(rand.Intn(10) + 1)
		for rollapp := range seqPerRollapp {
			_, err := suite.PostStateUpdate(suite.Ctx, rollapp, seqPerRollapp[rollapp], lastHeight+1, numOfBlocks)
			suite.Require().Nil(err)
			for k := 1; k <= int(numOfBlocks); k++ {
				rollappPacket := &commontypes.RollappPacket{
					RollappId: rollapp,
					Packet: &channeltypes.Packet{
						SourcePort:         "testSourcePort",
						SourceChannel:      "testSourceChannel",
						DestinationPort:    "testDestinationPort",
						DestinationChannel: "testDestinationChannel",
						Data:               []byte("testData"),
						Sequence:           uint64(sequence),
					},
					Status:      commontypes.Status_PENDING,
					ProofHeight: uint64(rollappBlocks[rollapp] + k),
				}
				suite.App.DelayedAckKeeper.SetRollappPacket(ctx, *rollappPacket)
				sequence++
			}
			rollappBlocks[rollapp] = rollappBlocks[rollapp] + int(numOfBlocks)

		}

		suite.Ctx = suite.Ctx.WithBlockHeight(suite.Ctx.BlockHeader().Height + 1)
		lastHeight = lastHeight + numOfBlocks
	}

	//progress finalization queue
	suite.App.RollappKeeper.FinalizeQueue(suite.Ctx)

	//test fraud
	for rollapp := range seqPerRollapp {
		err := suite.App.DelayedAckKeeper.HandleFraud(ctx, rollapp)
		suite.Require().Nil(err)
		break
	}

	// check invariant
	msg, bool := keeper.AllInvariants(suite.App.DelayedAckKeeper)(suite.Ctx)
	suite.Require().False(bool, msg)
}
