package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = (*MsgUpdateRewardAddress)(nil)

func (m *MsgUpdateRewardAddress) ValidateBasic() error {
	_, err := sdk.ValAddressFromBech32(m.Creator)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "get creator addr from bech32")
	}
	_, err = sdk.AccAddressFromBech32(m.RewardAddr)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "get reward addr from bech32")
	}
	return nil
}

func (m *MsgUpdateRewardAddress) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.ValAddressFromBech32(m.Creator)
	return []sdk.AccAddress{sdk.AccAddress(addr)}
}
