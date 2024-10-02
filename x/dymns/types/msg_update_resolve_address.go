package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgUpdateResolveAddress{}

// ValidateBasic performs basic validation for the MsgUpdateResolveAddress.
func (m *MsgUpdateResolveAddress) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if len(m.SubName) > dymnsutils.MaxDymNameLength {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "sub name is too long")
	}

	_, config := m.GetDymNameConfig()
	if err := config.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "config is invalid: %v", err)
	}

	if m.ChainId == "" {
		if m.ResolveTo != "" {
			if !dymnsutils.IsValidBech32AccountAddress(m.ResolveTo, true) {
				return errorsmod.Wrap(
					gerrc.ErrInvalidArgument,
					"resolve address must be a valid bech32 account address on host chain",
				)
			}
		}
	} else if dymnsutils.IsValidEIP155ChainId(m.ChainId) {
		return errorsmod.Wrap(
			gerrc.ErrInvalidArgument,
			"chain-id cannot be numeric-only",
		)
	}

	if !dymnsutils.IsValidBech32AccountAddress(m.Controller, true) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is not a valid bech32 account address")
	}

	return nil
}

// GetDymNameConfig casts MsgUpdateResolveAddress into DymNameConfig.
func (m *MsgUpdateResolveAddress) GetDymNameConfig() (name string, config DymNameConfig) {
	return m.Name, DymNameConfig{
		Type:    DymNameConfigType_DCT_NAME,
		ChainId: m.ChainId,
		Path:    m.SubName,
		Value:   m.ResolveTo,
	}
}

// GetSigners returns the required signers for the MsgUpdateResolveAddress.
func (m *MsgUpdateResolveAddress) GetSigners() []sdk.AccAddress {
	controller, err := sdk.AccAddressFromBech32(m.Controller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{controller}
}

// Route returns the message router key for the MsgUpdateResolveAddress.
func (m *MsgUpdateResolveAddress) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgUpdateResolveAddress.
func (m *MsgUpdateResolveAddress) Type() string {
	return TypeMsgUpdateResolveAddress
}

// GetSignBytes returns the raw bytes for the MsgUpdateResolveAddress.
func (m *MsgUpdateResolveAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
