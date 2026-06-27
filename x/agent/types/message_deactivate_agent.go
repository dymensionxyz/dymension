package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func NewMsgDeactivateAgent(owner, agentID string) *MsgDeactivateAgent {
	return &MsgDeactivateAgent{
		Owner:   owner,
		AgentId: agentID,
	}
}

func (m *MsgDeactivateAgent) ValidateBasic() error {
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner address")
	}
	return nil
}
