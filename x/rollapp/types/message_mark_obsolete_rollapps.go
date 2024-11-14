package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	TypeMsgMarkObsoleteRollapps = "mark_obsolete_rollapps"
)

var (
	_ sdk.Msg            = new(MsgMarkObsoleteRollapps)
	_ legacytx.LegacyMsg = new(MsgMarkObsoleteRollapps)
)

func (m MsgMarkObsoleteRollapps) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(gerrc.ErrInvalidArgument, errorsmod.Wrap(err, "authority must be a valid bech32 address"))
	}

	if len(m.DrsVersions) == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "at least one DRS version is required")
	}

	return nil
}

func (m MsgMarkObsoleteRollapps) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{signer}
}

func (m MsgMarkObsoleteRollapps) Type() string {
	return TypeMsgMarkObsoleteRollapps
}

func (m MsgMarkObsoleteRollapps) Route() string {
	return RouterKey
}

func (m MsgMarkObsoleteRollapps) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
