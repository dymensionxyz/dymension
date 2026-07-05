package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

// fakeVerifier stands in for the real GCP verifier: a valid GCP-signed token
// can't be produced locally, so the fake asserts the nonce the handler derived
// matches the nonce embedded in the (fake) token. token == expected nonce means
// accept; anything else rejects, exactly as the real verifier rejects a token
// whose nonce claim doesn't match.
type fakeVerifier struct {
	gotNonce  string
	gotPolicy tee.Policy
	calls     int
}

func (f *fakeVerifier) Verify(_ sdk.Context, policy tee.Policy, nonce, token string) error {
	f.calls++
	f.gotNonce = nonce
	f.gotPolicy = policy
	if token != nonce {
		return errors.New("nonce mismatch")
	}
	return nil
}

func setup(t *testing.T) (sdk.Context, *keeper.Keeper, *fakeVerifier) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	registry := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	v := &fakeVerifier{}
	// bankKeeper is nil: SubmitAttestedAction never charges fees. Registration
	// fee burning is covered by the apptesting-based suite in registry_test.go.
	k := keeper.NewKeeper(cdc, runtime.NewKVStoreService(key), v, nil)

	ctx := sdk.NewContext(cms, cmtproto.Header{Height: 10}, false, log.NewNopLogger())
	require.NoError(t, k.SetParams(ctx, types.Params{MaxActionBytes: 1024, PolicyRotationDelayBlocks: rotationDelay}))
	return ctx, k, v
}

const rotationDelay = 100

// validCertPEM generates a parseable self-signed cert so validatePolicy passes.
func validCertPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

// validPolicy returns a policy passing ValidateBasic, tagged by query so tests
// can tell old from new via fakeVerifier.gotPolicy.
func validPolicyT(t *testing.T, query string) tee.Policy {
	t.Helper()
	return tee.Policy{GcpRootCertPem: validCertPEM(t), PolicyValues: "{}", PolicyQuery: query, PolicyStructure: "package x"}
}

func owner(t *testing.T) string {
	t.Helper()
	_, _, addr := testdata.KeyTestPubAddr()
	return addr.String()
}

func seedAgent(t *testing.T, ctx sdk.Context, k *keeper.Keeper, id string, active bool) {
	t.Helper()
	require.NoError(t, k.SetAgent(ctx, types.Agent{
		Id:        id,
		Policy:    tee.Policy{},
		Active:    active,
		ActionSeq: 0,
	}))
}

func submitter(t *testing.T) string {
	t.Helper()
	_, _, addr := testdata.KeyTestPubAddr()
	return addr.String()
}

// validMsg builds a message whose token equals the nonce the handler will
// derive, so the fake verifier accepts it.
func validMsg(t *testing.T, agentID string, payload []byte, seq uint64) *types.MsgSubmitAttestedAction {
	t.Helper()
	return &types.MsgSubmitAttestedAction{
		Submitter: submitter(t),
		AgentId:   agentID,
		Payload:   payload,
		Token:     types.ActionNonce(agentID, payload, seq),
	}
}

func TestSubmitAttestedAction_Valid(t *testing.T) {
	ctx, k, v := setup(t)
	seedAgent(t, ctx, k, "agent1", true)
	ms := keeper.NewMsgServerImpl(*k)

	payload := []byte("do the thing")
	res, err := ms.SubmitAttestedAction(ctx, validMsg(t, "agent1", payload, 0))
	require.NoError(t, err)
	require.Equal(t, uint64(0), res.Seq)
	require.Equal(t, 1, v.calls)
	require.Equal(t, types.ActionNonce("agent1", payload, 0), v.gotNonce)

	entry, found := k.GetActionLogEntry(ctx, "agent1", 0)
	require.True(t, found)
	require.Equal(t, payload, entry.Payload)
	require.Equal(t, int64(10), entry.Height)

	agent, _ := k.GetAgent(ctx, "agent1")
	require.Equal(t, uint64(1), agent.ActionSeq)
}

func TestSubmitAttestedAction_TamperedPayload(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent1", true)
	ms := keeper.NewMsgServerImpl(*k)

	// Token minted for the original payload; the submitted payload differs, so
	// the re-derived nonce won't match the token's nonce.
	msg := validMsg(t, "agent1", []byte("original"), 0)
	msg.Payload = []byte("tampered")

	_, err := ms.SubmitAttestedAction(ctx, msg)
	require.Error(t, err)

	_, found := k.GetActionLogEntry(ctx, "agent1", 0)
	require.False(t, found)
	agent, _ := k.GetAgent(ctx, "agent1")
	require.Equal(t, uint64(0), agent.ActionSeq)
}

func TestSubmitAttestedAction_WrongSeq(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent1", true)
	ms := keeper.NewMsgServerImpl(*k)

	// Token minted for seq=5 but the agent is at seq=0.
	msg := validMsg(t, "agent1", []byte("p"), 5)

	_, err := ms.SubmitAttestedAction(ctx, msg)
	require.Error(t, err)

	agent, _ := k.GetAgent(ctx, "agent1")
	require.Equal(t, uint64(0), agent.ActionSeq)
}

func TestSubmitAttestedAction_ReplayAfterSuccess(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent1", true)
	ms := keeper.NewMsgServerImpl(*k)

	payload := []byte("once")
	msg := validMsg(t, "agent1", payload, 0)

	_, err := ms.SubmitAttestedAction(ctx, msg)
	require.NoError(t, err)

	// Re-submit the exact same (payload, token): action_seq is now 1, so the
	// re-derived nonce differs from the token's, and the verifier rejects.
	_, err = ms.SubmitAttestedAction(ctx, msg)
	require.Error(t, err)

	agent, _ := k.GetAgent(ctx, "agent1")
	require.Equal(t, uint64(1), agent.ActionSeq)
}

func TestSubmitAttestedAction_PayloadTooLarge(t *testing.T) {
	ctx, k, _ := setup(t)
	require.NoError(t, k.SetParams(ctx, types.Params{MaxActionBytes: 4}))
	seedAgent(t, ctx, k, "agent1", true)
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.SubmitAttestedAction(ctx, validMsg(t, "agent1", []byte("too long"), 0))
	require.Error(t, err)

	agent, _ := k.GetAgent(ctx, "agent1")
	require.Equal(t, uint64(0), agent.ActionSeq)
}

func TestSubmitAttestedAction_InactiveAgent(t *testing.T) {
	ctx, k, _ := setup(t)
	seedAgent(t, ctx, k, "agent1", false)
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.SubmitAttestedAction(ctx, validMsg(t, "agent1", []byte("p"), 0))
	require.Error(t, err)
}

func TestSubmitAttestedAction_NonexistentAgent(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.SubmitAttestedAction(ctx, validMsg(t, "ghost", []byte("p"), 0))
	require.Error(t, err)
}

// seedOwnedAgent stores an active agent with the given owner and old policy.
func seedOwnedAgent(t *testing.T, ctx sdk.Context, k *keeper.Keeper, id, own string, policy tee.Policy) {
	t.Helper()
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: id, Owner: own, Policy: policy, Active: true}))
}

func TestUpdateAgentPolicy_NonOwner(t *testing.T) {
	ctx, k, _ := setup(t)
	own := owner(t)
	seedOwnedAgent(t, ctx, k, "a1", own, validPolicyT(t, "old"))
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(owner(t), "a1", validPolicyT(t, "new")))
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestUpdateAgentPolicy_UnknownAgent(t *testing.T) {
	ctx, k, _ := setup(t)
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(owner(t), "ghost", validPolicyT(t, "new")))
	require.ErrorIs(t, err, types.ErrAgentNotFound)
}

func TestUpdateAgentPolicy_InactiveAgent(t *testing.T) {
	ctx, k, _ := setup(t)
	own := owner(t)
	require.NoError(t, k.SetAgent(ctx, types.Agent{Id: "a1", Owner: own, Policy: validPolicyT(t, "old"), Active: false}))
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", validPolicyT(t, "new")))
	require.ErrorIs(t, err, gerrc.ErrFailedPrecondition)
}

func TestUpdateAgentPolicy_MalformedPolicy(t *testing.T) {
	ctx, k, _ := setup(t)
	own := owner(t)
	seedOwnedAgent(t, ctx, k, "a1", own, validPolicyT(t, "old"))
	ms := keeper.NewMsgServerImpl(*k)

	bad := tee.Policy{GcpRootCertPem: "not a cert", PolicyQuery: "q", PolicyStructure: "s"}
	_, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", bad))
	require.ErrorIs(t, err, types.ErrInvalidPolicy)
}

func TestUpdateAgentPolicy_HappyPath(t *testing.T) {
	ctx, k, _ := setup(t)
	own := owner(t)
	oldPol := validPolicyT(t, "old")
	seedOwnedAgent(t, ctx, k, "a1", own, oldPol)
	ms := keeper.NewMsgServerImpl(*k)

	newPol := validPolicyT(t, "new")
	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol))
	require.NoError(t, err)
	wantHeight := ctx.BlockHeight() + rotationDelay
	require.Equal(t, wantHeight, res.ActivationHeight)

	agent, _ := k.GetAgent(ctx, "a1")
	require.Equal(t, oldPol, agent.Policy)
	require.NotNil(t, agent.PendingPolicy)
	require.Equal(t, newPol, *agent.PendingPolicy)
	require.Equal(t, wantHeight, agent.PendingPolicyHeight)
	require.Equal(t, uint64(0), agent.ActionSeq)
}

func TestUpdateAgentPolicy_Overwrite(t *testing.T) {
	ctx, k, _ := setup(t)
	own := owner(t)
	seedOwnedAgent(t, ctx, k, "a1", own, validPolicyT(t, "old"))
	ms := keeper.NewMsgServerImpl(*k)

	_, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", validPolicyT(t, "new1")))
	require.NoError(t, err)

	// advance a block, re-propose: pending must be replaced and timelock restarted.
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 5)
	newPol2 := validPolicyT(t, "new2")
	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol2))
	require.NoError(t, err)
	require.Equal(t, ctx.BlockHeight()+rotationDelay, res.ActivationHeight)

	agent, _ := k.GetAgent(ctx, "a1")
	require.Equal(t, newPol2, *agent.PendingPolicy)
	require.Equal(t, ctx.BlockHeight()+rotationDelay, agent.PendingPolicyHeight)
}

func TestSubmitAttestedAction_PendingPolicyBeforeActivation(t *testing.T) {
	ctx, k, v := setup(t)
	own := owner(t)
	oldPol := validPolicyT(t, "old")
	seedOwnedAgent(t, ctx, k, "a1", own, oldPol)
	ms := keeper.NewMsgServerImpl(*k)

	newPol := validPolicyT(t, "new")
	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol))
	require.NoError(t, err)

	// one block before activation: verification uses the old policy.
	ctx = ctx.WithBlockHeight(res.ActivationHeight - 1)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("p"), 0))
	require.NoError(t, err)
	require.Equal(t, oldPol, v.gotPolicy)

	agent, _ := k.GetAgent(ctx, "a1")
	require.Equal(t, oldPol, agent.Policy)
	require.NotNil(t, agent.PendingPolicy)
}

func TestSubmitAttestedAction_PendingPolicyAtActivation(t *testing.T) {
	ctx, k, v := setup(t)
	own := owner(t)
	seedOwnedAgent(t, ctx, k, "a1", own, validPolicyT(t, "old"))
	ms := keeper.NewMsgServerImpl(*k)

	newPol := validPolicyT(t, "new")
	res, err := ms.UpdateAgentPolicy(ctx, types.NewMsgUpdateAgentPolicy(own, "a1", newPol))
	require.NoError(t, err)

	// at activation height: verification uses the new policy and it is persisted.
	ctx = ctx.WithBlockHeight(res.ActivationHeight)
	_, err = ms.SubmitAttestedAction(ctx, validMsg(t, "a1", []byte("p"), 0))
	require.NoError(t, err)
	require.Equal(t, newPol, v.gotPolicy)

	agent, _ := k.GetAgent(ctx, "a1")
	require.Equal(t, newPol, agent.Policy)
	require.Nil(t, agent.PendingPolicy)
	require.Equal(t, int64(0), agent.PendingPolicyHeight)
}

func TestAgent_EffectivePolicy(t *testing.T) {
	oldPol := validPolicyT(t, "old")
	newPol := validPolicyT(t, "new")

	// no pending: always the active policy.
	a := types.Agent{Policy: oldPol}
	require.Equal(t, oldPol, a.EffectivePolicy(1000))

	a = types.Agent{Policy: oldPol, PendingPolicy: &newPol, PendingPolicyHeight: 100}
	require.Equal(t, oldPol, a.EffectivePolicy(99))
	require.Equal(t, newPol, a.EffectivePolicy(100))
	require.Equal(t, newPol, a.EffectivePolicy(101))
}
