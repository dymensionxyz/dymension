package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgUpdateResolveAddress{}

// ValidateBasic performs basic validation for the MsgUpdateResolveAddress.
func (m *MsgUpdateResolveAddress) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return ErrValidationFailed.Wrap("name is not a valid dym name")
	}

	if len(m.SubName) > MaxDymNameLength {
		return ErrDymNameTooLong
	}

	_, config := m.GetDymNameConfig()
	if err := config.Validate(); err != nil {
		return ErrValidationFailed.Wrapf("config is invalid: %v", err)
	}

	if m.ChainId == "" {
		if m.ResolveTo != "" {
			if _, err := sdk.AccAddressFromBech32(m.ResolveTo); err != nil {
				return ErrValidationFailed.Wrap(
					"resolve address must be a valid bech32 account address on host chain",
				)
			}
		}
	}

	if _, err := sdk.AccAddressFromBech32(m.Controller); err != nil {
		return ErrValidationFailed.Wrap("controller is not a valid bech32 account address")
	}

	return nil
}

// GetDymNameConfig casts MsgUpdateResolveAddress into DymNameConfig.
func (m *MsgUpdateResolveAddress) GetDymNameConfig() (name string, config DymNameConfig) {
	return m.Name, DymNameConfig{
		Type:    DymNameConfigType_NAME,
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
