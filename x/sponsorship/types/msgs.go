package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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
		return gerrc.ErrInvalidArgument.Wrapf(
			"voter '%s' must be a valid bech32 address: %s",
			m.Voter, err.Error(),
		)
	}

	err = ValidateGaugeWeights(m.Weights)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidDistribution, err.Error())
	}

	return nil
}

func (m MsgRevokeVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return gerrc.ErrInvalidArgument.Wrapf(
			"voter '%s' must be a valid bech32 address: %s",
			m.Voter, err.Error(),
		)
	}
	return nil
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return gerrc.ErrInvalidArgument.Wrapf(
			"authority '%s' must be a valid bech32 address: %s",
			m.Authority, err.Error(),
		)
	}

	err = m.NewParams.ValidateBasic()
	if err != nil {
		return errorsmod.Wrap(ErrInvalidParams, err.Error())
	}

	return nil
}

func (m MsgClaimRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return gerrc.ErrInvalidArgument.Wrapf(
			"sender '%s' must be a valid bech32 address: %s",
			m.Sender, err.Error(),
		)
	}
	return nil
}
