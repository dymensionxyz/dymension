package types

import (
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
		return ErrInvalidDistribution.Wrap(err.Error())
	}

	return nil
}

func (m MsgVote) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Voter)
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

func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{signer}
}
