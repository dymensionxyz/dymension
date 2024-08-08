package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		return errors.Join(ErrInvalidDistribution, err)
	}

	return nil
}

func (m MsgVote) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Voter)
	return []sdk.AccAddress{signer}
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
