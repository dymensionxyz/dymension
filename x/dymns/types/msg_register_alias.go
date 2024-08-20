package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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

// GetSigners returns the required signers for the MsgRegisterAlias.
func (m *MsgRegisterAlias) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgRegisterAlias.
func (m *MsgRegisterAlias) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgRegisterAlias.
func (m *MsgRegisterAlias) Type() string {
	return TypeMsgRegisterAlias
}

// GetSignBytes returns the raw bytes for the MsgRegisterAlias.
func (m *MsgRegisterAlias) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
