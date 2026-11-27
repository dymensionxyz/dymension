package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestGenesis_EscrowAndWindowStateRoundTrip(t *testing.T) {
	ctx, k, _ := setup(t)

	params := types.DefaultParams()
	params.AgentLivenessWindowBlocks = 100_800
	g := types.GenesisState{
		Params: params,
		Agents: []types.Agent{{
			Id:                     "a1",
			Owner:                  owner(t),
			Active:                 true,
			ActionSeq:              3,
			SpendDenom:             "adym",
			SpendLimitPerWindow:    math.NewInt(1000),
			SpendWindowBlocks:      10,
			SpendWindowStartHeight: 20,
			SpendWindowSpent:       math.NewInt(400),
			LastActionHeight:       17,
		}},
		Escrows: []types.AgentEscrow{{
			AgentId: "a1",
			Balance: sdk.NewCoins(sdk.NewInt64Coin("adym", 600)),
		}},
	}
	keeper.InitGenesis(ctx, k, g)

	out := keeper.ExportGenesis(ctx, k)
	require.Equal(t, g.Params, out.Params)
	require.Equal(t, g.Agents, out.Agents)
	require.Equal(t, g.Escrows, out.Escrows)

	// the imported window state is live: 600 remaining in escrow, 600 left in
	// the window budget at heights within the stored bucket
	agent, found := k.GetAgent(ctx, "a1")
	require.True(t, found)
	require.Equal(t, math.NewInt(600), agent.RemainingWindowBudget(25))
	require.Equal(t, math.NewInt(600), k.GetEscrowBalance(ctx, "a1").AmountOf("adym"))
}

// A zero-height restart must not reset heartbeat ages: block 1 of the new
// chain follows the export height, so every agent keeps its elapsed age.
func TestGenesis_ZeroHeightRebasePreservesHeartbeatAge(t *testing.T) {
	const exportHeight = 1_000_000
	ctx, k, _ := setup(t)
	ctx = ctx.WithBlockHeight(exportHeight)
	params := types.DefaultParams()
	params.AgentLivenessWindowBlocks = 100
	require.NoError(t, k.SetParams(ctx, params))
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "stale", Owner: owner(t), Active: true, LastActionHeight: 900_000}))
	// age 99: on the old chain it is fresh at exportHeight+1 and stale after
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "edge", Owner: owner(t), Active: true, LastActionHeight: exportHeight - 99}))
	require.False(t, k.IsAgentFresh(ctx.WithBlockHeight(exportHeight+1), "stale"))
	require.True(t, k.IsAgentFresh(ctx.WithBlockHeight(exportHeight+1), "edge"))
	require.False(t, k.IsAgentFresh(ctx.WithBlockHeight(exportHeight+2), "edge"))

	require.NoError(t, k.RebaseLastActionHeights(ctx, exportHeight))
	g := keeper.ExportGenesis(ctx, k)
	require.NoError(t, g.Validate())

	restarted, k2, _ := setup(t)
	keeper.InitGenesis(restarted, k2, *g)
	require.False(t, k2.IsAgentFresh(restarted.WithBlockHeight(1), "stale"))
	require.True(t, k2.IsAgentFresh(restarted.WithBlockHeight(1), "edge"))
	require.False(t, k2.IsAgentFresh(restarted.WithBlockHeight(2), "edge"))
}
