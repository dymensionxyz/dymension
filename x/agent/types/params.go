package types

import (
	"fmt"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// DefaultMaxActionBytes caps the size of a single agent action payload.
const DefaultMaxActionBytes = 100_000

// DefaultPolicyRotationDelayBlocks is ~7 days at 6s blocks.
const DefaultPolicyRotationDelayBlocks = 100_800

func DefaultParams() Params {
	return Params{
		AgentRegistrationFee:      commontypes.DYMCoin,
		MaxActionBytes:            DefaultMaxActionBytes,
		PolicyRotationDelayBlocks: DefaultPolicyRotationDelayBlocks,
	}
}

func (p Params) Validate() error {
	if err := p.AgentRegistrationFee.Validate(); err != nil {
		return fmt.Errorf("agent registration fee: %w", err)
	}
	if p.MaxActionBytes == 0 {
		return fmt.Errorf("max action bytes must be positive")
	}
	if p.PolicyRotationDelayBlocks == 0 {
		return fmt.Errorf("policy rotation delay blocks must be positive")
	}
	return nil
}
