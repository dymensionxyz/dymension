package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgTransferDymNameOwnership{}

// ValidateBasic performs basic validation for the MsgTransferDymNameOwnership.
func (m *MsgTransferDymNameOwnership) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if _, err := sdk.AccAddressFromBech32(m.NewOwner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "new owner is not a valid bech32 account address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	if strings.EqualFold(m.NewOwner, m.Owner) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "new owner must be different from the current owner")
	}

	return nil
}

// GetSigners returns the required signers for the MsgTransferDymNameOwnership.
func (m *MsgTransferDymNameOwnership) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgTransferDymNameOwnership.
func (m *MsgTransferDymNameOwnership) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgTransferDymNameOwnership.
func (m *MsgTransferDymNameOwnership) Type() string {
	return TypeMsgTransferDymNameOwnership
}

// GetSignBytes returns the raw bytes for the MsgTransferDymNameOwnership.
func (m *MsgTransferDymNameOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
