package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func NewMsgRegisterAgent(owner, agentID string, policy tee.Policy) *MsgRegisterAgent {
	return &MsgRegisterAgent{
		Owner:   owner,
		AgentId: agentID,
		Policy:  policy,
	}
}

func (m *MsgRegisterAgent) ValidateBasic() error {
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner address")
	}
	return validatePolicy(m.Policy)
}

// validatePolicy checks the embedded TEE policy is well-formed: the root cert
// PEM parses and the rego query/structure are present.
func validatePolicy(p tee.Policy) error {
	if _, err := p.PemCert(); err != nil {
		return errorsmod.Wrap(ErrInvalidPolicy, "gcp root cert pem")
	}
	if p.PolicyQuery == "" {
		return errorsmod.Wrap(ErrInvalidPolicy, "empty policy query")
	}
	if p.PolicyStructure == "" {
		return errorsmod.Wrap(ErrInvalidPolicy, "empty policy structure")
	}
	return nil
}
