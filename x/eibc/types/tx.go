package types

import (
	"encoding/hex"
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ = sdk.Msg(&MsgFulfillOrder{})
	_ = sdk.Msg(&MsgUpdateDemandOrder{})
)

func NewMsgFulfillOrder(fulfillerAddress, orderId, minFee string) *MsgFulfillOrder {
	return &MsgFulfillOrder{
		FulfillerAddress: fulfillerAddress,
		OrderId:          orderId,
	}
}

func (msg *MsgFulfillOrder) Route() string {
	return RouterKey
}

func (msg *MsgFulfillOrder) Type() string {
	return sdk.MsgTypeURL(msg)
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
	err := validateCommon(m.OrderId, m.FulfillerAddress)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (m *MsgFulfillOrder) GetFulfillerBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(m.FulfillerAddress)
}

func NewMsgUpdateDemandOrder(orderId, ownerAddr, newFee string) *MsgUpdateDemandOrder {
	return &MsgUpdateDemandOrder{
		OrderId: orderId,
		Owner:   ownerAddr,
		NewFee:  newFee,
	}
}

func (m *MsgUpdateDemandOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (m *MsgUpdateDemandOrder) ValidateBasic() error {
	err := validateCommon(m.OrderId, m.Owner)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	feeInt, ok := sdk.NewIntFromString(m.NewFee)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("parse fee: %s", m.NewFee))
	}

	if feeInt.IsNegative() {
		return ErrNegativeFee
	}

	return nil
}

func (m *MsgUpdateDemandOrder) GetSubmitterAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.Owner)
}

func isValidOrderId(orderId string) bool {
	hashBytes, err := hex.DecodeString(orderId)
	if err != nil {
		// The string is not a valid hexadecimal string
		return false
	}
	// SHA-256 hashes are 32 bytes long
	return len(hashBytes) == 32
}

func validateCommon(orderId, address string) error {
	if !isValidOrderId(orderId) {
		return ErrInvalidOrderID
	}
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	return nil
}
