package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	_ sdk.Msg = &MsgSetCanonicalClient{}
)

func (msg *MsgSetCanonicalClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid creator address (%s)", err)
	}
	if msg.ClientId == "" {
		return gerrc.ErrInvalidArgument.Wrap("empty client id")
	}
	return nil
}
