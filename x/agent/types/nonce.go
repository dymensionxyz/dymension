package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// ActionNonce derives the per-action nonce that binds an attestation token to a
// specific (agent, payload, sequence). The byte layout mirrors
// x/rollapp/types.TEENonce.Hash: fixed-width big-endian integers and raw byte
// fields concatenated, then sha256'd. The agent echoes this nonce back as a
// token claim, which the rego policy checks.
//
// Replay protection is structural: actionSeq advances after each successful
// action, so a re-submitted (payload, token) re-derives a different nonce that
// no longer matches the token's claim, and the verifier rejects it.
func ActionNonce(agentID string, payload []byte, actionSeq uint64) string {
	payloadHash := sha256.Sum256(payload)

	buf := []byte(agentID)
	buf = append(buf, payloadHash[:]...)

	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, actionSeq)
	buf = append(buf, seq...)

	hash := sha256.Sum256(buf)
	return hex.EncodeToString(hash[:])
}
