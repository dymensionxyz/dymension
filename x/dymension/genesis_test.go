package dymension_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dYmension/testutil/keeper"
	"github.com/dymensionxyz/dYmension/testutil/nullify"
	"github.com/dymensionxyz/dYmension/x/dymension"
	"github.com/dymensionxyz/dYmension/x/dymension/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.DymensionKeeper(t)
	dymension.InitGenesis(ctx, *k, genesisState)
	got := dymension.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
