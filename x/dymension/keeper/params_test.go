package keeper_test

import (
	"testing"

	testkeeper "github.com/dymensionxyz/dYmension/testutil/keeper"
	"github.com/dymensionxyz/dYmension/x/dymension/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.DymensionKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
