package inclusion

import (
	"crypto/sha256"
	"errors"

	"github.com/celestiaorg/nmt"
	"github.com/cometbft/cometbft/crypto/merkle"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/rollkit/celestia-openrpc/types/blob"
)

type InclusionProof struct {
	Namespace []byte
	Blob      []byte
	Nmtproofs [][]byte
	Nmtroots  [][]byte
	RowProofs [][]byte
	DataRoot  []byte
}

func (ip *InclusionProof) VerifyBlobInclusion() ([]byte, int, int, error) {

	var nmtProofs []*nmt.Proof
	for _, codedNMTProof := range ip.Nmtproofs {
		var unmarshalledProof nmt.Proof
		err := unmarshalledProof.UnmarshalJSON(codedNMTProof)
		if err != nil {
			return nil, 0, 0, err
		}
		nmtProofs = append(nmtProofs, &unmarshalledProof)
	}

	var b blob.Blob
	err := b.UnmarshalJSON(ip.Blob)
	if err != nil {
		return nil, 0, 0, err
	}

	shares, err := blob.SplitBlobs(b)
	if err != nil {
		return nil, 0, 0, err
	}
	index := 0

	for i, nmtProof := range nmtProofs {

		sharesNum := nmtProof.End() - nmtProof.Start()
		var leafs [][]byte

		for j := index; j < index+sharesNum; j++ {
			leaf := shares[j].ToBytes()
			leafs = append(leafs, leaf)
		}

		if !nmtProof.VerifyInclusion(sha256.New(), ip.Namespace, leafs, ip.Nmtroots[i]) {
			return nil, 0, 0, errors.New("blob not included")
		}
		index += sharesNum
	}

	var indexProof *merkle.Proof
	for i, rowProof := range ip.RowProofs {

		var proof cmtcrypto.Proof
		err := proof.Unmarshal(rowProof)
		if err != nil {
			return nil, 0, 0, err
		}
		rProof, err := merkle.ProofFromProto(&proof)
		if i == 0 {
			indexProof = rProof
		}
		if err != nil {
			return nil, 0, 0, err
		}
		err = rProof.Verify(ip.DataRoot, ip.Nmtroots[i])
		if err != nil {
			return nil, 0, 0, err
		}
	}

	return b.Commitment, nmtProofs[0].Start() + (int(indexProof.Total) / 2 * int(indexProof.Index)), len(shares), nil
}
