package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgRegisterName{}

// ValidateBasic performs basic validation for the MsgRegisterName.
func (m *MsgRegisterName) ValidateBasic() error {
	if len(m.Name) > dymnsutils.MaxDymNameLength {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"name is too long, maximum %d characters", dymnsutils.MaxDymNameLength,
		)
	}

	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if m.Duration < 1 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "duration must be at least 1 year")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address: %s", m.Owner)
	}

	if m.ConfirmPayment.IsNil() || m.ConfirmPayment.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "confirm payment is not set")
	} else if err := m.ConfirmPayment.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid confirm payment: %v", err)
	}

	if len(m.Contact) > MaxDymNameContactLength {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid contact length; got: %d, max: %d", len(m.Contact), MaxDymNameContactLength)
	}

	return nil
}
