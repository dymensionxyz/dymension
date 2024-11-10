package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	TypeMsgVote         = "vote"
	TypeMsgRevokeVote   = "revoke_vote"
	TypeMsgUpdateParams = "update_params"
)

var (
	_ sdk.Msg            = &MsgVote{}
	_ sdk.Msg            = &MsgRevokeVote{}
	_ sdk.Msg            = &MsgUpdateParams{}
	_ legacytx.LegacyMsg = &MsgVote{}
	_ legacytx.LegacyMsg = &MsgRevokeVote{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

func (m MsgVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"voter '%s' must be a valid bech32 address: %s",
			m.Voter, err.Error(),
		)
	}

	err = ValidateGaugeWeights(m.Weights)
	if err != nil {
		return ErrInvalidDistribution.Wrap(err.Error())
	}

	return nil
}

func (m MsgVote) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Voter)
	return []sdk.AccAddress{signer}
}

func (m *MsgVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgVote) Route() string {
	return RouterKey
}

func (m *MsgVote) Type() string {
	return TypeMsgVote
}

func (m MsgRevokeVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"voter '%s' must be a valid bech32 address: %s",
			m.Voter, err.Error(),
		)
	}
	return nil
}

func (m MsgRevokeVote) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Voter)
	return []sdk.AccAddress{signer}
}

func (m *MsgRevokeVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgRevokeVote) Route() string {
	return RouterKey
}

func (m *MsgRevokeVote) Type() string {
	return TypeMsgRevokeVote
}

func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{signer}
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"authority '%s' must be a valid bech32 address: %s",
			m.Authority, err.Error(),
		)
	}

	err = m.NewParams.Validate()
	if err != nil {
		return ErrInvalidParams.Wrap(err.Error())
	}

	return nil
}

func (m *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpdateParams) Route() string {
	return RouterKey
}

func (m *MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}
