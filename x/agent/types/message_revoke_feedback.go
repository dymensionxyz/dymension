package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func NewMsgRevokeFeedback(client, agentID string) *MsgRevokeFeedback {
	return &MsgRevokeFeedback{
		Client:  client,
		AgentId: agentID,
	}
}

func (m *MsgRevokeFeedback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Client); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "client address")
	}
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	return nil
}
