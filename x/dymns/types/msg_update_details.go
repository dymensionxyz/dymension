package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgUpdateDetails{}

// ValidateBasic performs basic validation for the MsgUpdateDetails.
func (m *MsgUpdateDetails) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if m.Contact != DoNotModifyDesc {
		if len(m.Contact) > MaxDymNameContactLength {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "contact is too long; max length: %d", MaxDymNameContactLength)
		}
	}

	if _, err := sdk.AccAddressFromBech32(m.Controller); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller is not a valid bech32 account address")
	}

	if m.Contact == DoNotModifyDesc && !m.ClearConfigs {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "message neither clears configs nor updates contact information")
	}

	return nil
}
