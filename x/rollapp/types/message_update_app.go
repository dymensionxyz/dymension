package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateApp{}

func NewMsgUpdateApp(creator string, id uint64, name, rollappId, description, image, url string, order int32) *MsgUpdateApp {
	return &MsgUpdateApp{
		Id:          id,
		Creator:     creator,
		Name:        name,
		RollappId:   rollappId,
		Description: description,
		Image:       image,
		Url:         url,
		Order:       order,
	}
}

func (msg *MsgUpdateApp) GetApp() App {
	return NewApp(
		msg.Id,
		msg.Name,
		msg.RollappId,
		msg.Description,
		msg.Image,
		msg.Url,
		msg.Order,
	)
}

func (msg *MsgUpdateApp) SetOrder(o int32) {
	msg.Order = o
}

func (msg *MsgUpdateApp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errors.Join(ErrInvalidCreatorAddress, err)
	}

	if msg.Id == 0 {
		return errorsmod.Wrap(ErrInvalidAppID, "App id cannot be zero")
	}

	app := msg.GetApp()
	if err = app.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
