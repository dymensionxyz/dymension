package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgFulfillOrder = "update_state"

var _ = sdk.Msg(&MsgFulfillOrder{})

func NewMsgFullfillOrder(fullfillerAddress string, orderId string) *MsgFulfillOrder {
	return &MsgFulfillOrder{
		FullfillerAddress: fullfillerAddress,
		OrderId:           orderId,
	}
}

func (msg *MsgFulfillOrder) Route() string {
	return RouterKey
}

func (msg *MsgFulfillOrder) Type() string {
	return TypeMsgFulfillOrder
}

func (msg *MsgFulfillOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FullfillerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgFulfillOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (m *MsgFulfillOrder) ValidateBasic() error {
	if m.OrderId == "" {
		return ErrInvalidOrderID
	}
	_, err := sdk.AccAddressFromBech32(m.FullfillerAddress)
	if err != nil {
		return err
	}
	return nil
}

func (m *MsgFulfillOrder) Validate() error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (m *MsgFulfillOrder) GetFullfillerBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(m.FullfillerAddress)
}
