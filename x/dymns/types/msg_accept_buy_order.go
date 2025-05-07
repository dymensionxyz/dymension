package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgAcceptBuyOrder{}

// ValidateBasic performs basic validation for the MsgAcceptBuyOrder.
func (m *MsgAcceptBuyOrder) ValidateBasic() error {
	if !IsValidBuyOrderId(m.OrderId) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer id is not a valid buy name offer id")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	if !m.MinAccept.IsValid() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid min-accept amount")
	} else if !m.MinAccept.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min-accept amount must be positive")
	}

	return nil
}
