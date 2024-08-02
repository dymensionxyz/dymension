package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.RollappKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.EqualValues(t, params.DisputePeriodInBlocks, response.Params.DisputePeriodInBlocks)
	require.Equal(t, params.RegistrationFee, response.Params.RegistrationFee)
}
