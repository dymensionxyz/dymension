package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// sanity, TODO: delete
func TestCoin(t *testing.T) {
	coin := sdk.NewCoin("adym", math.NewInt(100))
	require.NoError(t, coin.Validate())
}
