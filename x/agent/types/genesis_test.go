package types_test

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
	_, _, recipient := testdata.KeyTestPubAddr()
	_, _, otherRecipient := testdata.KeyTestPubAddr()
	valid := func() types.GenesisState {
		return types.GenesisState{
			Params: types.DefaultParams(),
			Agents: []types.Agent{spendingAgent()},
			Escrows: []types.AgentEscrow{
				{AgentId: "a1", Balance: sdk.NewCoins(sdk.NewInt64Coin("adym", 5))},
			},
			InvoiceSettlements: []types.GenesisInvoiceSettlement{{
				Recipient:  recipient.String(),
				InvoiceRef: []byte("inv\x00-1"),
				Settlement: types.InvoiceSettlement{AgentId: "a1", Amount: sdk.NewInt64Coin("adym", 5), Height: 3, Seq: 1},
			}},
		}
	}

	require.NoError(t, valid().Validate())
	require.NoError(t, types.DefaultGenesis().Validate())

	// settlements are payer-scoped: another agent may hold the same recipient/ref
	otherPayer := valid()
	a2 := spendingAgent()
	a2.Id = "a2"
	otherPayer.Agents = append(otherPayer.Agents, a2)
	dup := otherPayer.InvoiceSettlements[0]
	dup.Settlement.AgentId = "a2"
	otherPayer.InvoiceSettlements = append(otherPayer.InvoiceSettlements, dup)
	require.NoError(t, otherPayer.Validate())

	// a settlement recorded before invoice_ref_max_bytes was lowered still imports
	lowered := valid()
	lowered.Params.InvoiceRefMaxBytes = 1
	require.NoError(t, lowered.Validate())

	tests := []struct {
		name   string
		mutate func(*types.GenesisState)
	}{
		{"enabled agent with zero window blocks", func(g *types.GenesisState) { g.Agents[0].SpendWindowBlocks = 0 }},
		{"enabled agent with nil limit", func(g *types.GenesisState) { g.Agents[0].SpendLimitPerWindow = math.Int{} }},
		{"enabled agent with invalid denom", func(g *types.GenesisState) { g.Agents[0].SpendDenom = "1" }},
		{"window spent above limit", func(g *types.GenesisState) { g.Agents[0].SpendWindowSpent = math.NewInt(101) }},
		{"negative window spent", func(g *types.GenesisState) { g.Agents[0].SpendWindowSpent = math.NewInt(-1) }},
		{"window start not bucket-aligned", func(g *types.GenesisState) { g.Agents[0].SpendWindowStartHeight = 15 }},
		{"invalid spend recipient", func(g *types.GenesisState) { g.Agents[0].SpendRecipientAllowlist = []string{"invalid"} }},
		{"duplicate spend recipient", func(g *types.GenesisState) {
			g.Agents[0].SpendRecipientAllowlist = []string{recipient.String(), recipient.String()}
		}},
		{"equivalent case spend recipient", func(g *types.GenesisState) {
			g.Agents[0].SpendRecipientAllowlist = []string{recipient.String(), strings.ToUpper(recipient.String())}
		}},
		{"spend recipient allowlist above max", func(g *types.GenesisState) {
			g.Params.SpendRecipientAllowlistMax = 1
			g.Agents[0].SpendRecipientAllowlist = []string{recipient.String(), otherRecipient.String()}
		}},
		{"duplicate agent entry", func(g *types.GenesisState) { g.Agents = append(g.Agents, spendingAgent()) }},
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
		{"invoice settlement with invalid recipient", func(g *types.GenesisState) { g.InvoiceSettlements[0].Recipient = "invalid" }},
		{"invoice settlement with empty ref", func(g *types.GenesisState) { g.InvoiceSettlements[0].InvoiceRef = nil }},
		{"invoice settlement with non-canonical recipient", func(g *types.GenesisState) {
			g.InvoiceSettlements[0].Recipient = strings.ToUpper(recipient.String())
		}},
		{"invoice settlement for unknown agent", func(g *types.GenesisState) { g.InvoiceSettlements[0].Settlement.AgentId = "ghost" }},
		{"invoice settlement with invalid amount", func(g *types.GenesisState) {
			g.InvoiceSettlements[0].Settlement.Amount = sdk.Coin{Denom: "adym", Amount: math.NewInt(-1)}
		}},
		{"duplicate invoice settlement", func(g *types.GenesisState) {
			g.InvoiceSettlements = append(g.InvoiceSettlements, g.InvoiceSettlements[0])
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
