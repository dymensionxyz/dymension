package types

type FraudProofVerifierInterface interface {
	InitFromFraudProof(fraudProof *FraudProof)
	VerifyFraudProof(fraudProof *FraudProof) bool
}
