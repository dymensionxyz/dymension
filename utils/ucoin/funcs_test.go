package ucoin

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMulDec(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		coins := sdk.NewCoin("foo", sdk.NewInt(190238109238))
	})
}
