package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgSetController{}

// ValidateBasic performs basic validation for the MsgSetController.
func (m *MsgSetController) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if _, err := sdk.AccAddressFromBech32(m.Controller); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is not a valid bech32 account address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	return nil
}

// GetSigners returns the required signers for the MsgSetController.
func (m *MsgSetController) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgSetController.
func (m *MsgSetController) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgSetController.
func (m *MsgSetController) Type() string {
	return TypeMsgSetController
}

// GetSignBytes returns the raw bytes for the MsgSetController.
func (m *MsgSetController) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
