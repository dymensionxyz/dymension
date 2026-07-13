package keeper_test

import (
	"errors"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func policyA() tee.Policy {
	return tee.Policy{
		GcpRootCertPem:  "cert-a",
		PolicyValues:    `{"measurement":"aaa"}`,
		PolicyQuery:     "data.x.allow",
		PolicyStructure: "package x\nallow = true",
	}
}

func policyB() tee.Policy {
	p := policyA()
	p.PolicyValues = `{"measurement":"bbb"}`
	return p
}

func fingerprint(t *testing.T, p tee.Policy) string {
	t.Helper()
	fp, err := types.PolicyFingerprint(p)
	require.NoError(t, err)
	return fp
}

func seedAgentWithPolicy(t *testing.T, ctx sdk.Context, k *keeper.Keeper, id string, p tee.Policy) {
	t.Helper()
	require.NoError(t, k.SetAgent(ctx, types.Agent{
		Id:     id,
		Policy: p,
		Active: true,
	}))
}

func TestPolicyFingerprint_Stable(t *testing.T) {
	fp1 := fingerprint(t, policyA())
	fp2 := fingerprint(t, policyA())
	require.Equal(t, fp1, fp2)
	require.NoError(t, types.ValidateFingerprint(fp1))

	// a one-byte policy change yields a different fingerprint
	changed := policyA()
	changed.PolicyValues += " "
	require.NotEqual(t, fp1, fingerprint(t, changed))

	require.NotEqual(t, fp1, fingerprint(t, policyB()))
}

func TestRevokePolicy_AuthorityGated(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)
	fp := fingerprint(t, policyA())

	// non-authority is rejected and state is untouched
	_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(submitter(t), fp, "cve"))
	require.True(t, errors.Is(err, gerrc.ErrPermissionDenied))
	revoked, err := k.IsPolicyRevoked(ctx, fp)
	require.NoError(t, err)
	require.False(t, revoked)

	_, err = ms.UnrevokePolicy(ctx, types.NewMsgUnrevokePolicy(submitter(t), fp))
	require.True(t, errors.Is(err, gerrc.ErrPermissionDenied))

	// gov authority succeeds, idempotently
	for range 2 {
		_, err = ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fp, "cve"))
		require.NoError(t, err)
	}
	revoked, err = k.IsPolicyRevoked(ctx, fp)
	require.NoError(t, err)
	require.True(t, revoked)

	// unrevoke is idempotent too: absent fingerprint is a no-op
	for range 2 {
		_, err = ms.UnrevokePolicy(ctx, types.NewMsgUnrevokePolicy(govAuthority, fp))
		require.NoError(t, err)
	}
	revoked, err = k.IsPolicyRevoked(ctx, fp)
	require.NoError(t, err)
	require.False(t, revoked)
}

func TestRevokePolicy_ValidateBasic(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)

	for _, bad := range []string{"", "abc", "ABCDEF0000000000000000000000000000000000000000000000000000000000", "zz" + fingerprint(t, policyA())[2:]} {
		_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, bad, ""))
		require.True(t, errors.Is(err, gerrc.ErrInvalidArgument), bad)
		_, err = ms.UnrevokePolicy(ctx, types.NewMsgUnrevokePolicy(govAuthority, bad))
		require.True(t, errors.Is(err, gerrc.ErrInvalidArgument), bad)
	}

	_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy("not-bech32", fingerprint(t, policyA()), ""))
	require.True(t, errors.Is(err, gerrc.ErrInvalidArgument))
}

// TestSubmitAttestedAction_RevokedPolicy proves the core security property:
// once a policy is revoked, no agent pinning it can append actions — even with
// a token the verifier would accept — and unrevoking restores them with
// agent.Active untouched throughout.
func TestSubmitAttestedAction_RevokedPolicy(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)

	// two agents share policy A (same fleet image), a third runs policy B
	seedAgentWithPolicy(t, ctx, k, "fleet-1", policyA())
	seedAgentWithPolicy(t, ctx, k, "fleet-2", policyA())
	seedAgentWithPolicy(t, ctx, k, "other", policyB())

	_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fingerprint(t, policyA()), "bad image"))
	require.NoError(t, err)

	for _, id := range []string{"fleet-1", "fleet-2"} {
		_, err := ms.SubmitAttestedAction(ctx, validMsg(t, id, []byte("p"), 0))
		require.True(t, errors.Is(err, gerrc.ErrFailedPrecondition), id)
		require.ErrorContains(t, err, "policy revoked")

		agent, found := k.GetAgent(ctx, id)
		require.True(t, found)
		require.True(t, agent.Active, "revocation must not flip Active")
		require.Equal(t, uint64(0), agent.ActionSeq)
		_, found = k.GetActionLogEntry(ctx, id, 0)
		require.False(t, found)
	}

	// the unrelated policy keeps working
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "other", []byte("p"), 0))
	require.NoError(t, err)

	// unrevoke re-enables the fleet
	_, err = ms.UnrevokePolicy(ctx, types.NewMsgUnrevokePolicy(govAuthority, fingerprint(t, policyA())))
	require.NoError(t, err)
	for _, id := range []string{"fleet-1", "fleet-2"} {
		_, err := ms.SubmitAttestedAction(ctx, validMsg(t, id, []byte("p"), 0))
		require.NoError(t, err, id)
	}
}

func TestRegisterAgent_RevokedPolicy(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fingerprint(t, policyA()), "bad image"))
	require.NoError(t, err)

	_, err = ms.RegisterAgent(ctx, &types.MsgRegisterAgent{
		Owner:   submitter(t),
		AgentId: "new-agent",
		Policy:  policyA(),
	})
	require.True(t, errors.Is(err, gerrc.ErrFailedPrecondition))
	require.ErrorContains(t, err, "policy revoked")
	_, found := k.GetAgent(ctx, "new-agent")
	require.False(t, found)
}

func TestRevokedPoliciesQueries(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)
	seedAgentWithPolicy(t, ctx, k, "agent-a", policyA())

	fpA := fingerprint(t, policyA())
	fpB := fingerprint(t, policyB())

	// agent query surfaces the fingerprint and revoked=false pre-revocation
	agentRes, err := k.Agent(ctx, &types.QueryAgentRequest{AgentId: "agent-a"})
	require.NoError(t, err)
	require.Equal(t, fpA, agentRes.Fingerprint)
	require.False(t, agentRes.Revoked)

	for _, fp := range []string{fpA, fpB} {
		_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fp, ""))
		require.NoError(t, err)
	}

	all, err := k.RevokedPolicies(ctx, &types.QueryRevokedPoliciesRequest{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{fpA, fpB}, all.Fingerprints)
	require.IsIncreasing(t, all.Fingerprints)

	one, err := k.PolicyRevoked(ctx, &types.QueryPolicyRevokedRequest{Fingerprint: fpA})
	require.NoError(t, err)
	require.True(t, one.Revoked)

	notRevoked := "0000000000000000000000000000000000000000000000000000000000000000"
	none, err := k.PolicyRevoked(ctx, &types.QueryPolicyRevokedRequest{Fingerprint: notRevoked})
	require.NoError(t, err)
	require.False(t, none.Revoked)

	_, err = k.PolicyRevoked(ctx, &types.QueryPolicyRevokedRequest{Fingerprint: "nope"})
	require.True(t, errors.Is(err, gerrc.ErrInvalidArgument))

	agentRes, err = k.Agent(ctx, &types.QueryAgentRequest{AgentId: "agent-a"})
	require.NoError(t, err)
	require.True(t, agentRes.Revoked)
}

func TestUpdateAgentPolicy_RevokedPolicy(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)
	own := owner(t)
	seedOwnedAgent(t, ctx, k, "a1", own, validPolicyT(t, "old"))

	badPol := validPolicyT(t, "bad")
	_, err := ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fingerprint(t, badPol), "bad image"))
	require.NoError(t, err)

	_, err = ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", badPol))
	require.True(t, errors.Is(err, gerrc.ErrFailedPrecondition))
	require.ErrorContains(t, err, "policy revoked")

	agent, _ := k.GetAgent(ctx, "a1")
	require.Nil(t, agent.PendingPolicy, "revoked rotation must not be scheduled")
}

// TestSubmitAttestedAction_RevokedPendingPolicy proves promote-then-check: a
// scheduled rotation to a policy revoked after scheduling works until it
// matures, is blocked from maturity on (nothing persisted, nothing appended),
// and unrevoking restores the agent, letting the promotion land.
func TestSubmitAttestedAction_RevokedPendingPolicy(t *testing.T) {
	ctx, k, v := setup(t)
	ms := keeper.NewMsgServerImpl(*k)
	own := owner(t)
	oldPol := validPolicyT(t, "old")
	newPol := validPolicyT(t, "new")
	seedOwnedAgent(t, ctx, k, "a1", own, oldPol)

	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol))
	require.NoError(t, err)
	_, err = ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fingerprint(t, newPol), "bad image"))
	require.NoError(t, err)

	// before maturity the clean old policy is still effective
	ctx = ctx.WithBlockHeight(res.ActivationHeight - 1)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("p"), 0))
	require.NoError(t, err)
	require.Equal(t, oldPol, v.gotPolicy)

	// from maturity on, the promoted policy is revoked: blocked, no log entry,
	// no promotion persisted
	ctx = ctx.WithBlockHeight(res.ActivationHeight)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("q"), 1))
	require.True(t, errors.Is(err, gerrc.ErrFailedPrecondition))
	require.ErrorContains(t, err, "policy revoked")
	_, found := k.GetActionLogEntry(ctx, "a1", 1)
	require.False(t, found)
	agent, _ := k.GetAgent(ctx, "a1")
	require.True(t, agent.Active)
	require.Equal(t, oldPol, agent.Policy)
	require.NotNil(t, agent.PendingPolicy)
	require.Equal(t, uint64(1), agent.ActionSeq)

	// unrevoke: the pending policy promotes and verification uses it
	_, err = ms.UnrevokePolicy(ctx, types.NewMsgUnrevokePolicy(govAuthority, fingerprint(t, newPol)))
	require.NoError(t, err)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("q"), 1))
	require.NoError(t, err)
	require.Equal(t, newPol, v.gotPolicy)
	agent, _ = k.GetAgent(ctx, "a1")
	require.Equal(t, newPol, agent.Policy)
	require.Nil(t, agent.PendingPolicy)
}

// TestSubmitAttestedAction_RevokedActiveCleanPending proves the other ordering
// property: an agent whose active policy is revoked escapes via a matured
// clean rotation, because promotion happens before the denylist check.
func TestSubmitAttestedAction_RevokedActiveCleanPending(t *testing.T) {
	ctx, k, v := setup(t)
	ms := keeper.NewMsgServerImpl(*k)
	own := owner(t)
	oldPol := validPolicyT(t, "old")
	newPol := validPolicyT(t, "new")
	seedOwnedAgent(t, ctx, k, "a1", own, oldPol)

	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol))
	require.NoError(t, err)
	_, err = ms.RevokePolicy(ctx, types.NewMsgRevokePolicy(govAuthority, fingerprint(t, oldPol), "bad image"))
	require.NoError(t, err)

	// before maturity the revoked old policy is effective: blocked
	ctx = ctx.WithBlockHeight(res.ActivationHeight - 1)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("p"), 0))
	require.True(t, errors.Is(err, gerrc.ErrFailedPrecondition))

	// at maturity the clean pending policy promotes: allowed
	ctx = ctx.WithBlockHeight(res.ActivationHeight)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("p"), 0))
	require.NoError(t, err)
	require.Equal(t, newPol, v.gotPolicy)
	agent, _ := k.GetAgent(ctx, "a1")
	require.Equal(t, newPol, agent.Policy)
}

// TestAgentQuery_EffectiveFingerprint proves the Agent query reports the
// fingerprint/revoked status of the policy submit-time enforcement would use.
func TestAgentQuery_EffectiveFingerprint(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)
	own := owner(t)
	oldPol := validPolicyT(t, "old")
	newPol := validPolicyT(t, "new")
	seedOwnedAgent(t, ctx, k, "a1", own, oldPol)

	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol))
	require.NoError(t, err)

	before, err := k.Agent(ctx.WithBlockHeight(res.ActivationHeight-1), &types.QueryAgentRequest{AgentId: "a1"})
	require.NoError(t, err)
	require.Equal(t, fingerprint(t, oldPol), before.Fingerprint)

	after, err := k.Agent(ctx.WithBlockHeight(res.ActivationHeight), &types.QueryAgentRequest{AgentId: "a1"})
	require.NoError(t, err)
	require.Equal(t, fingerprint(t, newPol), after.Fingerprint)
}

func TestGenesisRoundTripRevokedPolicies(t *testing.T) {
	ctx, k, _ := setup(t)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	require.NoError(t, k.SetRevoked(ctx, fingerprint(t, policyA())))
	require.NoError(t, k.SetRevoked(ctx, fingerprint(t, policyB())))

	exported := keeper.ExportGenesis(ctx, k)
	require.NoError(t, exported.Validate())
	require.Len(t, exported.RevokedPolicies, 2)
	require.IsIncreasing(t, exported.RevokedPolicies)

	ctx2, k2, _ := setup(t)
	keeper.InitGenesis(ctx2, k2, *exported)
	require.Equal(t, exported, keeper.ExportGenesis(ctx2, k2))
}

func TestGenesisValidate_BadFingerprint(t *testing.T) {
	g := types.DefaultGenesis()
	g.RevokedPolicies = []string{"not-hex"}
	require.Error(t, g.Validate())

	g.RevokedPolicies = []string{fingerprint(t, policyA())}
	require.NoError(t, g.Validate())
}
