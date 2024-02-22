package keeper_test

import (
	"context"
	"encoding/json"
	"log"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *RollappTestSuite) createRollappBatch(goCtx context.Context) (string, uint64) {

	file, err := os.Open("../../../app/dainclusionproofs/non_inclusion_proof.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	proof := types.NonInclusionProof{}
	err = jsonDecoder.Decode(&proof)
	suite.Require().Nil(err)

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
		RollappIDs:       []string{rollapp.GetRollappId()},
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// check no index exists
	_, found := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(false, found)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "celestia.1.2.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef." + string(proof.Dataroot),
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err = suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	stateIndexInfo, _ := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	return stateIndexInfo.GetRollappId(), stateIndexInfo.GetIndex()
}

func (suite *RollappTestSuite) createRollappBatchWrongDataRoot(goCtx context.Context) (string, uint64) {

	file, err := os.Open("../../../app/dainclusionproofs/non_inclusion_proof.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	proof := types.NonInclusionProof{}
	err = jsonDecoder.Decode(&proof)
	suite.Require().Nil(err)

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
		RollappIDs:       []string{rollapp.GetRollappId()},
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// check no index exists
	_, found := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(false, found)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "celestia.1.2.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef.V6G1mPCsYwQevmqXpQltHKqBNP8ZL9N0hHpzQx4ipL0=",
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err = suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	stateIndexInfo, _ := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	return stateIndexInfo.GetRollappId(), stateIndexInfo.GetIndex()
}

func (suite *RollappTestSuite) createRollappBatchOutofSquare(goCtx context.Context) (string, uint64) {

	file, err := os.Open("../../../app/dainclusionproofs/non_inclusion_proof.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	proof := types.NonInclusionProof{}
	err = jsonDecoder.Decode(&proof)
	suite.Require().Nil(err)

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
		RollappIDs:       []string{rollapp.GetRollappId()},
	}
	suite.app.SequencerKeeper.SetSequencer(suite.ctx, sequencer)
	// register sequncer in sequencer as Proposer
	scheduler := sequencertypes.Scheduler{
		SequencerAddress: bob,
		Status:           sequencertypes.Proposer,
	}
	suite.app.SequencerKeeper.SetScheduler(suite.ctx, scheduler)

	// check no index exists
	_, found := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	suite.Require().EqualValues(false, found)

	// update state
	updateState := types.MsgUpdateState{
		Creator:     bob,
		RollappId:   rollapp.GetRollappId(),
		StartHeight: 1,
		NumBlocks:   3,
		DAPath:      "celestia.1.1340.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef." + string(proof.Dataroot),
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err = suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	stateIndexInfo, _ := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	return stateIndexInfo.GetRollappId(), stateIndexInfo.GetIndex()
}

func (suite *RollappTestSuite) TestValidNonAvaliableSubmission() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappBatch(goCtx)
	nonAvailableBatch := types.MsgNonAvailableBatch{
		Creator:           carol,
		RollappId:         rollapId,
		SlIndex:           SLIndex,
		DAPath:            "celestia.1.2.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef.T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
		NonInclusionProof: "{\"rproofs\": \"CCAaIEPfyrlgwKe73bHoJPu/s3g/cqpCD+GSBhprXt6dQKvfIiCxvoDp0he9J928CKGi9YRyq6e/TO3VhMmtf31kZeA9XSIgVQgl6ThdDbM23Jb12Qpj2432o8n326zTSifjYdENdwMiIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiDWJEaMImR0qENxIHxcF3dTz4zl4JNLS4qU/bst0GdwsSIgvHPiMyjd1cVH4kYeTZrUUQQfxAeiDy96jyGO2Zhnlmg=\", \"dataroot\": \"T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=\"}",
	}
	_, err := suite.msgServer.SubmitNonAvailableBatch(goCtx, &nonAvailableBatch)
	suite.ErrorContains(err, "span inside square size")

}

func (suite *RollappTestSuite) TestInvalidDAPathNonAvaliableSubmission() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappBatch(goCtx)
	nonAvailableBatch := types.MsgNonAvailableBatch{
		Creator:           carol,
		RollappId:         rollapId,
		SlIndex:           SLIndex,
		DAPath:            "",
		NonInclusionProof: "{\"rproofs\": \"CCAaIEPfyrlgwKe73bHoJPu/s3g/cqpCD+GSBhprXt6dQKvfIiCxvoDp0he9J928CKGi9YRyq6e/TO3VhMmtf31kZeA9XSIgVQgl6ThdDbM23Jb12Qpj2432o8n326zTSifjYdENdwMiIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiDWJEaMImR0qENxIHxcF3dTz4zl4JNLS4qU/bst0GdwsSIgvHPiMyjd1cVH4kYeTZrUUQQfxAeiDy96jyGO2Zhnlmg=\", \"dataroot\": \"T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=\"}",
	}
	_, err := suite.msgServer.SubmitNonAvailableBatch(goCtx, &nonAvailableBatch)
	suite.ErrorContains(err, "unable to decode da path")
}

func (suite *RollappTestSuite) TestInvalidProofNonAvaliableSubmission() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappBatch(goCtx)

	nonAvailableBatch := types.MsgNonAvailableBatch{
		Creator:           carol,
		RollappId:         rollapId,
		SlIndex:           SLIndex,
		DAPath:            "celestia.1.2.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef.T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
		NonInclusionProof: "{\"rproofs\": \"CCAaIJEj8f2IPQclrU2Idx3MUPvS6w+uYLsDI8Ra/V4FazqmIiDy+Px4UL5PvrKaDurYmnviREczHIR5IwJHqxI3LKKSDCIguFOX9794Pc38TRp4WRmaBekhBDGYEMP15NoGWnncib0iIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiD3uhkhupctQ422YZnvME78sHZn4BBoNff2m0bMQT81OyIg39Yb59vgAR5QHpQQowJsWcwcWc6UxcttOijHef0gcwo=\", \"dataroot\": \"T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=\"}",
	}
	_, err := suite.msgServer.SubmitNonAvailableBatch(goCtx, &nonAvailableBatch)
	suite.Require().ErrorContains(err, "unable to verify proof")
}

func (suite *RollappTestSuite) TestNonAvaliableSubmission() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappBatchOutofSquare(goCtx)

	nonAvailableBatch := types.MsgNonAvailableBatch{
		Creator:           carol,
		RollappId:         rollapId,
		SlIndex:           SLIndex,
		DAPath:            "celestia.1.1340.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef.T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
		NonInclusionProof: "{\"rproofs\": \"CCAaIEPfyrlgwKe73bHoJPu/s3g/cqpCD+GSBhprXt6dQKvfIiCxvoDp0he9J928CKGi9YRyq6e/TO3VhMmtf31kZeA9XSIgVQgl6ThdDbM23Jb12Qpj2432o8n326zTSifjYdENdwMiIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiDWJEaMImR0qENxIHxcF3dTz4zl4JNLS4qU/bst0GdwsSIgvHPiMyjd1cVH4kYeTZrUUQQfxAeiDy96jyGO2Zhnlmg=\", \"dataroot\": \"T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=\"}",
	}
	_, err := suite.msgServer.SubmitNonAvailableBatch(goCtx, &nonAvailableBatch)
	suite.Require().Nil(err)
}

func (suite *RollappTestSuite) TestInvalidDataRootNonAvaliableSubmission() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappBatchWrongDataRoot(goCtx)

	nonAvailableBatch := types.MsgNonAvailableBatch{
		Creator:           carol,
		RollappId:         rollapId,
		SlIndex:           SLIndex,
		DAPath:            "celestia.1.2.4.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e.e06c57a64b049d6463ef.T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
		NonInclusionProof: "{\"rproofs\": \"CCAaIEPfyrlgwKe73bHoJPu/s3g/cqpCD+GSBhprXt6dQKvfIiCxvoDp0he9J928CKGi9YRyq6e/TO3VhMmtf31kZeA9XSIgVQgl6ThdDbM23Jb12Qpj2432o8n326zTSifjYdENdwMiIK2G2DYjwb3UZvotBTsSD++LddcxR75oczW1aJfz4fiyIiDWJEaMImR0qENxIHxcF3dTz4zl4JNLS4qU/bst0GdwsSIgvHPiMyjd1cVH4kYeTZrUUQQfxAeiDy96jyGO2Zhnlmg=\", \"dataroot\": \"T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=\"}",
	}
	_, err := suite.msgServer.SubmitNonAvailableBatch(goCtx, &nonAvailableBatch)
	suite.Require().ErrorContains(err, "unable to verify proof")
}
