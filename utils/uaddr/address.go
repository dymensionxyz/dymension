package uaddr

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// FromBech32 casts an arbitrary-prefixed bech32 string to either sdk.AccAddress or sdk.ValAddress.
func FromBech32[T sdk.AccAddress | sdk.ValAddress](addr string) (T, error) {
	_, bytes, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return nil, fmt.Errorf("decoding bech32 addr: %w", err)
	}
	return T(bytes), nil
}
