package keeper_test

import (
	"errors"
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
	k := keeper.NewKeeper(cdc, runtime.NewKVStoreService(key), v)

	ctx := sdk.NewContext(cms, cmtproto.Header{Height: 10}, false, log.NewNopLogger())
	require.NoError(t, k.SetParams(ctx, types.Params{MaxActionBytes: 1024}))
	return ctx, k, v
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
