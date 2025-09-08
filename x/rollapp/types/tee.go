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
	binary.BigEndian.PutUint64(bzIx, n.FinalizedHeight)
	bz = append(bz, bzIx...)
	bz = append(bz, n.FinalizedStateRoot...)

	bzIx = make([]byte, 8)
	binary.BigEndian.PutUint64(bzIx, n.CurrHeight)
	bz = append(bz, bzIx...)
	bz = append(bz, n.CurrStateRoot...)

	hash := sha256.Sum256(bz)
	return hex.EncodeToString(hash[:])
}

func (n TEENonce) Validate() error {
	if n.FinalizedHeight == 0 {
		return gerrc.ErrInvalidArgument.Wrap("finalized height is required")
	}
	if len(n.FinalizedStateRoot) != 32 {
		return gerrc.ErrInvalidArgument.Wrap("finalized state root is required")
	}
	if n.CurrHeight == 0 {
		return gerrc.ErrInvalidArgument.Wrap("current height is required")
	}
	if len(n.CurrStateRoot) != 32 {
		return gerrc.ErrInvalidArgument.Wrap("current state root is required")
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
