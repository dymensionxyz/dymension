package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	TypeMsgSetCanonicalClient = "set_canonical_client"
	TypeMsgUpdateClient       = "update_client"
)

var (
	_ sdk.Msg = &MsgSetCanonicalClient{}
	_ sdk.Msg = &MsgUpdateClient{}
)

func (msg *MsgSetCanonicalClient) Route() string {
	return ModuleName
}

func (msg *MsgSetCanonicalClient) Type() string {
	return TypeMsgSetCanonicalClient
}

func (msg *MsgSetCanonicalClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetCanonicalClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid creator address (%s)", err)
	}
	if msg.ClientId == "" {
		return gerrc.ErrInvalidArgument.Wrap("empty client id")
	}
	return nil
}

func (msg *MsgUpdateClient) Route() string {
	return ModuleName
}

func (msg *MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

func (msg *MsgUpdateClient) GetSigners() []sdk.AccAddress {
	return msg.Inner.GetSigners()
}

func (msg *MsgUpdateClient) ValidateBasic() error {
	if msg.Inner == nil {
		return gerrc.ErrInvalidArgument.Wrap("inner is nil")
	}
	return msg.Inner.ValidateBasic()
}
