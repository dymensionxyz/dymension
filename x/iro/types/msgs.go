package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgCreatePlan{}
	_ sdk.Msg = &MsgBuy{}
	_ sdk.Msg = &MsgSell{}
	_ sdk.Msg = &MsgClaim{}
	_ sdk.Msg = &MsgUpdateParams{}
)

func (m *MsgCreatePlan) ValidateBasic() error {
	return nil
}

func (m *MsgCreatePlan) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{addr}
}

func (m *MsgBuy) ValidateBasic() error {
	// FIXME: Implement MsgBuy validation
	return nil
}

func (m *MsgBuy) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Buyer)
	return []sdk.AccAddress{addr}
}

func (m *MsgSell) ValidateBasic() error {
	// FIXME: Implement MsgSell validation
	return nil
}

func (m *MsgSell) GetSigners() []sdk.AccAddress {

	addr := sdk.MustAccAddressFromBech32(m.Seller)
	return []sdk.AccAddress{addr}
}

func (m *MsgClaim) ValidateBasic() error {

	return nil
}

func (m *MsgClaim) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Claimer)
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"authority '%s' must be a valid bech32 address: %s",
			m.Authority, err.Error(),
		)
	}

	err = m.NewParams.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr := sdk.MustAccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
