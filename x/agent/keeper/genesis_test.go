package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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
			SpendPrevWindowSpent:   math.NewInt(500),
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

	// the imported window state is live: 600 remaining in escrow; at height 25
	// the window budget is 1000 - 400 - 500*(10-5)/10 = 350
	agent, found := k.GetAgent(ctx, "a1")
	require.True(t, found)
	require.Equal(t, math.NewInt(350), agent.RemainingWindowBudget(25))
	require.Equal(t, math.NewInt(600), k.GetEscrowBalance(ctx, "a1").AmountOf("adym"))
}

// TestGenesis_AgentWithoutPrevWindowSpentImportsAsZero covers agents exported
// before spend_prev_window_spent existed: the absent field must validate and
// import as zero previous-window spend.
func TestGenesis_AgentWithoutPrevWindowSpentImportsAsZero(t *testing.T) {
	ctx, k, _ := setup(t)
	cdc := codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())

	params := types.DefaultParams()
	var g types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON([]byte(`{
		"params": `+string(cdc.MustMarshalJSON(&params))+`,
		"agents": [{
			"id": "a1",
			"spend_denom": "adym",
			"spend_limit_per_window": "1000",
			"spend_window_blocks": "10",
			"spend_window_start_height": "20",
			"spend_window_spent": "400"
		}]
	}`), &g))
	require.True(t, g.Agents[0].SpendPrevWindowSpent.IsNil())
	require.NoError(t, g.Validate())

	keeper.InitGenesis(ctx, k, g)
	agent, found := k.GetAgent(ctx, "a1")
	require.True(t, found)
	require.Equal(t, math.ZeroInt(), agent.SpendPrevWindowSpent)
	require.Equal(t, math.NewInt(600), agent.RemainingWindowBudget(25))
}
