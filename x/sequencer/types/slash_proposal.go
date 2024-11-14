package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgSlashSequencerProposal{}

// ValidateBasic performs basic validation for the MsgSlashSequencerProposal.
func (m *MsgSlashSequencerProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"authority is not a valid bech32 address: %s", m.Authority,
		)
	}
	if m.Rewardee != "" {
		if _, err := sdk.AccAddressFromBech32(m.Rewardee); err != nil {
			return errorsmod.Wrapf(
				errors.Join(gerrc.ErrInvalidArgument, err),
				"rewardee is not a valid bech32 address: %s", m.Authority,
			)
		}
	}

	return nil
}

// Returns acc address if rewardee field is not empty
func (m *MsgSlashSequencerProposal) MustRewardee() *sdk.AccAddress {
	if m.Rewardee == "" {
		return nil
	}
	rewardee, _ := sdk.AccAddressFromBech32(m.Rewardee)
	return &rewardee
}

// GetSigners returns the required signers for the MsgSlashSequencerProposal.
func (m *MsgSlashSequencerProposal) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}
