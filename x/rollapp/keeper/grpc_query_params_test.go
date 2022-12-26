package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	params.DeployerWhitelist = []types.DeployerParams{{Address: "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3", MaxRollapps: 0}, {Address: "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3", MaxRollapps: 100}}
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.EqualValues(t, params.DisputePeriodInBlocks, response.Params.DisputePeriodInBlocks)
	require.EqualValues(t, len(params.DeployerWhitelist), len(response.Params.DeployerWhitelist))
	for i := range params.DeployerWhitelist {
		require.EqualValues(t, params.DeployerWhitelist[i], response.Params.DeployerWhitelist[i])
	}
}
