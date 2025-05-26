package types

import (
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateGauge{}
	_ sdk.Msg = &MsgAddToGauge{}
)

// NewMsgCreateAssetGauge creates a message to create a gauge with the provided parameters.
func NewMsgCreateAssetGauge(isPerpetual bool, owner sdk.AccAddress, distributeTo lockuptypes.QueryCondition, coins sdk.Coins, startTime time.Time, numEpochsPaidOver uint64) *MsgCreateGauge {
	return &MsgCreateGauge{
		IsPerpetual:       isPerpetual,
		Owner:             owner.String(),
		DistributeTo:      &MsgCreateGauge_Asset{Asset: &distributeTo},
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
	}
}

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

	switch distr := m.DistributeTo.(type) {
	case *MsgCreateGauge_Asset:
		condition := distr.Asset
		if sdk.ValidateDenom(condition.Denom) != nil {
			return errors.New("denom should be valid for the condition")
		}
		if condition.Duration < 0 || condition.LockAge < 0 {
			return errors.New("duration and lock age should be positive")
		}
	case *MsgCreateGauge_Endorsement:
		if distr.Endorsement.RollappId == "" {
			return errors.New("rollapp id should be set")
		}
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
