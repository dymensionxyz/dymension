package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgPlaceSellOrder{}

// ValidateBasic performs basic validation for the MsgPlaceSellOrder.
func (m *MsgPlaceSellOrder) ValidateBasic() error {
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

	so := m.ToSellOrder()

	// put a dummy expire at to validate, as zero expire at is invalid,
	// and we don't have context of time at this point
	so.ExpireAt = 1

	if err := so.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order: %v", err.Error())
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	return nil
}

// ToSellOrder converts the MsgPlaceSellOrder to a SellOrder.
func (m *MsgPlaceSellOrder) ToSellOrder() SellOrder {
	so := SellOrder{
		AssetId:   m.AssetId,
		AssetType: m.AssetType,
		MinPrice:  m.MinPrice,
		SellPrice: m.SellPrice,
	}

	if !so.HasSetSellPrice() {
		so.SellPrice = nil
	}

	return so
}
