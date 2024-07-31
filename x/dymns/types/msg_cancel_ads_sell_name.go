package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgCancelAdsSellName{}

func (m *MsgCancelAdsSellName) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return ErrValidationFailed.Wrap("name is not a valid dym name")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return ErrValidationFailed.Wrap("owner is not a valid bech32 account address")
	}

	return nil
}

func (m *MsgCancelAdsSellName) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

func (m *MsgCancelAdsSellName) Route() string {
	return RouterKey
}

func (m *MsgCancelAdsSellName) Type() string {
	return TypeMsgCancelAdsSellName
}

func (m *MsgCancelAdsSellName) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
