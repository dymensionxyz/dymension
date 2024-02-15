package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgNonAvailableBatch = "submit_nonavailable"

var _ sdk.Msg = &MsgNonAvailableBatch{}

func NewMsgSubmitNonAvailableBatch(creator string, rollappID string, fraudproof string) *MsgNonAvailableBatch {
	return &MsgNonAvailableBatch{
		Creator:   creator,
		RollappId: rollappID,
	}
}

func (msg *MsgNonAvailableBatch) Route() string {
	return RouterKey
}

func (msg *MsgNonAvailableBatch) Type() string {
	return TypeMsgNonAvailableBatch
}

func (msg *MsgNonAvailableBatch) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgNonAvailableBatch) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgNonAvailableBatch) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
