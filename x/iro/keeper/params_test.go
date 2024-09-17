package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.IROKeeper(t)
	params := types.DefaultParams()
	params.CreationFee = math.NewInt(100).MulRaw(1e18)
	k.SetParams(ctx, params)
	require.EqualValues(t, params, k.GetParams(ctx))
}
