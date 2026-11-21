package types_test

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestAgentSpendRecipientAllowed_EquivalentAddressEncoding(t *testing.T) {
	_, _, allowed := testdata.KeyTestPubAddr()
	_, _, denied := testdata.KeyTestPubAddr()
	agent := types.Agent{SpendRecipientAllowlist: []string{allowed.String()}}

	require.True(t, agent.SpendRecipientAllowed(strings.ToUpper(allowed.String())))
	require.False(t, agent.SpendRecipientAllowed(denied.String()))
	require.True(t, (types.Agent{}).SpendRecipientAllowed(denied.String()))
}

// TestAgentSpendWindow mirrors the eIBC on-demand LP bucket tests: the window
// is an absolute-aligned tumbling bucket that fully resets on rollover.
func TestAgentSpendWindow(t *testing.T) {
	a := types.Agent{
		SpendDenom:          "adym",
		SpendLimitPerWindow: math.NewInt(100),
		SpendWindowBlocks:   10,
	}

	// nil window-spent decodes as zero
	require.True(t, a.SpendAllows(15, math.NewInt(100)))
	require.False(t, a.SpendAllows(15, math.NewInt(101)))

	a.RecordSpend(15, math.NewInt(60))
	require.Equal(t, uint64(10), a.SpendWindowStartHeight)
	require.True(t, a.SpendAllows(19, math.NewInt(40)))
	require.False(t, a.SpendAllows(19, math.NewInt(41)))

	// rollover: capacity resets even though the prior window was exhausted
	a.RecordSpend(19, math.NewInt(40))
	require.False(t, a.SpendAllows(19, math.NewInt(1)))
	require.True(t, a.SpendAllows(20, math.NewInt(100)))

	a.RecordSpend(20, math.NewInt(100))
	require.Equal(t, uint64(20), a.SpendWindowStartHeight)
	require.Equal(t, math.NewInt(100), a.SpendWindowSpent)
	require.False(t, a.SpendAllows(29, math.NewInt(1)))
	require.True(t, a.SpendAllows(30, math.NewInt(100)))
}

func TestAgentSpendDisabled(t *testing.T) {
	a := types.Agent{}
	require.False(t, a.SpendEnabled())
	require.Equal(t, math.ZeroInt(), a.RemainingWindowBudget(100))
	require.False(t, a.SpendAllows(100, math.NewInt(1)))
}

// TestAgentRecipientSpendWindow checks the per-recipient window shares the
// aggregate's SpendBucket alignment and resets on rollover.
func TestAgentRecipientSpendWindow(t *testing.T) {
	_, _, recipient := testdata.KeyTestPubAddr()
	a := types.Agent{
		SpendDenom:                      "adym",
		SpendLimitPerWindow:             math.NewInt(100),
		SpendWindowBlocks:               10,
		SpendRecipientAllowlist:         []string{recipient.String()},
		SpendLimitPerRecipientPerWindow: math.NewInt(40),
	}
	require.True(t, a.RecipientSpendEnforced())

	// a missing record is an empty window
	var rec types.RecipientSpendWindow
	require.True(t, a.RecipientSpendAllows(rec, 15, math.NewInt(40)))
	require.False(t, a.RecipientSpendAllows(rec, 15, math.NewInt(41)))

	rec = a.RecordRecipientSpend(rec, 15, math.NewInt(30))
	require.Equal(t, a.SpendBucket(15), rec.WindowStartHeight)
	require.True(t, a.RecipientSpendAllows(rec, 19, math.NewInt(10)))
	require.False(t, a.RecipientSpendAllows(rec, 19, math.NewInt(11)))

	// the next bucket ignores the stale total and overwrites it on write
	require.True(t, a.RecipientSpendAllows(rec, 20, math.NewInt(40)))
	rec = a.RecordRecipientSpend(rec, 20, math.NewInt(5))
	require.Equal(t, uint64(20), rec.WindowStartHeight)
	require.Equal(t, math.NewInt(5), rec.Spent)
}

func TestAgentRecipientSpendEnforced(t *testing.T) {
	_, _, recipient := testdata.KeyTestPubAddr()
	require.False(t, (types.Agent{}).RecipientSpendEnforced())
	require.False(t, (types.Agent{SpendLimitPerRecipientPerWindow: math.ZeroInt(), SpendRecipientAllowlist: []string{recipient.String()}}).RecipientSpendEnforced())
	require.False(t, (types.Agent{SpendLimitPerRecipientPerWindow: math.NewInt(1)}).RecipientSpendEnforced())
}
