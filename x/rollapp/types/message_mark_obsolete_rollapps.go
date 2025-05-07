package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = new(MsgMarkObsoleteRollapps)

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
