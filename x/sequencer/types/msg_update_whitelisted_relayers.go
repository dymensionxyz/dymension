package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uaddr"
)

const (
	maxWhitelistedRelayers = 10

	TypeMsgUpdateWhitelistedRelayers = "update_whitelisted_relayers"
)

var (
	_ sdk.Msg            = new(MsgUpdateWhitelistedRelayers)
	_ legacytx.LegacyMsg = new(MsgUpdateWhitelistedRelayers)
)

func (m *MsgUpdateWhitelistedRelayers) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "get creator addr from bech32")
	}
	err = ValidateWhitelistedRelayers(m.Relayers)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "validate whitelisted relayers")
	}
	return nil
}

func ValidateWhitelistedRelayers(wr []string) error {
	if len(wr) > maxWhitelistedRelayers {
		return fmt.Errorf("maximum allowed relayers is %d", maxWhitelistedRelayers)
	}
	relayers := make(map[string]struct{}, len(wr))
	for _, r := range wr {
		if _, ok := relayers[r]; ok {
			return fmt.Errorf("duplicated relayer: %s", r)
		}
		relayers[r] = struct{}{}

		relayer, err := uaddr.FromBech32[sdk.AccAddress](r)
		if err != nil {
			return fmt.Errorf("convert bech32 to relayer address: %s: %w", r, err)
		}
		err = sdk.VerifyAddressFormat(relayer)
		if err != nil {
			return fmt.Errorf("invalid relayer address: %s: %w", r, err)
		}
	}
	return nil
}

func (m *MsgUpdateWhitelistedRelayers) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Creator)
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateWhitelistedRelayers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}

func (m *MsgUpdateWhitelistedRelayers) Route() string {
	return RouterKey
}

func (m *MsgUpdateWhitelistedRelayers) Type() string {
	return TypeMsgUpdateWhitelistedRelayers
}
