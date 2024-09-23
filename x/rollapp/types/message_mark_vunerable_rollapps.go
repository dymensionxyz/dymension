package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = new(MsgMarkVulnerableRollapps)

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
