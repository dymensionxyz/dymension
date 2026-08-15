package types

import (
	"fmt"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// DefaultMaxActionBytes caps the size of a single agent action payload.
const DefaultMaxActionBytes = 100_000

// DefaultPolicyRotationDelayBlocks is ~7 days at 6s blocks.
const DefaultPolicyRotationDelayBlocks = 100_800

// DefaultFeedbackTagMaxBytes caps the byte length of a feedback dimension tag.
const DefaultFeedbackTagMaxBytes = 32

// DefaultValidationTagMaxBytes caps validation response tags.
const DefaultValidationTagMaxBytes = 32

// DefaultValidationUriMaxBytes caps validation request and response URIs.
const DefaultValidationUriMaxBytes = 512

// DefaultSpendRecipientAllowlistMax bounds per-agent recipient policy state.
const DefaultSpendRecipientAllowlistMax = 50

func DefaultParams() Params {
	return Params{
		AgentRegistrationFee:       commontypes.DYMCoin,
		MaxActionBytes:             DefaultMaxActionBytes,
		PolicyRotationDelayBlocks:  DefaultPolicyRotationDelayBlocks,
		FeedbackFee:                commontypes.DYMCoin,
		FeedbackTagMaxBytes:        DefaultFeedbackTagMaxBytes,
		ValidationRequestFee:       commontypes.DYMCoin,
		ValidationTagMaxBytes:      DefaultValidationTagMaxBytes,
		ValidationUriMaxBytes:      DefaultValidationUriMaxBytes,
		SpendRecipientAllowlistMax: DefaultSpendRecipientAllowlistMax,
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
	if err := p.FeedbackFee.Validate(); err != nil {
		return fmt.Errorf("feedback fee: %w", err)
	}
	if p.FeedbackTagMaxBytes == 0 {
		return fmt.Errorf("feedback tag max bytes must be positive")
	}
	if err := p.ValidationRequestFee.Validate(); err != nil {
		return fmt.Errorf("validation request fee: %w", err)
	}
	if p.ValidationTagMaxBytes == 0 {
		return fmt.Errorf("validation tag max bytes must be positive")
	}
	if p.ValidationUriMaxBytes == 0 {
		return fmt.Errorf("validation uri max bytes must be positive")
	}
	if p.SpendRecipientAllowlistMax == 0 {
		return fmt.Errorf("spend recipient allowlist max must be positive")
	}
	return nil
}
