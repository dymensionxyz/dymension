package keeper_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rollkit/celestia-openrpc/types/blob"
	openrpcns "github.com/rollkit/celestia-openrpc/types/namespace"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (suite *RollappTestSuite) createRollappValidBatch(goCtx context.Context, commitment []byte, namespace []byte, dataRoot []byte) (string, uint64) {

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
		DAPath:      "celestia.1.1.22." + string(commitment) + "." + string(namespace) + "." + string(dataRoot),
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	stateIndexInfo, _ := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	return stateIndexInfo.GetRollappId(), stateIndexInfo.GetIndex()
}

func (suite *RollappTestSuite) createRollappWrongCommitmentBatch(goCtx context.Context, commitment []byte, namespace []byte, dataRoot []byte) (string, uint64) {

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
		DAPath:      "celestia.1.1.22.aa5b76fe9c42a5aff1fcfe1cc5088b3941cb1cc854c22ce6c0c0fb98a5461f8e." + string(namespace) + "." + string(dataRoot),
		Version:     3,
		BDs:         types.BlockDescriptors{BD: []types.BlockDescriptor{{Height: 1}, {Height: 1}, {Height: 2}}},
	}

	_, err := suite.msgServer.UpdateState(goCtx, &updateState)
	suite.Require().Nil(err)

	stateIndexInfo, _ := suite.app.RollappKeeper.GetLatestStateInfoIndex(suite.ctx, rollapp.GetRollappId())
	return stateIndexInfo.GetRollappId(), stateIndexInfo.GetIndex()
}

func (suite *RollappTestSuite) TestValidSubmission() {

	file, err := os.Open("../../../app/dainclusionproofs/blob_inclusion_proof.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	proof := types.BlobInclusionProof{}
	err = jsonDecoder.Decode(&proof)
	suite.Require().Nil(err)

	s, err := os.ReadFile("../../../app/dainclusionproofs/blob_inclusion_proof.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	nameidstr := "e06c57a64b049d6463ef"
	namespaceBytes, err := hex.DecodeString(nameidstr)
	suite.Require().Nil(err)
	ns, err := openrpcns.New(openrpcns.NamespaceVersionZero, append(openrpcns.NamespaceVersionZeroPrefix, namespaceBytes...))
	suite.Require().Nil(err)

	var b blob.Blob
	err = b.UnmarshalJSON(proof.Blob)
	suite.Require().Nil(err)

	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappValidBatch(goCtx, b.Commitment, ns.Bytes(), proof.Dataroot)
	wrongCommitmentBatch := types.MsgWrongCommitmentBatch{
		Creator:        carol,
		RollappId:      rollapId,
		SlIndex:        SLIndex,
		DAPath:         "celestia.1.1.22." + string(b.Commitment) + "." + string(ns.Bytes()) + ".T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
		InclusionProof: string(s),
	}
	_, err = suite.msgServer.SubmitWrongCommitmentBatch(goCtx, &wrongCommitmentBatch)
	suite.Require().ErrorIs(err, sdkerrors.ErrInvalidRequest)

}

func (suite *RollappTestSuite) TestWrongCommitmentSubmission() {
	file, err := os.Open("../../../app/dainclusionproofs/blob_inclusion_proof.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	proof := types.BlobInclusionProof{}
	err = jsonDecoder.Decode(&proof)
	suite.Require().Nil(err)

	s, err := os.ReadFile("../../../app/dainclusionproofs/blob_inclusion_proof.json") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	nameidstr := "e06c57a64b049d6463ef"
	namespaceBytes, err := hex.DecodeString(nameidstr)
	suite.Require().Nil(err)
	ns, err := openrpcns.New(openrpcns.NamespaceVersionZero, append(openrpcns.NamespaceVersionZeroPrefix, namespaceBytes...))
	suite.Require().Nil(err)

	var b blob.Blob
	err = b.UnmarshalJSON(proof.Blob)
	suite.Require().Nil(err)

	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	rollapId, SLIndex := suite.createRollappWrongCommitmentBatch(goCtx, b.Commitment, ns.Bytes(), proof.Dataroot)
	wrongCommitmentBatch := types.MsgWrongCommitmentBatch{
		Creator:        carol,
		RollappId:      rollapId,
		SlIndex:        SLIndex,
		DAPath:         "celestia.1.1.22." + string(b.Commitment) + "." + string(ns.Bytes()) + ".T1SVEdrCgznblHlHsPgtEV6Ui1F7liW+Kfut/aUBjPo=",
		InclusionProof: string(s),
	}
	_, err = suite.msgServer.SubmitWrongCommitmentBatch(goCtx, &wrongCommitmentBatch)
	suite.Require().Nil(err)
}
