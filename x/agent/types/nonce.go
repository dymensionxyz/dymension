package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"cosmossdk.io/math"
)

// Nonce domain tags. Actions and transfers share one action_seq counter, so
// without domain separation a token minted for a pending transfer could be
// replayed by any observer as a plain action at the same seq — advancing the
// counter and killing the enclave-authorized payment. The tag makes the two
// nonce spaces disjoint: a token only verifies for the message type it was
// minted for.
const (
	nonceDomainAction   byte = 0x01
	nonceDomainTransfer byte = 0x02
)

// ActionNonce derives the per-action nonce that binds an attestation token to
// a specific (agent, payload, sequence). The byte layout mirrors
// x/rollapp/types.TEENonce.Hash: raw byte fields and fixed-width big-endian
// integers concatenated, then sha256'd. The agent echoes this nonce back as a
// token claim, which the rego policy checks.
//
// Replay protection is structural: actionSeq advances after each successful
// action, so a re-submitted (payload, token) re-derives a different nonce that
// no longer matches the token's claim, and the verifier rejects it.
func ActionNonce(agentID string, payload []byte, actionSeq uint64) string {
	return attestNonce(nonceDomainAction, agentID, payload, actionSeq)
}

// TransferNonce is ActionNonce in the transfer domain: same layout, disjoint
// nonce space (see the domain tag comment above).
func TransferNonce(agentID string, payload []byte, actionSeq uint64) string {
	return attestNonce(nonceDomainTransfer, agentID, payload, actionSeq)
}

// attestNonce hashes agentID ‖ sha256(payload) ‖ seq ‖ domain. The suffix after
// agentID is fixed-width (32+8+1), so the encoding is unambiguous even though
// agentID is variable-length.
func attestNonce(domain byte, agentID string, payload []byte, actionSeq uint64) string {
	payloadHash := sha256.Sum256(payload)

	buf := []byte(agentID)
	buf = append(buf, payloadHash[:]...)

	seq := make([]byte, 8)
	binary.BigEndian.PutUint64(seq, actionSeq)
	buf = append(buf, seq...)
	buf = append(buf, domain)

	hash := sha256.Sum256(buf)
	return hex.EncodeToString(hash[:])
}

// AttestedTransferBytes is the canonical payload an attested transfer's nonce
// commits to, so the enclave attests to this exact recipient, denom and
// amount rather than an opaque blob. Fields are NUL-separated: 0x00 cannot
// appear in a bech32 address, a denom or a decimal string, so the encoding is
// unambiguous and deterministic.
func AttestedTransferBytes(recipient, spendDenom string, amount math.Int, memo []byte) []byte {
	buf := make([]byte, 0, len(recipient)+len(spendDenom)+len(memo)+24)
	buf = append(buf, recipient...)
	buf = append(buf, 0x00)
	buf = append(buf, spendDenom...)
	buf = append(buf, 0x00)
	buf = append(buf, amount.String()...)
	buf = append(buf, 0x00)
	buf = append(buf, memo...)
	return buf
}
