package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.SequencerKeeper(t)
	params := types.DefaultParams()
	params.MinBond = sdk.NewCoin("testdenom", sdk.NewInt(100))

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
