package ucoin

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMulDec(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		d, _ := sdk.NewDecFromStr("0.5")
		coins := sdk.NewCoins(
			sdk.NewCoin("foo", sdk.NewInt(2)),
			sdk.NewCoin("bar", sdk.NewInt(3)),
		)
		res := MulDec(d, coins...)
		require.Equal(t, sdk.NewInt(1), res[0].Amount)
		require.Equal(t, sdk.NewInt(1), res[0].Amount)
	})
}
