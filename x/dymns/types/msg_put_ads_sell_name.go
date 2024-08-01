package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgPutAdsSellName{}

// ValidateBasic performs basic validation for the MsgPutAdsSellName.
func (m *MsgPutAdsSellName) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return ErrValidationFailed.Wrap("name is not a valid dym name")
	}

	so := m.ToSellOrder()

	// put a dummy expire at to validate, as zero expire at is invalid,
	// and we don't have context of time at this point
	so.ExpireAt = 1

	if err := so.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("invalid order: %v", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return ErrValidationFailed.Wrap("owner is not a valid bech32 account address")
	}

	return nil
}

// ToSellOrder converts the MsgPutAdsSellName to a SellOrder.
func (m *MsgPutAdsSellName) ToSellOrder() SellOrder {
	so := SellOrder{
		Name:      m.Name,
		MinPrice:  m.MinPrice,
		SellPrice: m.SellPrice,
	}

	if !so.HasSetSellPrice() {
		so.SellPrice = nil
	}

	return so
}

// GetSigners returns the required signers for the MsgPutAdsSellName.
func (m *MsgPutAdsSellName) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgPutAdsSellName.
func (m *MsgPutAdsSellName) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgPutAdsSellName.
func (m *MsgPutAdsSellName) Type() string {
	return TypeMsgPutAdsSellName
}

// GetSignBytes returns the raw bytes for the MsgPutAdsSellName.
func (m *MsgPutAdsSellName) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
