package keeper

import (
	"bytes"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	dyminttypes "github.com/dymensionxyz/dymint/types"
	pb "github.com/dymensionxyz/dymint/types/pb/dymint"
	"github.com/gogo/protobuf/proto"
	"github.com/rollkit/celestia-openrpc/types/blob"
)

func (k *Keeper) VerifyFraudProof(ctx sdk.Context, rollappID string, fp *fraudtypes.FraudProof, ip *inclusion.InclusionProof) error {
	err := k.ValidateFraudProof(ctx, rollappID, fp, ip)
	if err != nil {
		return err
	}

	err = k.fraudProofVerifier.InitFromFraudProof(fp)
	if err != nil {
		return err
	}
	err = k.fraudProofVerifier.VerifyFraudProof(fp)
	if err != nil {
		return err
	}

	return nil
}

// validate fraud proof preState Hash against the state update posted on the hub
func (k *Keeper) ValidateFraudProof(ctx sdk.Context, rollappID string, fp *fraudtypes.FraudProof, ip *inclusion.InclusionProof) error {
	//validate the fp struct and witnesses
	_, err := fp.ValidateBasic()
	if err != nil {
		return err
	}

	//validate the fraudproof against the commited state
	blockHeight := fp.BlockHeight + 1
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, uint64(blockHeight))
	if err != nil {
		return err
	}
	idx := blockHeight - int64(stateInfo.StartHeight)
	blockDescriptor := stateInfo.BDs.BD[idx]

	if blockDescriptor.IntermediateStatesRoots == nil {
		return types.ErrMissingIntermediateStatesRoots
	}

	found := false
	for idx, isr := range blockDescriptor.IntermediateStatesRoots {
		//skip the last ISR
		if idx == len(blockDescriptor.IntermediateStatesRoots)-1 {
			break
		}
		if bytes.Equal(isr, fp.PreStateAppHash) {
			//fmt.Println("found", idx)
			found = true
			break
		}
	}

	if !found {
		return types.ErrInvalidPreStateAppHash
	}

	if bytes.Equal(blockDescriptor.IntermediateStatesRoots[idx+1], fp.ExpectedValidAppHash) {
		return types.ErrInvalidExpectedAppHash
	}

	//blob inclusion validation
	_, _, _, err = ip.VerifyBlobInclusion()
	if err != nil {
		return err
	}
	//TODO(srene): dataroot and commitment validation against sequencer's data

	found = false
	blocks, err := getBlobBlocks(ip.Blob)
	if err != nil {
		return err
	}

	//locate tx in blob
	//isrs are not included in the blob by now, so we locate by isr index (to be checked)
	for _, block := range blocks {
		//fmt.Println(len(block.Data.Txs), idx, len(blockDescriptor.IntermediateStatesRoots))

		if int64(block.Header.Height) == blockHeight {
			idxCmp := idx - 1
			if idxCmp < 0 {
				idxCmp = 0
			}
			if bytes.Equal(block.Data.Txs[idxCmp], fp.FraudulentDeliverTx.Tx) {
				found = true
				break
			}
		}
	}
	if !found {
		return types.ErrBlobInclusionNotValidated
	}
	// TODO: Validate the fraudulent state transition is contained in the block header

	return nil
}

func getBlobBlocks(b []byte) ([]*dyminttypes.Block, error) {

	var decodedBlob blob.Blob
	err := decodedBlob.UnmarshalJSON(b)
	if err != nil {
		return nil, err
	}

	var batch pb.Batch
	err = proto.Unmarshal(decodedBlob.Data, &batch)
	if err != nil {
		return nil, err
	}
	parsedBatch := new(dyminttypes.Batch)
	err = parsedBatch.FromProto(&batch)
	if err != nil {
		return nil, err
	}

	return parsedBatch.Blocks, nil

}
