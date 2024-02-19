package inclusion

type NonInclusionProof struct {
	RowProofs [][]byte
	DataRoot  []byte
}

func (nip *NonInclusionProof) VerifyNonInclusion(namespace []byte, span int, length int, dataRoot []byte) error {

	//TODO (srene): validate nmt root to data root using included merkle proofs
	return nil
}
