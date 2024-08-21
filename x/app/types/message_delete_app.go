package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgDeleteApp = "delete_app"

var _ sdk.Msg = &MsgDeleteApp{}

func NewMsgDeleteApp(creator, name, rollappId string) *MsgDeleteApp {
	return &MsgDeleteApp{
		Creator:   creator,
		Name:      name,
		RollappId: rollappId,
	}
}

func (msg *MsgDeleteApp) Route() string {
	return RouterKey
}

func (msg *MsgDeleteApp) Type() string {
	return TypeMsgDeleteApp
}

func (msg *MsgDeleteApp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeleteApp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteApp) GetApp() App {
	return NewApp(
		msg.Name,
		msg.RollappId,
		"",
		"",
		"",
	)
}

func (msg *MsgDeleteApp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCreatorAddress, err.Error())
	}

	if len(msg.Name) == 0 {
		return errorsmod.Wrap(ErrInvalidName, "Name cannot be empty")
	}

	if len(msg.RollappId) == 0 {
		return errorsmod.Wrap(ErrInvalidRollappId, "RollappId cannot be empty")
	}

	return nil
}
