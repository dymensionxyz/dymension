package rollapp_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/rollapp"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		RollappList: []types.Rollapp{
			{
				RollappId: "0",
			},
			{
				RollappId: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RollappKeeper(t)
	rollapp.InitGenesis(ctx, *k, genesisState)
	got := rollapp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.RollappList, got.RollappList)
	// this line is used by starport scaffolding # genesis/test/assert
}
