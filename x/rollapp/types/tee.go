package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// https://cloud.google.com/confidential-computing/confidential-space/docs/reference/token-claims
// 'One or more nonces for the attestation token. The values are echoed from the token options sent in the custom token request. Each nonce must be between 8 and 88 bytes inclusive. A maximum of six nonces are allowed.'
func (n TEENonce) Hash() string {
	bz := []byte(n.RollappId)
	bzIx := make([]byte, 8)
	binary.BigEndian.PutUint64(bzIx, n.Height)
	bz = append(bz, bzIx...)
	bz = append(bz, n.StateRoot...)
	hash := sha256.Sum256(bz)
	return hex.EncodeToString(hash[:])
}

func (n TEENonce) Validate() error {
	if n.Height == 0 {
		return gerrc.ErrInvalidArgument.Wrap("state index is required")
	}
	if len(n.StateRoot) != 32 {
		return gerrc.ErrInvalidArgument.Wrap("state root is required")
	}
	if n.RollappId == "" {
		return gerrc.ErrInvalidArgument.Wrap("rollapp id is required")
	}
	return nil
}

func (m *MsgFastFinalizeWithTEE) ValidateBasic() error {
	if err := m.Nonce.Validate(); err != nil {
		return errorsmod.Wrap(err, "nonce validation")
	}
	if m.AttestationToken == "" {
		return gerrc.ErrInvalidArgument.Wrap("attestation token is required")
	}
	return nil
}
