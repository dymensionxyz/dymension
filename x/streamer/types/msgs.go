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
	_ sdk.Msg = &MsgCreateStream{}
	_ sdk.Msg = &MsgTerminateStream{}
	_ sdk.Msg = &MsgReplaceStream{}
	_ sdk.Msg = &MsgUpdateStream{}
)

// ValidateBasic checks that the update params message is valid.
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

// ValidateBasic checks that the create stream message is valid.
func (m MsgCreateStream) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	// Either DistributeToRecords or Sponsored must be defined, but not both
	hasDistributeToRecords := len(m.DistributeToRecords) > 0
	if hasDistributeToRecords == m.Sponsored {
		return errors.New("either distribute_to_records or sponsored must be defined, but not both")
	}

	// If using distribute to records, validate them
	if hasDistributeToRecords {
		for _, record := range m.DistributeToRecords {
			if err := record.ValidateBasic(); err != nil {
				return errorsmod.Wrapf(err, "invalid distribution record")
			}
		}
	}

	// Validate coins
	if err := m.Coins.Validate(); err != nil {
		return errorsmod.Wrapf(err, "coins should be valid")
	}

	if m.Coins.Empty() {
		return errors.New("coins should not be empty")
	}

	if !m.Coins.IsAllPositive() {
		return errors.New("all coins must be positive")
	}

	// Validate start time
	if m.StartTime.Equal(time.Time{}) {
		return errors.New("start time should be set")
	}

	// Validate epoch identifier
	if m.DistrEpochIdentifier == "" {
		return errors.New("epoch identifier should be set")
	}

	// Validate number of epochs
	if m.NumEpochsPaidOver == 0 {
		return errors.New("number of epochs paid over should be greater than 0")
	}

	return nil
}

// ValidateBasic checks that the terminate stream message is valid.
func (m MsgTerminateStream) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	if m.StreamId == 0 {
		return errors.New("stream id should be greater than 0")
	}

	return nil
}

// ValidateBasic checks that the replace stream message is valid.
func (m MsgReplaceStream) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	if m.StreamId == 0 {
		return errors.New("stream id should be greater than 0")
	}

	if len(m.Records) == 0 {
		return errors.New("records should not be empty")
	}

	// Validate each record
	for _, record := range m.Records {
		if err := record.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(err, "invalid distribution record")
		}
	}

	return nil
}

// ValidateBasic checks that the update stream message is valid.
func (m MsgUpdateStream) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(
			sdkerrors.ErrInvalidAddress,
			errorsmod.Wrapf(err, "authority must be a valid bech32 address: %s", m.Authority),
		)
	}

	if m.StreamId == 0 {
		return errors.New("stream id should be greater than 0")
	}

	if len(m.Records) == 0 {
		return errors.New("records should not be empty")
	}

	// Validate each record
	for _, record := range m.Records {
		if err := record.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(err, "invalid distribution record")
		}
	}

	return nil
}