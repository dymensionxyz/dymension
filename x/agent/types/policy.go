package types

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

// PolicyFingerprint is the lowercase hex SHA-256 of the deterministic proto
// marshaling of the policy. Two agents pinning a byte-identical policy (same
// root cert, rego, and expected measurement) share a fingerprint; that is
// exactly the set we revoke together when an image is found vulnerable.
// tee.Policy contains only string fields, so Marshal() is stable across
// validators — no map iteration, no floats.
func PolicyFingerprint(p tee.Policy) (string, error) {
	b, err := p.Marshal()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

// ValidateFingerprint checks the fingerprint is 64-char lowercase hex, the
// canonical form produced by PolicyFingerprint.
func ValidateFingerprint(fp string) error {
	if len(fp) != 64 {
		return gerrc.ErrInvalidArgument.Wrapf("fingerprint must be 64 hex chars, got %d", len(fp))
	}
	for _, c := range fp {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return gerrc.ErrInvalidArgument.Wrap("fingerprint must be lowercase hex")
		}
	}
	return nil
}
