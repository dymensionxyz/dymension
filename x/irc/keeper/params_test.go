package keeper_test

import (
	"testing"

	testkeeper "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/x/irc/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, _, ctx := testkeeper.IRCKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
