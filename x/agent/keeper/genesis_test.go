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
			// an unset Int round-trips as zero
			SpendLimitPerRecipientPerWindow: math.ZeroInt(),
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

func TestGenesis_RecipientSpendRoundTripDropsStaleWindows(t *testing.T) {
	ctx, k, _ := setup(t)
	ctx = ctx.WithBlockHeight(25)
	current := sdk.AccAddress("current_recipient___").String()
	stale := sdk.AccAddress("stale_recipient_____").String()

	g := types.GenesisState{
		Params: types.DefaultParams(),
		Agents: []types.Agent{{
			Id:                              "a1",
			Owner:                           owner(t),
			Active:                          true,
			SpendDenom:                      "adym",
			SpendLimitPerWindow:             math.NewInt(1000),
			SpendWindowBlocks:               10,
			SpendWindowStartHeight:          20,
			SpendWindowSpent:                math.NewInt(300),
			SpendRecipientAllowlist:         []string{current, stale},
			SpendLimitPerRecipientPerWindow: math.NewInt(400),
		}},
		RecipientSpends: []types.RecipientSpend{
			{AgentId: "a1", Recipient: current, Window: types.RecipientSpendWindow{WindowStartHeight: 20, Spent: math.NewInt(300)}},
			{AgentId: "a1", Recipient: stale, Window: types.RecipientSpendWindow{WindowStartHeight: 10, Spent: math.NewInt(400)}},
		},
	}
	require.NoError(t, g.Validate())
	keeper.InitGenesis(ctx, k, g)

	out := keeper.ExportGenesis(ctx, k)
	require.Equal(t, g.RecipientSpends[:1], out.RecipientSpends)
	require.NoError(t, out.Validate())

	require.NoError(t, keeper.InvariantRecipientSpendWithinLimit(*k)(ctx))

	ctx2, k2, _ := setup(t)
	ctx2 = ctx2.WithBlockHeight(25)
	keeper.InitGenesis(ctx2, k2, *out)
	require.Equal(t, out.RecipientSpends, keeper.ExportGenesis(ctx2, k2).RecipientSpends)
}
