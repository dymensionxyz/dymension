package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// This legth is because we use sha256 to hash the order id
	maxLengthOfOrderID = 64
)

const TypeMsgFulfillOrder = "fulfill_order"

var _ = sdk.Msg(&MsgFulfillOrder{})

func NewMsgFulfillOrder(fulfillerAddress string, orderId string) *MsgFulfillOrder {
	return &MsgFulfillOrder{
		FulfillerAddress: fulfillerAddress,
		OrderId:          orderId,
	}
}

func (msg *MsgFulfillOrder) Route() string {
	return RouterKey
}

func (msg *MsgFulfillOrder) Type() string {
	return TypeMsgFulfillOrder
}

func (msg *MsgFulfillOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.FulfillerAddress)
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
	if m.OrderId == "" || len(m.OrderId) > maxLengthOfOrderID {
		return ErrInvalidOrderID
	}
	_, err := sdk.AccAddressFromBech32(m.FulfillerAddress)
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

func (m *MsgFulfillOrder) GetFulfillerBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(m.FulfillerAddress)
}
