package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func NewMsgUpdateAgentPolicy(owner, agentID string, policy tee.Policy) *MsgUpdateAgentPolicy {
	return &MsgUpdateAgentPolicy{
		Owner:     owner,
		AgentId:   agentID,
		NewPolicy: policy,
	}
}

func (m *MsgUpdateAgentPolicy) ValidateBasic() error {
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner address")
	}
	return validatePolicy(m.NewPolicy)
}
