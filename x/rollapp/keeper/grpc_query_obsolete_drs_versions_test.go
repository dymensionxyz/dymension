package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestObsoleteDRSVersionsQuery(t *testing.T) {
	keeper, ctx := testkeeper.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	const obsoleteDRSVersion uint32 = 1234567890
	err := keeper.SetObsoleteDRSVersion(ctx, obsoleteDRSVersion)
	require.NoError(t, err)

	response, err := keeper.ObsoleteDRSVersions(wctx, &types.QueryObsoleteDRSVersionsRequest{})
	require.NoError(t, err)

	expected, err := keeper.GetAllObsoleteDRSVersions(ctx)
	require.NoError(t, err)
	require.EqualValues(t, expected, response.DrsVersions)
}
