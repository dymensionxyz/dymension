package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestGetAgentOwner(t *testing.T) {
	ctx, k, _ := setup(t)

	const (
		agentID    = "agent-1"
		agentOwner = "dym1owner"
	)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: agentID, Owner: agentOwner}))

	gotOwner, found := k.GetAgentOwner(ctx, agentID)
	require.True(t, found)
	require.Equal(t, agentOwner, gotOwner)

	gotOwner, found = k.GetAgentOwner(ctx, "missing-agent")
	require.False(t, found)
	require.Empty(t, gotOwner)
}
