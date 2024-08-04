package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgCancelBuyOrder{}

// ValidateBasic performs basic validation for the MsgCancelBuyOrder.
func (m *MsgCancelBuyOrder) ValidateBasic() error {
	if !dymnsutils.IsValidBuyOfferId(m.OfferId) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer id is not a valid buy name offer id")
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	return nil
}

// GetSigners returns the required signers for the MsgCancelBuyOrder.
func (m *MsgCancelBuyOrder) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

// Route returns the message router key for the MsgCancelBuyOrder.
func (m *MsgCancelBuyOrder) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgCancelBuyOrder.
func (m *MsgCancelBuyOrder) Type() string {
	return TypeMsgCancelBuyOrder
}

// GetSignBytes returns the raw bytes for the MsgCancelBuyOrder.
func (m *MsgCancelBuyOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
