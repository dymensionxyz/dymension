package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgTransferDymNameOwnership{}

// ValidateBasic performs basic validation for the MsgTransferDymNameOwnership.
func (m *MsgTransferDymNameOwnership) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if _, err := sdk.AccAddressFromBech32(m.NewOwner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "new owner is not a valid bech32 account address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	if strings.EqualFold(m.NewOwner, m.Owner) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "new owner must be different from the current owner")
	}

	return nil
}
