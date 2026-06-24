package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (m *MsgSubmitAttestedAction) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Submitter); err != nil {
		return gerrc.ErrInvalidArgument.Wrap("submitter")
	}
	if m.AgentId == "" {
		return gerrc.ErrInvalidArgument.Wrap("agent id is required")
	}
	if m.Token == "" {
		return gerrc.ErrInvalidArgument.Wrap("token is required")
	}
	return nil
}
