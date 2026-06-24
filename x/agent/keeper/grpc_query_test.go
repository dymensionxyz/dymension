package keeper_test

import (
	"crypto/sha256"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// submitAction appends one attested action to the agent's log via the real
// msg-server path (the fake verifier accepts a token equal to the derived
// nonce), so query tests read genuinely-appended entries.
func submitAction(t *testing.T, ctx sdk.Context, k *keeper.Keeper, agentID string, payload []byte, seq uint64) {
	t.Helper()
	ms := keeper.NewMsgServerImpl(*k)
	_, err := ms.SubmitAttestedAction(ctx, &types.MsgSubmitAttestedAction{
		Submitter: submitter(t),
		AgentId:   agentID,
		Payload:   payload,
		Token:     types.ActionNonce(agentID, payload, seq),
	})
	require.NoError(t, err)
}

func TestParamsQuery(t *testing.T) {
	ctx, k, _ := setup(t)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	res, err := k.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), res.Params)
	require.Equal(t, uint64(types.DefaultMaxActionBytes), res.Params.MaxActionBytes)
}

func TestAgentQuery(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent-1", true)

	res, err := k.Agent(ctx, &types.QueryAgentRequest{Id: "agent-1"})
	require.NoError(t, err)
	require.Equal(t, "agent-1", res.Agent.Id)
	require.True(t, res.Agent.Active)

	_, err = k.Agent(ctx, &types.QueryAgentRequest{Id: "missing"})
	require.Equal(t, codes.NotFound, status.Code(err))

	_, err = k.Agent(ctx, &types.QueryAgentRequest{Id: ""})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestAgentsQueryPagination(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent-1", true)
	seedAgent(t, ctx, k, "agent-2", true)
	seedAgent(t, ctx, k, "agent-3", true)

	all, err := k.Agents(ctx, &types.QueryAgentsRequest{})
	require.NoError(t, err)
	require.Len(t, all.Agents, 3)

	first, err := k.Agents(ctx, &types.QueryAgentsRequest{
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	require.NoError(t, err)
	require.Len(t, first.Agents, 2)
	require.NotNil(t, first.Pagination.NextKey)
	require.Equal(t, uint64(3), first.Pagination.Total)

	second, err := k.Agents(ctx, &types.QueryAgentsRequest{
		Pagination: &query.PageRequest{Key: first.Pagination.NextKey},
	})
	require.NoError(t, err)
	require.Len(t, second.Agents, 1)
}

func TestAgentActionsQueryScopedAndPaginated(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent-1", true)
	seedAgent(t, ctx, k, "agent-2", true)
	submitAction(t, ctx, k, "agent-1", []byte("a-0"), 0)
	submitAction(t, ctx, k, "agent-1", []byte("a-1"), 1)
	submitAction(t, ctx, k, "agent-1", []byte("a-2"), 2)
	submitAction(t, ctx, k, "agent-2", []byte("b-0"), 0)

	res, err := k.AgentActions(ctx, &types.QueryAgentActionsRequest{AgentId: "agent-1"})
	require.NoError(t, err)
	require.Len(t, res.Actions, 3)
	for i, a := range res.Actions {
		require.Equal(t, "agent-1", a.AgentId)
		require.Equal(t, uint64(i), a.Seq)
	}

	// pagination is scoped to the agent
	page, err := k.AgentActions(ctx, &types.QueryAgentActionsRequest{
		AgentId:    "agent-1",
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	require.NoError(t, err)
	require.Len(t, page.Actions, 2)
	require.Equal(t, uint64(3), page.Pagination.Total)

	// other agent's log is separate
	other, err := k.AgentActions(ctx, &types.QueryAgentActionsRequest{AgentId: "agent-2"})
	require.NoError(t, err)
	require.Len(t, other.Actions, 1)
	require.Equal(t, "agent-2", other.Actions[0].AgentId)

	_, err = k.AgentActions(ctx, &types.QueryAgentActionsRequest{AgentId: ""})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestAgentActionQuery(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent-1", true)
	submitAction(t, ctx, k, "agent-1", []byte("first"), 0)
	submitAction(t, ctx, k, "agent-1", []byte("second"), 1)

	res, err := k.AgentAction(ctx, &types.QueryAgentActionRequest{AgentId: "agent-1", Seq: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(1), res.Action.Seq)
	require.Equal(t, "agent-1", res.Action.AgentId)
	want := sha256.Sum256(res.Action.Payload)
	require.Equal(t, want[:], res.Action.PayloadHash)

	_, err = k.AgentAction(ctx, &types.QueryAgentActionRequest{AgentId: "agent-1", Seq: 99})
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestGenesisRoundTrip(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent-1", true)
	seedAgent(t, ctx, k, "agent-2", true)
	submitAction(t, ctx, k, "agent-1", []byte("x"), 0)
	submitAction(t, ctx, k, "agent-1", []byte("y"), 1)

	exported := keeper.ExportGenesis(ctx, k)
	require.NoError(t, exported.Validate())
	require.Len(t, exported.Agents, 2)
	require.Len(t, exported.ActionLog, 2)

	ctx2, k2, _ := setup(t)
	keeper.InitGenesis(ctx2, k2, *exported)

	reexported := keeper.ExportGenesis(ctx2, k2)
	require.Equal(t, exported, reexported)
}
