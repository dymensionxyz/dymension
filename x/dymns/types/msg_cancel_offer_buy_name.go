package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgCancelOfferBuyName{}

func (m *MsgCancelOfferBuyName) ValidateBasic() error {
	if !dymnsutils.IsValidBuyNameOfferId(m.OfferId) {
		return ErrValidationFailed.Wrap("offer id is not a valid buy name offer id")
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return ErrValidationFailed.Wrap("buyer is not a valid bech32 account address")
	}

	return nil
}

func (m *MsgCancelOfferBuyName) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

func (m *MsgCancelOfferBuyName) Route() string {
	return RouterKey
}

func (m *MsgCancelOfferBuyName) Type() string {
	return TypeMsgCancelOfferBuyName
}

func (m *MsgCancelOfferBuyName) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
