package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgUpdateDetails{}

func (m *MsgUpdateDetails) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return ErrValidationFailed.Wrap("name is not a valid dym name")
	}

	if m.Contact != DoNotModifyDesc {
		if len(m.Contact) > MaxDymNameContactLength {
			return ErrValidationFailed.Wrapf("contact is too long; max length: %d", MaxDymNameContactLength)
		}
	}

	if _, err := sdk.AccAddressFromBech32(m.Controller); err != nil {
		return ErrValidationFailed.Wrap("controller is not a valid bech32 account address")
	}

	if m.Contact == DoNotModifyDesc && !m.ClearConfigs {
		return ErrValidationFailed.Wrap("message neither clears configs nor updates contact information")
	}

	return nil
}

func (m *MsgUpdateDetails) GetSigners() []sdk.AccAddress {
	controller, err := sdk.AccAddressFromBech32(m.Controller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{controller}
}

func (m *MsgUpdateDetails) Route() string {
	return RouterKey
}

func (m *MsgUpdateDetails) Type() string {
	return TypeMsgUpdateDetails
}

func (m *MsgUpdateDetails) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
