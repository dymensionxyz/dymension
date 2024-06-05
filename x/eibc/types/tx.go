package types

import (
	"encoding/hex"
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = sdk.Msg(&MsgFulfillOrder{})
var _ = sdk.Msg(&MsgUpdateDemandOrder{})

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
	err := validateCommon(m.OrderId, m.FulfillerAddress, m.MinFee)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (m *MsgFulfillOrder) GetFulfillerBech32Address() []byte {
	return sdk.MustAccAddressFromBech32(m.FulfillerAddress)
}

func (m *MsgUpdateDemandOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.RecipientAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (m *MsgUpdateDemandOrder) ValidateBasic() error {
	err := validateCommon(m.OrderId, m.RecipientAddress, m.NewFee)
	if err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	return nil
}

func (m *MsgUpdateDemandOrder) GetSubmitterAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(m.RecipientAddress)
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

func validateCommon(orderId, address, fee string) error {
	if !isValidOrderId(orderId) {
		return ErrInvalidOrderID
	}
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	feeInt, ok := sdk.NewIntFromString(fee)
	if !ok {
		return fmt.Errorf("parse fee: %s", fee)
	}

	if !feeInt.IsPositive() {
		return ErrNegativeFee
	}

	return nil
}
