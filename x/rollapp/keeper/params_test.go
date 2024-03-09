package keeper_test

import (
	"testing"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.RollappKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params.DisputePeriodInBlocks, k.DisputePeriodInBlocks(ctx))
	require.EqualValues(t, len(params.DeployerWhitelist), len(k.DeployerWhitelist(ctx)))
	for i := range params.DeployerWhitelist {
		require.EqualValues(t, params.DeployerWhitelist[i], k.DeployerWhitelist(ctx)[i])
	}
}

func TestGetParamsWithDeployerWhitelist(t *testing.T) {
	k, ctx := testkeeper.RollappKeeper(t)

	params := types.DefaultParams()
	params.DeployerWhitelist = []types.DeployerParams{{Address: sample.AccAddress()}, {Address: sample.AccAddress()}}

	k.SetParams(ctx, params)

	require.EqualValues(t, params.DisputePeriodInBlocks, k.DisputePeriodInBlocks(ctx))
	require.EqualValues(t, len(params.DeployerWhitelist), len(k.DeployerWhitelist(ctx)))
	for i := range params.DeployerWhitelist {
		require.EqualValues(t, params.DeployerWhitelist[i], k.DeployerWhitelist(ctx)[i])
	}
}
