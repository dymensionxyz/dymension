package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgKickProposer{}

func NewMsgKickProposer(creator string) *MsgKickProposer {
	return &MsgKickProposer{
		Creator: creator,
	}
}

func (m *MsgKickProposer) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "get creator addr from bech32")
	}
	return nil
}
