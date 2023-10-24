package types

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgCreateStream = "create_stream"
)

var _ sdk.Msg = &MsgCreateStream{}

// NewMsgCreateStream creates a message to create a stream with the provided parameters.
func NewMsgCreateStream(distributeTo sdk.AccAddress, coins sdk.Coins, startTime time.Time, distrEpochIdentifier string, numEpochsPaidOver uint64) *MsgCreateStream {
	return &MsgCreateStream{
		DistributeTo:         distributeTo.String(),
		Coins:                coins,
		StartTime:            startTime,
		DistrEpochIdentifier: distrEpochIdentifier,
		NumEpochsPaidOver:    numEpochsPaidOver,
	}
}

// Route takes a create stream message, then returns the RouterKey used for slashing.
func (m MsgCreateStream) Route() string { return RouterKey }

// Type takes a create stream message, then returns a create stream message type.
func (m MsgCreateStream) Type() string { return TypeMsgCreateStream }

// ValidateBasic checks that the create stream message is valid.
func (m MsgCreateStream) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.DistributeTo)
	if err != nil {
		return errors.New("distribute_to should be set")
	}

	if m.Coins.Empty() {
		return errors.New("initial rewards should not be empty")
	}

	for _, coin := range m.Coins {
		if sdk.ValidateDenom(coin.Denom) != nil {
			return errors.New("denom should be valid for the condition")
		}
	}
	if m.StartTime.Equal(time.Time{}) {
		return errors.New("distribution start time should be set")
	}

	if m.DistrEpochIdentifier == "" {
		return errors.New("distribution epoch identifier should be set")
	}

	if m.NumEpochsPaidOver == 0 {
		return errors.New("distribution period should be at least 1 epoch")
	}

	return nil
}

// GetSignBytes takes a create stream message and turns it into a byte array.
func (m MsgCreateStream) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners takes a create stream message and returns the owner in a byte array.
func (m MsgCreateStream) GetSigners() []sdk.AccAddress {
	// owner, _ := sdk.AccAddressFromBech32(m.Owner)
	return []sdk.AccAddress{}
}
