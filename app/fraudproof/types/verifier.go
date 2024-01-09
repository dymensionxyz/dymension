package types

type FraudProofVerifier interface {
	InitFromFraudProof(fraudProof *FraudProof) error
	VerifyFraudProof(fraudProof *FraudProof) error
}
