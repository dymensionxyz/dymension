package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.IROKeeper(t)
	params := types.DefaultParams()
	params.CreationFee = sdk.NewCoin("test", sdk.NewInt(100))
	k.SetParams(ctx, params)
	require.EqualValues(t, params, k.GetParams(ctx))
}
