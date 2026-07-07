package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func spendingAgent() types.Agent {
	return types.Agent{
		Id:                     "a1",
		Active:                 true,
		SpendDenom:             "adym",
		SpendLimitPerWindow:    math.NewInt(100),
		SpendWindowBlocks:      10,
		SpendWindowStartHeight: 20,
		SpendWindowSpent:       math.NewInt(40),
	}
}

func TestGenesisValidate(t *testing.T) {
	valid := func() types.GenesisState {
		return types.GenesisState{
			Params: types.DefaultParams(),
			Agents: []types.Agent{spendingAgent()},
			Escrows: []types.AgentEscrow{
				{AgentId: "a1", Balance: sdk.NewCoins(sdk.NewInt64Coin("adym", 5))},
			},
		}
	}

	require.NoError(t, valid().Validate())
	require.NoError(t, types.DefaultGenesis().Validate())

	tests := []struct {
		name   string
		mutate func(*types.GenesisState)
	}{
		{"enabled agent with zero window blocks", func(g *types.GenesisState) { g.Agents[0].SpendWindowBlocks = 0 }},
		{"enabled agent with nil limit", func(g *types.GenesisState) { g.Agents[0].SpendLimitPerWindow = math.Int{} }},
		{"enabled agent with invalid denom", func(g *types.GenesisState) { g.Agents[0].SpendDenom = "1" }},
		{"window spent above limit", func(g *types.GenesisState) { g.Agents[0].SpendWindowSpent = math.NewInt(101) }},
		{"negative window spent", func(g *types.GenesisState) { g.Agents[0].SpendWindowSpent = math.NewInt(-1) }},
		{"disabled agent with window blocks", func(g *types.GenesisState) {
			g.Agents[0] = types.Agent{Id: "a1", SpendWindowBlocks: 10}
		}},
		{"disabled agent with window state", func(g *types.GenesisState) {
			g.Agents[0] = types.Agent{Id: "a1", SpendWindowSpent: math.NewInt(1)}
		}},
		{"escrow for unknown agent", func(g *types.GenesisState) { g.Escrows[0].AgentId = "ghost" }},
		{"duplicate escrow entry", func(g *types.GenesisState) { g.Escrows = append(g.Escrows, g.Escrows[0]) }},
		{"empty escrow balance", func(g *types.GenesisState) { g.Escrows[0].Balance = sdk.NewCoins() }},
		{"invalid escrow balance", func(g *types.GenesisState) {
			g.Escrows[0].Balance = sdk.Coins{sdk.Coin{Denom: "adym", Amount: math.NewInt(-1)}}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := valid()
			tc.mutate(&g)
			require.Error(t, g.Validate())
		})
	}
}

// TestGenesisValidate_RuntimeStatesAccepted ensures validation does not reject
// states the runtime actually produces: a freshly registered agent (nil Ints)
// and an exported disabled agent (zero-decoded Ints).
func TestGenesisValidate_RuntimeStatesAccepted(t *testing.T) {
	g := types.GenesisState{
		Params: types.DefaultParams(),
		Agents: []types.Agent{
			{Id: "fresh"},
			{Id: "exported-disabled", SpendLimitPerWindow: math.ZeroInt(), SpendWindowSpent: math.ZeroInt()},
		},
	}
	require.NoError(t, g.Validate())
}
