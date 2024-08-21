package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgCreateApp = "create_app"

var _ sdk.Msg = &MsgCreateApp{}

func NewMsgCreateApp(creator, name, rollappId, description, image, url string, order int32) *MsgCreateApp {
	return &MsgCreateApp{
		Creator:     creator,
		Name:        name,
		RollappId:   rollappId,
		Description: description,
		Image:       image,
		Url:         url,
		Order:       order,
	}
}

func (msg *MsgCreateApp) Route() string {
	return RouterKey
}

func (msg *MsgCreateApp) Type() string {
	return TypeMsgCreateApp
}

func (msg *MsgCreateApp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateApp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateApp) GetApp() App {
	return NewApp(
		msg.Name,
		msg.RollappId,
		msg.Description,
		msg.Image,
		msg.Url,
	)
}

func (msg *MsgCreateApp) ValidateBasic() error {
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
