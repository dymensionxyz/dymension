package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateApp = "update_app"

var _ sdk.Msg = &MsgUpdateApp{}

func NewMsgUpdateApp(creator, name, rollappId, description, image, url string, order int32) *MsgUpdateApp {
	return &MsgUpdateApp{
		Creator:     creator,
		Name:        name,
		RollappId:   rollappId,
		Description: description,
		Image:       image,
		Url:         url,
		Order:       order,
	}
}

func (msg *MsgUpdateApp) Route() string {
	return RouterKey
}

func (msg *MsgUpdateApp) Type() string {
	return TypeMsgUpdateApp
}

func (msg *MsgUpdateApp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateApp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateApp) GetApp() App {
	return NewApp(
		msg.Name,
		msg.RollappId,
		msg.Description,
		msg.Image,
		msg.Url,
		msg.Order,
	)
}

func (msg *MsgUpdateApp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCreatorAddress, err.Error())
	}

	app := msg.GetApp()
	if err = app.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
