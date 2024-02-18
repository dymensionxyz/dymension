package keeper

import (
	"crypto/sha256"

	"github.com/celestiaorg/nmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/rollkit/celestia-openrpc/types/blob"
)

func (k *Keeper) VerifyWrongCommitmentBatch(ctx sdk.Context, msg *types.MsgWrongCommitmentBatch, inclusionProof *types.InclusionProof) error {

	/*stateInfo, found := k.GetStateInfo(ctx, msg.GetRollappId(), msg.GetSlIndex())
	if !found {
		return nil
	}
	DAPath := stateInfo.GetDAPath()
	DAMetaDataSequencer, err := types.NewDAMetaData(DAPath)
	if err != nil {
		return nil
	}*/
	DaMetaDataSubmitted, err := types.NewDAMetaData(msg.GetDAPath())
	if err != nil {
		return nil
	}
	//var namespace []byte
	b, _, err := k.blobsAndCommitments(DaMetaDataSubmitted.GetNameSpace(), inclusionProof.Blob)
	if err != nil {
		return err
	}
	err = k.verifyBlobInclusion(ctx, DaMetaDataSubmitted.GetNameSpace(), b, inclusionProof.Nmtproofs, inclusionProof.Nmtroots, inclusionProof.RowProofs, inclusionProof.DataRoot)
	if err != nil {
		return err
	}

	return nil
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
