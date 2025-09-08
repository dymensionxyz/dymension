package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// https://cloud.google.com/confidential-computing/confidential-space/docs/reference/token-claims
// 'One or more nonces for the attestation token. The values are echoed from the token options sent in the custom token request. Each nonce must be between 8 and 88 bytes inclusive. A maximum of six nonces are allowed.'
func (n TEENonce) Hash() string {
	bz := []byte(n.RollappId)
	bzIx := make([]byte, 8)
	binary.BigEndian.PutUint64(bzIx, n.StateIndex)
	bz = append(bz, bzIx...)
	bz = append(bz, n.StateRoot...)
	hash := sha256.Sum256(bz)
	return hex.EncodeToString(hash[:])
}
