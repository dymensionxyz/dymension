package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgCancelBuyOrder{}

// ValidateBasic performs basic validation for the MsgCancelBuyOrder.
func (m *MsgCancelBuyOrder) ValidateBasic() error {
	if !IsValidBuyOrderId(m.OrderId) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer id is not a valid buy name offer id")
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	return nil
}
