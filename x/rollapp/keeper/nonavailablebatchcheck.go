package keeper

import (
	"crypto/sha256"

	"github.com/celestiaorg/nmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/rollkit/celestia-openrpc/types/blob"
)

func (k *Keeper) VerifyNonAvailableBatch(ctx sdk.Context, msg *types.MsgNonAvailableBatch) error {

	stateInfo, found := k.GetStateInfo(ctx, msg.GetRollappId(), msg.GetSlIndex())
	if !found {
		return nil
	}
	DAPath := stateInfo.GetDAPath()
	DAMetaData, err := types.NewDAMetaData(DAPath)
	if err != nil {
		return nil
	}
	//var namespace []byte
	blob, _, err := k.blobsAndCommitments(DAMetaData.GetNameSpace(), msg.GetBlob())
	if err != nil {
		return err
	}
	switch msg.GetCase() {
	case types.NonAvaliableCase_wrongcommitment:
		err := k.verifyBlobInclusion(ctx, DAMetaData.GetNameSpace(), blob, msg.GetNmtproofs(), msg.GetNmtroots(), msg.GetRproofs(), msg.GetDataroot())
		if err != nil {
			return err
		}
		return types.ErrWrongCommitment

	case types.NonAvaliableCase_notavailable:
		err := k.verifyBlobNonInclusion(ctx, DAMetaData.GetNameSpace(), msg.GetRproofs(), msg.GetDataroot())
		if err != nil {
			return err
		}
		return types.ErrBatchNotAvailable
	case types.NonAvaliableCase_invaliddata:
		err := k.validateIncludedData(ctx, blob)
		if err != nil {
			return err
		}
		return types.ErrInvalidBlobData
	default:
		return types.ErrSubmitNonAvailableBatchWrongCase
	}
}
func (k *Keeper) verifyBlobNonInclusion(ctx sdk.Context, namespace []byte, rProofs [][]byte, dataRoot []byte) error {
	//TODO (srene): Implement non-inclusion proof validation
	return types.ErrUnableToVerifyProof
}

func (k *Keeper) validateIncludedData(ctx sdk.Context, b *blob.Blob) error {
	//TODO (srene): Implement check invalid data blob
	return types.ErrUnableToVerifyProof
}

func (k *Keeper) verifyBlobInclusion(ctx sdk.Context, namespace []byte, b *blob.Blob, nProofs [][]byte, rowRoots [][]byte, rProofs [][]byte, dataRoot []byte) error {
	var nmtProofs []*nmt.Proof
	for _, codedNMTProof := range nProofs {
		var unmarshalledProof nmt.Proof
		err := unmarshalledProof.UnmarshalJSON(codedNMTProof)
		if err != nil {
			return types.ErrUnableToVerifyProof
		}
		nmtProofs = append(nmtProofs, &unmarshalledProof)
	}
	shares, err := blob.SplitBlobs(*b)
	if err != nil {
		return types.ErrUnableToVerifyProof
	}
	index := 0

	for i, nmtProof := range nmtProofs {
		sharesNum := nmtProof.End() - nmtProof.Start()
		var leafs [][]byte

		for j := index; j < index+sharesNum; j++ {
			leaf := shares[j].ToBytes()
			leafs = append(leafs, leaf)
		}
		if !nmtProof.VerifyInclusion(sha256.New(), namespace, leafs, rowRoots[i]) {
			return types.ErrUnableToVerifyProof
		}

		index += sharesNum
	}
	//TODO (srene): validate nmt root to data root using included merkle proofs
	return nil
}

func (k *Keeper) blobsAndCommitments(namespace []byte, daBlob []byte) (*blob.Blob, []byte, error) {
	b, err := blob.NewBlobV0(namespace, daBlob)
	if err != nil {
		return nil, nil, err
	}

	commitment, err := blob.CreateCommitment(b)
	if err != nil {
		return nil, nil, err
	}

	return b, commitment, nil
}
