package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.RollappKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params.DisputePeriodInBlocks, k.DisputePeriodInBlocks(ctx))
	require.EqualValues(t, params.RegistrationFee, k.RegistrationFee(ctx))
}

func TestGetParamsWithRegistrationFee(t *testing.T) {
	k, ctx := testkeeper.RollappKeeper(t)
	params := types.DefaultParams()

	params.RegistrationFee, _ = sdk.ParseCoinNormalized(registrationFee)
	k.SetParams(ctx, params)

	require.EqualValues(t, params.RegistrationFee, k.RegistrationFee(ctx))
}
