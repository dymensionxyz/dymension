package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgRegisterAlias{}

// ValidateBasic performs basic validation for the MsgRegisterAlias.
func (m *MsgRegisterAlias) ValidateBasic() error {
	if len(m.Alias) > dymnsutils.MaxAliasLength {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"alias is too long, maximum %d characters", dymnsutils.MaxAliasLength,
		)
	}

	if !dymnsutils.IsValidAlias(m.Alias) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias is not a valid alias format")
	}

	if !dymnsutils.IsValidChainIdFormat(m.RollappId) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "RollApp ID is not a valid chain id format")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address: %s", m.Owner)
	}

	if m.ConfirmPayment.IsNil() || m.ConfirmPayment.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "confirm payment is not set")
	} else if err := m.ConfirmPayment.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid confirm payment: %v", err)
	}

	return nil
}
