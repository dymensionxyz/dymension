package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	TypeMsgUpdateRewardAddress = "update_reward_address"
)

var (
	_ sdk.Msg            = (*MsgUpdateRewardAddress)(nil)
	_ legacytx.LegacyMsg = (*MsgUpdateRewardAddress)(nil)
)

func (m *MsgUpdateRewardAddress) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Creator)
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
	addr, _ := sdk.AccAddressFromBech32(m.Creator)
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateRewardAddress) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpdateRewardAddress) Route() string {
	return RouterKey
}

func (m *MsgUpdateRewardAddress) Type() string {
	return TypeMsgUpdateRewardAddress
}
