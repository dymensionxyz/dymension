package keeper_test

import (
	"crypto/sha256"
	"testing"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func setup(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(types.StoreKey)
	logger := log.NewNopLogger()
	stateStore := integration.CreateMultiStore(keys, logger)
	cdc := params.MakeEncodingConfig().Codec

	k := keeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[types.StoreKey]))
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	return k, ctx
}

func seedAgent(t *testing.T, k *keeper.Keeper, ctx sdk.Context, id string, actions int) {
	t.Helper()
	require.NoError(t, k.SetAgent(ctx, types.Agent{
		Id:        id,
		Owner:     "dym1owner",
		Active:    true,
		ActionSeq: uint64(actions),
	}))
	for seq := 0; seq < actions; seq++ {
		payload := []byte(id + "-payload-" + string(rune('a'+seq)))
		h := sha256.Sum256(payload)
		require.NoError(t, k.SetActionLogEntry(ctx, types.ActionLogEntry{
			AgentId:     id,
			Seq:         uint64(seq),
			Payload:     payload,
			PayloadHash: h[:],
			Height:      int64(seq),
			Time:        time.Unix(int64(seq), 0).UTC(),
		}))
	}
}

func TestParamsQuery(t *testing.T) {
	k, ctx := setup(t)
	res, err := k.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), res.Params)
	require.Equal(t, uint64(types.DefaultMaxActionBytes), res.Params.MaxActionBytes)
}

func TestAgentQuery(t *testing.T) {
	k, ctx := setup(t)
	seedAgent(t, k, ctx, "agent-1", 0)

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
	k, ctx := setup(t)
	seedAgent(t, k, ctx, "agent-1", 0)
	seedAgent(t, k, ctx, "agent-2", 0)
	seedAgent(t, k, ctx, "agent-3", 0)

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
	k, ctx := setup(t)
	seedAgent(t, k, ctx, "agent-1", 3)
	seedAgent(t, k, ctx, "agent-2", 1)

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
	k, ctx := setup(t)
	seedAgent(t, k, ctx, "agent-1", 2)

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
	k, ctx := setup(t)
	seedAgent(t, k, ctx, "agent-1", 2)
	seedAgent(t, k, ctx, "agent-2", 0)

	exported := keeper.ExportGenesis(ctx, k)
	require.NoError(t, exported.Validate())
	require.Len(t, exported.Agents, 2)
	require.Len(t, exported.ActionLog, 2)

	k2, ctx2 := setup(t)
	keeper.InitGenesis(ctx2, k2, *exported)

	reexported := keeper.ExportGenesis(ctx2, k2)
	require.Equal(t, exported, reexported)
}
