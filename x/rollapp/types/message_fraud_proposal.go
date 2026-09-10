package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgRollappFraudProposal{}

// ValidateBasic performs basic validation for the MsgRollappFraudProposal.
func (m *MsgRollappFraudProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"authority is not a valid bech32 address: %s", m.Authority,
		)
	}
	if _, err := NewChainID(m.RollappId); err != nil {
		return errorsmod.Wrap(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"rollapp ID is invalid",
		)
	}
	if m.FraudHeight == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "fraud height must be greater than zero")
	}
	if m.PunishSequencerAddress != "" {
		if _, err := sdk.AccAddressFromBech32(m.PunishSequencerAddress); err != nil {
			return errorsmod.Wrapf(
				errors.Join(gerrc.ErrInvalidArgument, err),
				"punish sequencer address is not a valid bech32 address: %s", m.PunishSequencerAddress,
			)
		}
	}
	if m.Rewardee != "" {
		if _, err := sdk.AccAddressFromBech32(m.Rewardee); err != nil {
			return errorsmod.Wrapf(
				errors.Join(gerrc.ErrInvalidArgument, err),
				"rewardee is not a valid bech32 address: %s", m.Rewardee,
			)
		}
	}

	return nil
}

// Returns acc address if rewardee field is not empty
func (m *MsgRollappFraudProposal) MustRewardee() *sdk.AccAddress {
	if m.Rewardee == "" {
		return nil
	}
	rewardee, _ := sdk.AccAddressFromBech32(m.Rewardee)
	return &rewardee
}
