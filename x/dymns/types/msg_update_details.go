package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgUpdateDetails{}

// ValidateBasic performs basic validation for the MsgUpdateDetails.
func (m *MsgUpdateDetails) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if m.Contact != DoNotModifyDesc {
		if len(m.Contact) > MaxDymNameContactLength {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "contact is too long; max length: %d", MaxDymNameContactLength)
		}
	}

	if _, err := sdk.AccAddressFromBech32(m.Controller); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is not a valid bech32 account address")
	}

	if m.Contact == DoNotModifyDesc && !m.ClearConfigs {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "message neither clears configs nor updates contact information")
	}

	return nil
}

// GetSigners returns the required signers for the MsgUpdateDetails.
func (m *MsgUpdateDetails) GetSigners() []sdk.AccAddress {
	controller, err := sdk.AccAddressFromBech32(m.Controller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{controller}
}

// Route returns the message router key for the MsgUpdateDetails.
func (m *MsgUpdateDetails) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgUpdateDetails.
func (m *MsgUpdateDetails) Type() string {
	return TypeMsgUpdateDetails
}

// GetSignBytes returns the raw bytes for the MsgUpdateDetails.
func (m *MsgUpdateDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
