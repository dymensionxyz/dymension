package types

import (
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateGauge{}
	_ sdk.Msg = &MsgAddToGauge{}
)

// ValidateBasic checks that the create gauge message is valid.
func (m MsgCreateGauge) ValidateBasic() error {
	if m.Owner == "" {
		return errors.New("owner should be set")
	}
	if m.StartTime.Equal(time.Time{}) {
		return errors.New("distribution start time should be set")
	}
	if m.NumEpochsPaidOver == 0 {
		return errors.New("distribution period should be at least 1 epoch")
	}
	if m.IsPerpetual && m.NumEpochsPaidOver != 1 {
		return errors.New("distribution period should be 1 epoch for perpetual gauge")
	}

	if err := m.Coins.Validate(); err != nil {
		return errorsmod.Wrapf(err, "coins should be valid")
	}

	if m.Coins.Empty() && !m.IsPerpetual {
		return errors.New("coins should be set for non-perpetual gauge")
	}

	switch m.GaugeType {
	case GaugeType_GAUGE_TYPE_ASSET:
		if m.Asset == nil {
			return errors.New("asset must be set for asset gauge type")
		}
		if sdk.ValidateDenom(m.Asset.Denom) != nil {
			return errors.New("denom should be valid for the condition")
		}
		if m.Asset.Duration < 0 || m.Asset.LockAge < 0 {
			return errors.New("duration and lock age should be positive")
		}
		// we explicitly allow empty duration and lock age, as it will distribute to all locks
	case GaugeType_GAUGE_TYPE_ENDORSEMENT:
		if m.Endorsement == nil {
			return errors.New("endorsement must be set for endorsement gauge type")
		}
		if m.Endorsement.RollappId == "" {
			return errors.New("rollapp id should be set")
		}
		if !m.Endorsement.EpochRewards.Empty() {
			return errors.New("epoch rewards should be empty on creation")
		}
	default:
		return errors.New("unsupported gauge type")
	}

	return nil
}

// NewMsgAddToGauge creates a message to add rewards to a specific gauge.
func NewMsgAddToGauge(owner sdk.AccAddress, gaugeId uint64, rewards sdk.Coins) *MsgAddToGauge {
	return &MsgAddToGauge{
		Owner:   owner.String(),
		GaugeId: gaugeId,
		Rewards: rewards,
	}
}

// ValidateBasic checks that the add to gauge message is valid.
func (m MsgAddToGauge) ValidateBasic() error {
	if m.Owner == "" {
		return errors.New("owner should be set")
	}

	if err := m.Rewards.Validate(); err != nil {
		return errorsmod.Wrapf(err, "coins should be valid")
	}
	if m.Rewards.Empty() {
		return errors.New("additional rewards should not be empty")
	}

	return nil
}

func (m MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	err = m.Params.ValidateBasic()
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidRequest,
			errorsmod.Wrapf(err, "failed to validate params"),
		)
	}

	return nil
}
