package types

import (
	"fmt"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// DefaultMaxActionBytes caps the size of a single agent action payload.
const DefaultMaxActionBytes = 100_000

func DefaultParams() Params {
	return Params{
		AgentRegistrationFee: commontypes.DYMCoin,
		MaxActionBytes:       DefaultMaxActionBytes,
	}
}

func (p Params) Validate() error {
	if err := p.AgentRegistrationFee.Validate(); err != nil {
		return fmt.Errorf("agent registration fee: %w", err)
	}
	if p.MaxActionBytes == 0 {
		return fmt.Errorf("max action bytes must be positive")
	}
	return nil
}
