package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	_ sdk.Msg            = &MsgUpdateParams{}
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
)

// ValidateBasic performs basic validation for the MsgUpdateParams.
func (m *MsgUpdateParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"authority is not a valid bech32 address: %s", m.Authority,
		)
	}

	if m.NewPriceParams == nil && m.NewChainsParams == nil && m.NewMiscParams == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "at least one of the new params must be provided")
	}

	if m.NewPriceParams != nil {
		if err := m.NewPriceParams.Validate(); err != nil {
			return errorsmod.Wrapf(
				errors.Join(gerrc.ErrInvalidArgument, err),
				"failed to validate new price params",
			)
		}
	}

	if m.NewChainsParams != nil {
		if err := m.NewChainsParams.Validate(); err != nil {
			return errorsmod.Wrapf(
				errors.Join(gerrc.ErrInvalidArgument, err),
				"failed to validate new chains params",
			)
		}
	}

	if m.NewMiscParams != nil {
		if err := m.NewMiscParams.Validate(); err != nil {
			return errorsmod.Wrapf(
				errors.Join(gerrc.ErrInvalidArgument, err),
				"failed to validate new misc params",
			)
		}
	}

	return nil
}

// GetSigners returns the required signers for the MsgUpdateParams.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// Route returns the message router key for the MsgUpdateParams.
func (m *MsgUpdateParams) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgUpdateParams.
func (m *MsgUpdateParams) Type() string {
	return TypeMsgUpdateParams
}

// GetSignBytes returns the raw bytes for the MsgUpdateParams.
func (m *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
