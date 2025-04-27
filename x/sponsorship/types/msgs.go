package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgVote         = "vote"
	TypeMsgRevokeVote   = "revoke_vote"
	TypeMsgClaimRewards = "claim_rewards"
	TypeMsgUpdateParams = "update_params"
)

var (
	_ sdk.Msg = &MsgVote{}
	_ sdk.Msg = &MsgRevokeVote{}
	_ sdk.Msg = &MsgClaimRewards{}
	_ sdk.Msg = &MsgUpdateParams{}
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

	err = m.NewParams.ValidateBasic()
	if err != nil {
		return ErrInvalidParams.Wrap(err.Error())
	}

	return nil
}

func (m *MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

func (m MsgClaimRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf(
			"sender '%s' must be a valid bech32 address: %s",
			m.Sender, err.Error(),
		)
	}
	return nil
}

func (m MsgClaimRewards) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{signer}
}

func (m *MsgClaimRewards) Type() string {
	return TypeMsgClaimRewards
}
