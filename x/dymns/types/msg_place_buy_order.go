package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgPlaceBuyOrder{}

// ValidateBasic performs basic validation for the MsgPlaceBuyOrder.
func (m *MsgPlaceBuyOrder) ValidateBasic() error {
	switch m.AssetType {
	case TypeName:
		if !dymnsutils.IsValidDymName(m.AssetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "name is not a valid dym name: %s", m.AssetId)
		}
	case TypeAlias:
		if !dymnsutils.IsValidAlias(m.AssetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias is not a valid alias: %s", m.AssetId)
		}
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", m.AssetType)
	}

	if err := ValidateOrderParams(m.Params, m.AssetType); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	if m.ContinueOrderId != "" && !IsValidBuyOrderId(m.ContinueOrderId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"continue offer id is not a valid offer id: %s", m.ContinueOrderId,
		)
	}

	if !m.Offer.IsValid() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid offer amount")
	} else if !m.Offer.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer amount must be positive")
	}

	return nil
}
