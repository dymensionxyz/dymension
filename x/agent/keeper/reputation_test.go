package keeper_test

import (
	"errors"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

// setupReputation seeds an agent with an owner and action history under params
// with a zero feedback fee (bankKeeper is nil in the unit setup; fee charging
// is covered by ReputationFeeTestSuite).
func setupReputation(t *testing.T) (sdk.Context, *keeper.Keeper, types.MsgServer, string) {
	t.Helper()
	ctx, k, _ := setup(t)
	params := types.DefaultParams()
	params.FeedbackFee = sdk.NewCoin(params.FeedbackFee.Denom, math.ZeroInt())
	require.NoError(t, k.SetParams(ctx, params))
	own := owner(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{
		Id:        "agent1",
		Owner:     own,
		Policy:    tee.Policy{},
		Active:    true,
		ActionSeq: 3,
	}))
	return ctx, k, keeper.NewMsgServerImpl(*k), own
}

func client(t *testing.T) string {
	t.Helper()
	_, _, addr := testdata.KeyTestPubAddr()
	return addr.String()
}

func feedbackMsg(cl string, score uint32) *types.MsgSubmitFeedback {
	return types.NewMsgSubmitFeedback(cl, "agent1", score, "liveness", "", 0)
}

func TestSubmitFeedback_First(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl := client(t)

	_, err := ms.SubmitFeedback(ctx, types.NewMsgSubmitFeedback(cl, "agent1", 9000, "liveness", "quality", 2))
	require.NoError(t, err)

	fb, found := k.GetFeedback(ctx, "agent1", cl)
	require.True(t, found)
	require.Equal(t, uint32(9000), fb.Score)
	require.Equal(t, "liveness", fb.Tag1)
	require.Equal(t, "quality", fb.Tag2)
	require.Equal(t, uint64(2), fb.EvidenceSeq)
	require.Equal(t, int64(10), fb.Height)

	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 1, ScoreSum: 9000}, rep)

	res, err := k.AgentReputation(ctx, &types.QueryAgentReputationRequest{AgentId: "agent1"})
	require.NoError(t, err)
	require.Equal(t, uint32(9000), res.AverageScore)
}

func TestSubmitFeedback_SecondClient(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)

	_, err := ms.SubmitFeedback(ctx, feedbackMsg(client(t), 9000))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, feedbackMsg(client(t), 8000))
	require.NoError(t, err)

	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, uint64(2), rep.Count)
	require.Equal(t, uint64(17000), rep.ScoreSum)

	res, err := k.AgentReputation(ctx, &types.QueryAgentReputationRequest{AgentId: "agent1"})
	require.NoError(t, err)
	require.Equal(t, uint32(8500), res.AverageScore)
}

func TestSubmitFeedback_ResubmissionOverwrites(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl := client(t)

	_, err := ms.SubmitFeedback(ctx, types.NewMsgSubmitFeedback(cl, "agent1", 9000, "liveness", "", 0))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, types.NewMsgSubmitFeedback(cl, "agent1", 4000, "reliability", "quality", 1))
	require.NoError(t, err)

	fb, found := k.GetFeedback(ctx, "agent1", cl)
	require.True(t, found)
	require.Equal(t, uint32(4000), fb.Score)
	require.Equal(t, "reliability", fb.Tag1)
	require.Equal(t, "quality", fb.Tag2)
	require.Equal(t, uint64(1), fb.EvidenceSeq)

	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 1, ScoreSum: 4000}, rep)
}

func TestRevokeFeedback(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl1 := client(t)
	cl2 := client(t)

	_, err := ms.SubmitFeedback(ctx, feedbackMsg(cl1, 9000))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, feedbackMsg(cl2, 7000))
	require.NoError(t, err)

	_, err = ms.RevokeFeedback(ctx, types.NewMsgRevokeFeedback(cl1, "agent1"))
	require.NoError(t, err)

	_, found := k.GetFeedback(ctx, "agent1", cl1)
	require.False(t, found)
	rep, found := k.GetReputation(ctx, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 1, ScoreSum: 7000}, rep)

	// revoking the last feedback deletes the aggregate entirely
	_, err = ms.RevokeFeedback(ctx, types.NewMsgRevokeFeedback(cl2, "agent1"))
	require.NoError(t, err)
	_, found = k.GetReputation(ctx, "agent1")
	require.False(t, found)

	res, err := k.AgentReputation(ctx, &types.QueryAgentReputationRequest{AgentId: "agent1"})
	require.NoError(t, err)
	require.Equal(t, types.Reputation{}, res.Reputation)
	require.Equal(t, uint32(0), res.AverageScore)
}

func TestRevokeFeedback_NotFound(t *testing.T) {
	ctx, _, ms, _ := setupReputation(t)

	_, err := ms.RevokeFeedback(ctx, types.NewMsgRevokeFeedback(client(t), "agent1"))
	require.ErrorIs(t, err, types.ErrFeedbackNotFound)
}

func TestSubmitFeedback_Errors(t *testing.T) {
	ctx, k, ms, own := setupReputation(t)

	cases := []struct {
		name string
		msg  *types.MsgSubmitFeedback
		err  error
	}{
		{"score too high", types.NewMsgSubmitFeedback(client(t), "agent1", 10001, "liveness", "", 0), types.ErrInvalidScore},
		{"empty tag1", types.NewMsgSubmitFeedback(client(t), "agent1", 100, "", "", 0), types.ErrInvalidTag},
		{"control char tag", types.NewMsgSubmitFeedback(client(t), "agent1", 100, "live\nness", "", 0), types.ErrInvalidTag},
		{"oversized tag1", types.NewMsgSubmitFeedback(client(t), "agent1", 100, strings.Repeat("a", 33), "", 0), types.ErrInvalidTag},
		{"oversized tag2", types.NewMsgSubmitFeedback(client(t), "agent1", 100, "liveness", strings.Repeat("a", 33), 0), types.ErrInvalidTag},
		{"evidence at action seq", types.NewMsgSubmitFeedback(client(t), "agent1", 100, "liveness", "", 3), types.ErrInvalidEvidence},
		{"unknown agent", types.NewMsgSubmitFeedback(client(t), "ghost", 100, "liveness", "", 0), types.ErrAgentNotFound},
		{"self feedback", types.NewMsgSubmitFeedback(own, "agent1", 100, "liveness", "", 0), types.ErrSelfFeedback},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.SubmitFeedback(ctx, tc.msg)
			require.ErrorIs(t, err, tc.err)

			_, found := k.GetFeedback(ctx, tc.msg.AgentId, tc.msg.Client)
			require.False(t, found)
			_, found = k.GetReputation(ctx, tc.msg.AgentId)
			require.False(t, found)
		})
	}
}

func TestSubmitFeedback_ZeroActionHistory(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "fresh", Owner: owner(t), Active: true, ActionSeq: 0}))

	_, err := ms.SubmitFeedback(ctx, types.NewMsgSubmitFeedback(client(t), "fresh", 100, "liveness", "", 0))
	require.ErrorIs(t, err, types.ErrInvalidEvidence)
}

func TestAgentFeedbackQuery(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	cl := client(t)

	_, err := ms.SubmitFeedback(ctx, feedbackMsg(cl, 5000))
	require.NoError(t, err)

	res, err := k.AgentFeedback(ctx, &types.QueryAgentFeedbackRequest{AgentId: "agent1", Client: cl})
	require.NoError(t, err)
	require.Equal(t, uint32(5000), res.Feedback.Score)

	_, err = k.AgentFeedback(ctx, &types.QueryAgentFeedbackRequest{AgentId: "agent1", Client: client(t)})
	require.ErrorIs(t, err, types.ErrFeedbackNotFound)

	_, err = k.AgentFeedback(ctx, &types.QueryAgentFeedbackRequest{AgentId: "", Client: cl})
	require.True(t, errors.Is(err, gerrc.ErrInvalidArgument))
	_, err = k.AgentFeedback(ctx, &types.QueryAgentFeedbackRequest{AgentId: "agent1", Client: ""})
	require.True(t, errors.Is(err, gerrc.ErrInvalidArgument))
}

func TestAgentFeedbacksQueryScopedAndPaginated(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "agent2", Owner: owner(t), Active: true, ActionSeq: 1}))

	for i := 0; i < 3; i++ {
		_, err := ms.SubmitFeedback(ctx, feedbackMsg(client(t), uint32(1000*(i+1)))) //nolint:gosec
		require.NoError(t, err)
	}
	_, err := ms.SubmitFeedback(ctx, types.NewMsgSubmitFeedback(client(t), "agent2", 100, "liveness", "", 0))
	require.NoError(t, err)

	all, err := k.AgentFeedbacks(ctx, &types.QueryAgentFeedbacksRequest{AgentId: "agent1"})
	require.NoError(t, err)
	require.Len(t, all.Feedbacks, 3)
	for _, f := range all.Feedbacks {
		require.Equal(t, "agent1", f.AgentId)
	}

	page, err := k.AgentFeedbacks(ctx, &types.QueryAgentFeedbacksRequest{
		AgentId:    "agent1",
		Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
	})
	require.NoError(t, err)
	require.Len(t, page.Feedbacks, 2)
	require.NotNil(t, page.Pagination.NextKey)
	require.Equal(t, uint64(3), page.Pagination.Total)

	rest, err := k.AgentFeedbacks(ctx, &types.QueryAgentFeedbacksRequest{
		AgentId:    "agent1",
		Pagination: &query.PageRequest{Key: page.Pagination.NextKey},
	})
	require.NoError(t, err)
	require.Len(t, rest.Feedbacks, 1)

	_, err = k.AgentFeedbacks(ctx, &types.QueryAgentFeedbacksRequest{AgentId: ""})
	require.True(t, errors.Is(err, gerrc.ErrInvalidArgument))
}

func TestGenesisRoundTripWithFeedback(t *testing.T) {
	ctx, k, ms, _ := setupReputation(t)

	_, err := ms.SubmitFeedback(ctx, feedbackMsg(client(t), 9000))
	require.NoError(t, err)
	_, err = ms.SubmitFeedback(ctx, feedbackMsg(client(t), 3000))
	require.NoError(t, err)

	exported := keeper.ExportGenesis(ctx, k)
	require.NoError(t, exported.Validate())
	require.Len(t, exported.Feedbacks, 2)

	ctx2, k2, _ := setup(t)
	keeper.InitGenesis(ctx2, k2, *exported)
	require.Equal(t, exported, keeper.ExportGenesis(ctx2, k2))

	// aggregates are recomputed from the imported feedback set
	rep, found := k2.GetReputation(ctx2, "agent1")
	require.True(t, found)
	require.Equal(t, types.Reputation{AgentId: "agent1", Count: 2, ScoreSum: 12000}, rep)
}

func TestGenesisValidate_Feedbacks(t *testing.T) {
	fb := func(agentID, cl string, score uint32) types.Feedback {
		return types.Feedback{AgentId: agentID, Client: cl, Score: score, Tag1: "liveness"}
	}

	g := types.GenesisState{
		Params:    types.DefaultParams(),
		Feedbacks: []types.Feedback{fb("a1", "c1", 100), fb("a1", "c1", 200)},
	}
	require.ErrorContains(t, g.Validate(), "duplicate feedback")

	g = types.GenesisState{
		Params:    types.DefaultParams(),
		Feedbacks: []types.Feedback{fb("a1", "c1", 10001)},
	}
	require.ErrorContains(t, g.Validate(), "exceeds max")

	g = types.GenesisState{
		Params:    types.DefaultParams(),
		Feedbacks: []types.Feedback{fb("a1", "c1", 100), fb("a1", "c2", 100), fb("a2", "c1", 100)},
	}
	require.NoError(t, g.Validate())
}

// ReputationFeeTestSuite covers the bank interactions (fee charge + burn),
// which need the full app.
type ReputationFeeTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer types.MsgServer
}

func TestReputationFeeTestSuite(t *testing.T) {
	suite.Run(t, new(ReputationFeeTestSuite))
}

func (s *ReputationFeeTestSuite) SetupTest() {
	app := apptesting.Setup(s.T())
	s.App = app
	s.Ctx = app.NewContext(false)
	s.msgServer = keeper.NewMsgServerImpl(*s.App.AgentKeeper)
}

func (s *ReputationFeeTestSuite) TestSubmitFeedback_FeeBurned() {
	k := s.App.AgentKeeper
	params, err := k.GetParams(s.Ctx)
	s.Require().NoError(err)
	fee := params.FeedbackFee
	s.Require().False(fee.IsZero())

	_, _, own := testdata.KeyTestPubAddr()
	s.Require().NoError(k.SetAgent(s.Ctx, types.Agent{Id: "agent1", Owner: own.String(), Active: true, ActionSeq: 1}))

	_, _, cl := testdata.KeyTestPubAddr()
	s.FundAcc(cl, sdk.NewCoins(sdk.NewCoin(fee.Denom, fee.Amount.MulRaw(10))))
	balBefore := s.App.BankKeeper.GetBalance(s.Ctx, cl, fee.Denom)
	supplyBefore := s.App.BankKeeper.GetSupply(s.Ctx, fee.Denom)

	_, err = s.msgServer.SubmitFeedback(s.Ctx, types.NewMsgSubmitFeedback(cl.String(), "agent1", 9000, "liveness", "", 0))
	s.Require().NoError(err)

	balAfter := s.App.BankKeeper.GetBalance(s.Ctx, cl, fee.Denom)
	s.Require().Equal(balBefore.Amount.Sub(fee.Amount), balAfter.Amount)
	supplyAfter := s.App.BankKeeper.GetSupply(s.Ctx, fee.Denom)
	s.Require().Equal(supplyBefore.Amount.Sub(fee.Amount), supplyAfter.Amount)

	// revoke charges nothing
	balBeforeRevoke := s.App.BankKeeper.GetBalance(s.Ctx, cl, fee.Denom)
	_, err = s.msgServer.RevokeFeedback(s.Ctx, types.NewMsgRevokeFeedback(cl.String(), "agent1"))
	s.Require().NoError(err)
	s.Require().Equal(balBeforeRevoke, s.App.BankKeeper.GetBalance(s.Ctx, cl, fee.Denom))
}

func (s *ReputationFeeTestSuite) TestSubmitFeedback_InsufficientFee() {
	k := s.App.AgentKeeper
	_, _, own := testdata.KeyTestPubAddr()
	s.Require().NoError(k.SetAgent(s.Ctx, types.Agent{Id: "agent1", Owner: own.String(), Active: true, ActionSeq: 1}))

	_, _, cl := testdata.KeyTestPubAddr()
	_, err := s.msgServer.SubmitFeedback(s.Ctx, types.NewMsgSubmitFeedback(cl.String(), "agent1", 9000, "liveness", "", 0))
	s.Require().ErrorIs(err, types.ErrFeedbackFeePayment)

	_, found := k.GetFeedback(s.Ctx, "agent1", cl.String())
	s.Require().False(found)
}
