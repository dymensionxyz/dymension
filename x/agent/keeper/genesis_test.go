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

	g := types.GenesisState{
		Params: types.DefaultParams(),
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
		}},
		Escrows: []types.AgentEscrow{{
			AgentId: "a1",
			Balance: sdk.NewCoins(sdk.NewInt64Coin("adym", 600)),
		}},
	}
	keeper.InitGenesis(ctx, k, g)

	out := keeper.ExportGenesis(ctx, k)
	require.Equal(t, g.Agents, out.Agents)
	require.Equal(t, g.Escrows, out.Escrows)

	// the imported window state is live: 600 remaining in escrow, 600 left in
	// the window budget at heights within the stored bucket
	agent, found := k.GetAgent(ctx, "a1")
	require.True(t, found)
	require.Equal(t, math.NewInt(600), agent.RemainingWindowBudget(25))
	require.Equal(t, math.NewInt(600), k.GetEscrowBalance(ctx, "a1").AmountOf("adym"))
}
