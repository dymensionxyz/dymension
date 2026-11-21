package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func NewMsgFundAgentEscrow(funder, agentID string, amount sdk.Coins) *MsgFundAgentEscrow {
	return &MsgFundAgentEscrow{
		Funder:  funder,
		AgentId: agentID,
		Amount:  amount,
	}
}

func (m *MsgFundAgentEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Funder); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "funder address")
	}
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	return validateEscrowAmount(m.Amount)
}

func NewMsgWithdrawAgentEscrow(owner, agentID string, amount sdk.Coins) *MsgWithdrawAgentEscrow {
	return &MsgWithdrawAgentEscrow{
		Owner:   owner,
		AgentId: agentID,
		Amount:  amount,
	}
}

func (m *MsgWithdrawAgentEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner address")
	}
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	return validateEscrowAmount(m.Amount)
}

func validateEscrowAmount(amount sdk.Coins) error {
	if err := amount.Validate(); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "amount")
	}
	if amount.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty amount")
	}
	return nil
}

func NewMsgUpdateAgentSpendPolicy(owner, agentID, spendDenom string, spendLimitPerWindow math.Int, spendWindowBlocks uint64) *MsgUpdateAgentSpendPolicy {
	return &MsgUpdateAgentSpendPolicy{
		Owner:               owner,
		AgentId:             agentID,
		SpendDenom:          spendDenom,
		SpendLimitPerWindow: spendLimitPerWindow,
		SpendWindowBlocks:   spendWindowBlocks,
	}
}

func (m *MsgUpdateAgentSpendPolicy) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner address")
	}
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	recipients := make(map[string]struct{}, len(m.RecipientAllowlist))
	for _, recipient := range m.RecipientAllowlist {
		address, err := sdk.AccAddressFromBech32(recipient)
		if err != nil {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "recipient allowlist address")
		}
		key := string(address)
		if _, duplicate := recipients[key]; duplicate {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "duplicate recipient allowlist address")
		}
		recipients[key] = struct{}{}
	}
	if err := ValidateSpendPolicy(m.SpendDenom, m.SpendLimitPerWindow, m.SpendWindowBlocks); err != nil {
		return err
	}
	return ValidateRecipientSpendLimit(m.SpendLimitPerWindow, m.SpendLimitPerRecipientPerWindow, len(m.RecipientAllowlist))
}

// ValidateRecipientSpendLimit checks the optional per-recipient window cap.
// It requires an allowlist because per-recipient state is only bounded by the
// allowlist size, and it may not exceed the aggregate cap, above which it
// would never bind. Call after ValidateSpendPolicy.
func ValidateRecipientSpendLimit(spendLimitPerWindow, perRecipient math.Int, allowlistLen int) error {
	if perRecipient.IsNil() || perRecipient.IsZero() {
		return nil
	}
	if perRecipient.IsNegative() {
		return errorsmod.Wrap(ErrInvalidSpendPolicy, "negative spend limit per recipient per window")
	}
	if allowlistLen == 0 {
		return errorsmod.Wrap(ErrInvalidSpendPolicy, "spend limit per recipient per window requires a recipient allowlist")
	}
	if spendLimitPerWindow.IsNil() || perRecipient.GT(spendLimitPerWindow) {
		return errorsmod.Wrap(ErrInvalidSpendPolicy, "spend limit per recipient per window exceeds spend limit per window")
	}
	return nil
}

// ValidateSpendPolicy checks the spend-policy invariant shared by
// MsgUpdateAgentSpendPolicy and genesis-imported agents: the budget fields are
// set iff the denom is, so SpendBucket can never divide by zero on an enabled
// agent.
func ValidateSpendPolicy(spendDenom string, spendLimitPerWindow math.Int, spendWindowBlocks uint64) error {
	if spendDenom == "" {
		// disabling spending: the budget fields must be unset
		if !spendLimitPerWindow.IsNil() && !spendLimitPerWindow.IsZero() {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend limit set without spend denom")
		}
		if spendWindowBlocks != 0 {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend window blocks set without spend denom")
		}
		return nil
	}
	if sdk.ValidateDenom(spendDenom) != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend denom")
	}
	if spendLimitPerWindow.IsNil() || !spendLimitPerWindow.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend limit per window must be positive")
	}
	if spendWindowBlocks == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "spend window blocks must be positive")
	}
	return nil
}
