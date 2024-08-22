package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgAddApp = "add_app"

var _ sdk.Msg = &MsgAddApp{}

func NewMsgAddApp(creator, name, rollappId, description, image, url string, order int32) *MsgAddApp {
	return &MsgAddApp{
		Creator:     creator,
		Name:        name,
		RollappId:   rollappId,
		Description: description,
		Image:       image,
		Url:         url,
		Order:       order,
	}
}

func (msg *MsgAddApp) Route() string {
	return RouterKey
}

func (msg *MsgAddApp) Type() string {
	return TypeMsgAddApp
}

func (msg *MsgAddApp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAddApp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddApp) GetApp() App {
	return NewApp(
		msg.Name,
		msg.RollappId,
		msg.Description,
		msg.Image,
		msg.Url,
	)
}

func (msg *MsgAddApp) ValidateBasic() error {
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
