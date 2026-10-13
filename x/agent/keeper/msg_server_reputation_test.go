package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// clientWithAgent returns a fresh client address that owns a live agent with
// the given id.
func clientWithAgent(t *testing.T, ctx sdk.Context, k *keeper.Keeper, agentID string) string {
	t.Helper()
	cl := client(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: agentID, Owner: cl, Active: true}))
	return cl
}

func attestedFeedbackMsg(cl string, score uint32, clientAgentID string) *types.MsgSubmitFeedback {
	msg := feedbackMsg(cl, score)
	msg.ClientAgentId = clientAgentID
	return msg
}

// requireRepMatchesRows asserts the stored aggregate equals one recomputed
// from the stored feedback rows.
func requireRepMatchesRows(t *testing.T, ctx sdk.Context, k *keeper.Keeper) {
	t.Helper()
	res, err := k.AgentFeedbacks(ctx, &types.QueryAgentFeedbacksRequest{AgentId: "agent1"})
	require.NoError(t, err)
	want := types.Reputation{AgentId: "agent1"}
	for _, fb := range res.Feedbacks {
		want.Add(fb)
	}
	rep, found := k.GetReputation(ctx, "agent1")
	if want.Count == 0 {
		require.False(t, found)
		return
	}
	require.True(t, found)
	require.Equal(t, want, rep)
}

func TestSubmitFeedback_AttestedIncrementsBothBuckets(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl := clientWithAgent(t, ctx, k, "client-agent")

	_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(cl, 9000, "client-agent"))
	require.NoError(t, err)

	fb, found := k.GetFeedback(ctx, "agent1", cl)
	require.True(t, found)
	require.True(t, fb.FromAttestedClient)

	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 1, ScoreSum: 9000, AttestedCount: 1, AttestedScoreSum: 9000}, rep)

	res, err := k.AgentReputation(ctx, &types.QueryAgentReputationRequest{AgentId: "agent1"})
	require.NoError(t, err)
	require.Equal(t, rep, res.Reputation)
}

func TestSubmitFeedback_UnattestedIncrementsTotalOnly(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	attestedCl := clientWithAgent(t, ctx, k, "client-agent")
	cl := client(t)

	_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(attestedCl, 9000, "client-agent"))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, feedbackMsg(cl, 1000))
	require.NoError(t, err)

	fb, found := k.GetFeedback(ctx, "agent1", cl)
	require.True(t, found)
	require.False(t, fb.FromAttestedClient)

	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 10000, AttestedCount: 1, AttestedScoreSum: 9000}, rep)
}

func TestSubmitFeedback_InvalidClientAgent(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl := clientWithAgent(t, ctx, k, "own-live")
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "own-inactive", Owner: cl, Active: false}))
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "not-own", Owner: owner(t), Active: true}))

	// an existing attested row from another client must stay untouched
	other := clientWithAgent(t, ctx, k, "other-agent")
	_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(other, 5000, "other-agent"))
	require.NoError(t, err)
	repBefore, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)

	for _, id := range []string{"not-own", "own-inactive", "ghost"} {
		t.Run(id, func(t *testing.T) {
			_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(cl, 9000, id))
			require.ErrorIs(t, err, types.ErrInvalidClientAgent)

			_, found := k.GetFeedback(ctx, "agent1", cl)
			require.False(t, found)
			rep, found := k.GetReputation(ctx, "agent1")
			require.True(t, found)
			require.Equal(t, repBefore, rep)
		})
	}
}

func TestSubmitFeedback_AttestedStatusFlips(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl := clientWithAgent(t, ctx, k, "client-agent")
	other := clientWithAgent(t, ctx, k, "other-agent")
	_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(other, 2000, "other-agent"))
	require.NoError(t, err)

	steps := []struct {
		name          string
		score         uint32
		clientAgentID string
		want          types.Reputation
	}{
		{"unattested first", 3000, "", types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 5000, AttestedCount: 1, AttestedScoreSum: 2000}},
		{"unattested to attested", 7000, "client-agent", types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 9000, AttestedCount: 2, AttestedScoreSum: 9000}},
		{"attested to attested", 8000, "client-agent", types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 10000, AttestedCount: 2, AttestedScoreSum: 10000}},
		{"attested to unattested", 4000, "", types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 6000, AttestedCount: 1, AttestedScoreSum: 2000}},
	}
	for _, s := range steps {
		_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(cl, s.score, s.clientAgentID))
		require.NoError(t, err, s.name)

		rep, found := k.GetReputation(ctx, "agent1")
		require.True(t, found, s.name)
		require.Equal(t, s.want, rep, s.name)
		requireRepMatchesRows(t, ctx, k)
	}
}

func TestRevokeFeedback_AttestedBuckets(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	attestedCl := clientWithAgent(t, ctx, k, "client-agent")
	cl := client(t)

	_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(attestedCl, 9000, "client-agent"))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, feedbackMsg(cl, 1000))
	require.NoError(t, err)

	// unattested revoke leaves the attested bucket alone
	_, err = ms.RevokeFeedback(ctx, types.NewMsgRevokeFeedback(cl, "agent1"))
	require.NoError(t, err)
	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 1, ScoreSum: 9000, AttestedCount: 1, AttestedScoreSum: 9000}, rep)

	// deactivating the client's agent after submit does not change which
	// bucket the frozen row is removed from
	ca, found := k.GetAgent(ctx, "client-agent")
	require.True(t, found)
	ca.Active = false
	require.NoError(t, k.SetAgent(ctx, ca))

	_, err = ms.SubmitFeedback(ctx, feedbackMsg(cl, 2000))
	require.NoError(t, err)
	_, err = ms.RevokeFeedback(ctx, types.NewMsgRevokeFeedback(attestedCl, "agent1"))
	require.NoError(t, err)
	rep, found = k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 1, ScoreSum: 2000}, rep)
	requireRepMatchesRows(t, ctx, k)

	_, err = ms.RevokeFeedback(ctx, types.NewMsgRevokeFeedback(cl, "agent1"))
	require.NoError(t, err)
	requireRepMatchesRows(t, ctx, k)
}

func TestGenesisRoundTripWithAttestedFeedback(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	attestedCl := clientWithAgent(t, ctx, k, "client-agent")

	_, err := ms.SubmitFeedback(ctx, attestedFeedbackMsg(attestedCl, 9000, "client-agent"))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, feedbackMsg(client(t), 3000))
	require.NoError(t, err)

	exported := keeper.ExportGenesis(ctx, k)
	ctx2, k2, _ := setup(t)
	keeper.InitGenesis(ctx2, k2, *exported)

	rep, found := k2.GetReputation(ctx2, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 12000, AttestedCount: 1, AttestedScoreSum: 9000}, rep)
}
