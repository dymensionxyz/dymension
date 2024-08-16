package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// FallbackAddress is used for reverse lookup using fallback mechanism.
// Fallback mechanism is used in reverse-lookup, to find all possible Dym-Name-Address from account address in bytes,
// compatible for coin-type-60 chains only (host-chain, RollApps)
type FallbackAddress []byte

// ValidateBasic performs basic validation for the FallbackAddress.
func (m FallbackAddress) ValidateBasic() error {
	if length := len(m); length != 20 && length != 32 {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "fallback address must be 20 or 32 bytes, got: %d", length)
	}

	return nil
}

// String returns the fallback-address in lowercase hex format.
func (m FallbackAddress) String() string {
	return strings.ToLower(dymnsutils.GetHexAddressFromBytes(m))
}
