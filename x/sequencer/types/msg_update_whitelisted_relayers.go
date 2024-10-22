package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = new(MsgUpdateWhitelistedRelayers)

func (m *MsgUpdateWhitelistedRelayers) ValidateBasic() error {
	_, err := sdk.ValAddressFromBech32(m.Creator)
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
	relayers := make(map[string]struct{}, len(wr))
	for _, r := range wr {
		if _, ok := relayers[r]; ok {
			return fmt.Errorf("duplicated relayer: %s", r)
		}
		relayers[r] = struct{}{}

		relayer, err := Bech32ToAddr[sdk.AccAddress](r)
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

// Bech32ToAddr casts an arbitrary-prefixed bech32 string to either sdk.AccAddress or sdk.ValAddress.
// TODO: move to sdk-utils.
func Bech32ToAddr[T sdk.AccAddress | sdk.ValAddress](addr string) (T, error) {
	_, bytes, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return nil, fmt.Errorf("decoding bech32 addr: %w", err)
	}
	return T(bytes), nil
}

func (m *MsgUpdateWhitelistedRelayers) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.ValAddressFromBech32(m.Creator)
	return []sdk.AccAddress{sdk.AccAddress(addr)}
}
