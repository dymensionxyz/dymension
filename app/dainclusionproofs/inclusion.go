package inclusion

import (
	"crypto/sha256"
	"errors"

	"github.com/celestiaorg/nmt"
	"github.com/rollkit/celestia-openrpc/types/blob"
)

type InclusionProof struct {
	Blob      []byte
	Nmtproofs [][]byte
	Nmtroots  [][]byte
	RowProofs [][]byte
	DataRoot  []byte
}

func (ip *InclusionProof) VerifyBlobInclusion(namespace []byte, dataRoot []byte) error {
	var nmtProofs []*nmt.Proof
	for _, codedNMTProof := range ip.Nmtproofs {
		var unmarshalledProof nmt.Proof
		err := unmarshalledProof.UnmarshalJSON(codedNMTProof)
		if err != nil {
			return err
		}
		nmtProofs = append(nmtProofs, &unmarshalledProof)
	}

	b, _, err := ip.blobsAndCommitments(namespace, ip.Blob)
	if err != nil {
		return err
	}

	shares, err := blob.SplitBlobs(*b)
	if err != nil {
		return err
	}
	index := 0

	for i, nmtProof := range nmtProofs {
		sharesNum := nmtProof.End() - nmtProof.Start()
		var leafs [][]byte

		for j := index; j < index+sharesNum; j++ {
			leaf := shares[j].ToBytes()
			leafs = append(leafs, leaf)
		}
		if !nmtProof.VerifyInclusion(sha256.New(), namespace, leafs, ip.Nmtroots[i]) {
			return errors.New("blob not included")
		}

		index += sharesNum
	}
	//TODO (srene): validate nmt root to data root using included merkle proofs
	return nil
}

func (ip *InclusionProof) blobsAndCommitments(namespace []byte, daBlob []byte) (*blob.Blob, []byte, error) {
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
