package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgCancelSellOrder{}

// ValidateBasic performs basic validation for the MsgCancelSellOrder.
func (m *MsgCancelSellOrder) ValidateBasic() error {
	if m.AssetType == TypeName {
		if !dymnsutils.IsValidDymName(m.AssetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "name is not a valid dym name: %s", m.AssetId)
		}
	} else if m.AssetType == TypeAlias {
		if !dymnsutils.IsValidAlias(m.AssetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias is not a valid alias: %s", m.AssetId)
		}
	} else {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", m.AssetType)
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	return nil
}
