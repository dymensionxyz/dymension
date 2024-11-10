package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	TypeMsgMarkVulnerableRollapps = "mark_vulnerable_rollapps"
)

var (
	_ sdk.Msg            = new(MsgMarkVulnerableRollapps)
	_ legacytx.LegacyMsg = new(MsgMarkVulnerableRollapps)
)

func (m MsgMarkVulnerableRollapps) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errors.Join(gerrc.ErrInvalidArgument, errorsmod.Wrap(err, "authority must be a valid bech32 address"))
	}

	if len(m.DrsVersions) == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "at least one DRS version is required")
	}

	return nil
}

func (m MsgMarkVulnerableRollapps) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{signer}
}

func (m MsgMarkVulnerableRollapps) Type() string {
	return TypeMsgMarkVulnerableRollapps
}

func (m MsgMarkVulnerableRollapps) Route() string {
	return RouterKey
}

func (m MsgMarkVulnerableRollapps) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
