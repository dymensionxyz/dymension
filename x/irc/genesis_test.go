package irc_test

import (
	"testing"

	keepertest "github.com/dymensionxyz/dymension/testutil/keeper"
	"github.com/dymensionxyz/dymension/testutil/nullify"
	"github.com/dymensionxyz/dymension/x/irc"
	"github.com/dymensionxyz/dymension/x/irc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		IRCRequestList: []types.IRCRequest{
			{
				ReqId: 0,
			},
			{
				ReqId: 1,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, _, ctx := keepertest.IRCKeeper(t)
	irc.InitGenesis(ctx, k, genesisState)
	got := irc.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.IRCRequestList, got.IRCRequestList)
	// this line is used by starport scaffolding # genesis/test/assert
}
