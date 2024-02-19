package inclusion

import (
	"bytes"
	"errors"

	"github.com/cometbft/cometbft/crypto/merkle"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

type NonInclusionProof struct {
	RowProof []byte
}

func (nip *NonInclusionProof) VerifyNonInclusion(namespace []byte, span int, length int, dataRoot []byte) error {

	var proof cmtcrypto.Proof
	err := proof.Unmarshal(nip.RowProof)
	if err != nil {
		return err
	}
	rProof, err := merkle.ProofFromProto(&proof)
	if err != nil {
		return err
	}
	computedHash := rProof.ComputeRootHash()
	if err != nil {
		return err
	}
	if !bytes.Equal(computedHash, dataRoot) {
		return errors.New("data root not matching")
	}
	if span > int(rProof.Total) {
		return errors.New("span out of square size")
	}
	return nil
}
