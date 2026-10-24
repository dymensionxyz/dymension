package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

func TestMsgBuy_ValidateBasic(t *testing.T) {
	tcs := []struct {
		name    string
		maxCost math.Int
		expErr  error
	}{
		{name: "unset", maxCost: math.Int{}},
		{name: "zero", maxCost: math.ZeroInt()},
		{name: "positive", maxCost: math.NewInt(1)},
		{name: "negative", maxCost: math.NewInt(-1), expErr: types.ErrInvalidPurchaseAmount},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.MsgBuy{MaxCost: tc.maxCost}
			require.ErrorIs(t, msg.ValidateBasic(), tc.expErr)
		})
	}
}

func TestMsgBuyExactSpend_ValidateBasic(t *testing.T) {
	tcs := []struct {
		name         string
		minTokensOut math.Int
		expErr       error
	}{
		{name: "unset", minTokensOut: math.Int{}},
		{name: "zero", minTokensOut: math.ZeroInt()},
		{name: "positive", minTokensOut: math.NewInt(1)},
		{name: "negative", minTokensOut: math.NewInt(-1), expErr: types.ErrInvalidPurchaseAmount},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			msg := types.MsgBuyExactSpend{MinTokensOut: tc.minTokensOut}
			require.ErrorIs(t, msg.ValidateBasic(), tc.expErr)
		})
	}
}
