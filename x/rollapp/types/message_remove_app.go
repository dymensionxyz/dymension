package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgRemoveApp = "remove_app"

var _ sdk.Msg = &MsgRemoveApp{}

func NewMsgRemoveApp(creator, name, rollappId string) *MsgRemoveApp {
	return &MsgRemoveApp{
		Creator:   creator,
		Name:      name,
		RollappId: rollappId,
	}
}

func (msg *MsgRemoveApp) Route() string {
	return RouterKey
}

func (msg *MsgRemoveApp) Type() string {
	return TypeMsgRemoveApp
}

func (msg *MsgRemoveApp) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgRemoveApp) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveApp) GetApp() App {
	return NewApp(
		msg.Name,
		msg.RollappId,
		"",
		"",
		"",
	)
}

func (msg *MsgRemoveApp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCreatorAddress, err.Error())
	}

	if len(msg.Name) == 0 {
		return errorsmod.Wrap(ErrInvalidAppName, "App name cannot be empty")
	}

	if len(msg.RollappId) == 0 {
		return errorsmod.Wrap(ErrInvalidRollappID, "RollappId cannot be empty")
	}

	return nil
}
