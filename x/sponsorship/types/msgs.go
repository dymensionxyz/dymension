package types

import (
	dymerrors "github.com/dymensionxyz/dymension/v3/internal/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "voter '%s' must be a valid bech32 address", m.Voter)
	}

	err = ValidateGaugeWeights(m.Weights)
	if err != nil {
		return dymerrors.Join(ErrInvalidDistribution, err)
	}

	return nil
}

func (m MsgRevokeVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "voter '%s' must be a valid bech32 address", m.Voter)
	}
	return nil
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "authority '%s' must be a valid bech32 address", m.Authority)
	}

	err = m.NewParams.ValidateBasic()
	if err != nil {
		return dymerrors.Join(ErrInvalidParams, err)
	}

	return nil
}

func (m MsgClaimRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return dymerrors.Joinf(sdkerrors.ErrInvalidAddress, err, "sender '%s' must be a valid bech32 address", m.Sender)
	}
	return nil
}
