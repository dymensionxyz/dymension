package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var _ sdk.Msg = &MsgTransferOwnership{}

func (m *MsgTransferOwnership) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return ErrValidationFailed.Wrap("name is not a valid dym name")
	}

	if _, err := sdk.AccAddressFromBech32(m.NewOwner); err != nil {
		return ErrValidationFailed.Wrap("new owner is not a valid bech32 account address")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return ErrValidationFailed.Wrap("owner is not a valid bech32 account address")
	}

	if strings.EqualFold(m.NewOwner, m.Owner) {
		return ErrValidationFailed.Wrap("new owner must be different from the current owner")
	}

	return nil
}

func (m *MsgTransferOwnership) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

func (m *MsgTransferOwnership) Route() string {
	return RouterKey
}

func (m *MsgTransferOwnership) Type() string {
	return TypeMsgTransferOwnership
}

func (m *MsgTransferOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
