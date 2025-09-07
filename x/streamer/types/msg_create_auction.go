package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateAuction = "create_auction"
)

var _ sdk.Msg = &MsgCreateAuction{}

// NewMsgCreateAuction creates a new MsgCreateAuction instance
func NewMsgCreateAuction(
	authority string,
	allocation sdk.Coin,
	startTime time.Time,
	duration time.Duration,
	acceptedTokens []string,
	initialDiscount, maxDiscount math.LegacyDec,
	vestingPeriod time.Duration,
	vestingStartAfterAuctionEnd time.Duration,
	streamParams StreamParams,
) *MsgCreateAuction {
	return &MsgCreateAuction{
		Authority:                   authority,
		Allocation:                  allocation,
		StartTime:                   startTime,
		Duration:                    duration,
		AcceptedTokens:              acceptedTokens,
		InitialDiscount:             initialDiscount,
		MaxDiscount:                 maxDiscount,
		VestingPeriod:               vestingPeriod,
		VestingStartAfterAuctionEnd: vestingStartAfterAuctionEnd,
		StreamParams:                streamParams,
	}
}

// Route returns the name of the module
func (msg MsgCreateAuction) Route() string { return RouterKey }

// Type returns the action
func (msg MsgCreateAuction) Type() string { return TypeMsgCreateAuction }

// ValidateBasic runs basic stateless validity checks
func (msg *MsgCreateAuction) ValidateBasic() error {
	// Validate authority
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
	}

	// Validate allocation
	if !msg.Allocation.IsValid() || msg.Allocation.IsZero() {
		return errorsmod.Wrap(ErrInvalidProposal, "invalid allocation amount")
	}

	// Validate start time
	if msg.StartTime.IsZero() {
		return errorsmod.Wrap(ErrInvalidProposal, "start time cannot be zero")
	}

	// Validate duration
	if msg.Duration <= 0 {
		return errorsmod.Wrap(ErrInvalidProposal, "duration must be positive")
	}

	// Validate accepted tokens
	if len(msg.AcceptedTokens) == 0 {
		return errorsmod.Wrap(ErrInvalidProposal, "at least one accepted token must be specified")
	}

	// Validate discounts
	if msg.InitialDiscount.IsNegative() || msg.InitialDiscount.GTE(math.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidProposal, "initial discount must be between 0 and 1")
	}

	if msg.MaxDiscount.IsNegative() || msg.MaxDiscount.GTE(math.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidProposal, "max discount must be between 0 and 1")
	}

	if msg.InitialDiscount.GT(msg.MaxDiscount) {
		return errorsmod.Wrap(ErrInvalidProposal, "initial discount cannot be greater than max discount")
	}

	// Validate vesting parameters
	if msg.VestingPeriod <= 0 {
		return errorsmod.Wrap(ErrInvalidProposal, "vesting period must be positive")
	}

	if msg.VestingStartAfterAuctionEnd < 0 {
		return errorsmod.Wrap(ErrInvalidProposal, "vesting start delay cannot be negative")
	}

	// Validate stream parameters
	if err := msg.StreamParams.ValidateBasic(); err != nil {
		return errorsmod.Wrap(ErrInvalidProposal, err.Error())
	}

	return nil
}

// GetSignBytes returns the raw bytes for the message
func (msg *MsgCreateAuction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for the message
func (msg *MsgCreateAuction) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// ValidateBasic validates the stream parameters
func (sp StreamParams) ValidateBasic() error {
	if sp.DistrEpochIdentifier == "" {
		return errorsmod.New(ModuleName, 1, "distribution epoch identifier cannot be empty")
	}

	if sp.NumEpochsPaidOver == 0 {
		return errorsmod.New(ModuleName, 2, "number of epochs paid over must be positive")
	}

	return nil
}
