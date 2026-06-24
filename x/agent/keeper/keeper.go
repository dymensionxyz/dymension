package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

type Keeper struct {
	params collections.Item[types.Params]

	// agents keyed by agent id
	agents collections.Map[string, types.Agent]

	// action log keyed by (agent_id, seq)
	actionLog collections.Map[collections.Pair[string, uint64], types.ActionLogEntry]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service store.KVStoreService,
) *Keeper {
	sb := collections.NewSchemaBuilder(service)

	params := collections.NewItem(sb,
		collections.NewPrefix(types.KeyParams), types.KeyParams,
		collcompat.ProtoValue[types.Params](cdc))

	agents := collections.NewMap(sb,
		collections.NewPrefix(types.KeyAgents), types.KeyAgents,
		collections.StringKey, collcompat.ProtoValue[types.Agent](cdc))

	actionLog := collections.NewMap(sb,
		collections.NewPrefix(types.KeyActionLog), types.KeyActionLog,
		collections.PairKeyCodec(collections.StringKey, collections.Uint64Key),
		collcompat.ProtoValue[types.ActionLogEntry](cdc))

	return &Keeper{
		params:    params,
		agents:    agents,
		actionLog: actionLog,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	return k.params.Set(ctx, p)
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) SetAgent(ctx sdk.Context, a types.Agent) error {
	return k.agents.Set(ctx, a.Id, a)
}

func (k Keeper) GetAgent(ctx sdk.Context, id string) (types.Agent, error) {
	return k.agents.Get(ctx, id)
}

func (k Keeper) GetAllAgentsPaginated(ctx sdk.Context, pageReq *query.PageRequest) ([]types.Agent, *query.PageResponse, error) {
	return collcompat.CollectionPaginate(ctx, k.agents, pageReq,
		func(_ string, a types.Agent) (types.Agent, error) { return a, nil })
}

func (k Keeper) GetAllAgents(ctx sdk.Context) ([]types.Agent, error) {
	var agents []types.Agent
	err := k.agents.Walk(ctx, nil, func(_ string, a types.Agent) (bool, error) {
		agents = append(agents, a)
		return false, nil
	})
	return agents, err
}

func (k Keeper) SetActionLogEntry(ctx sdk.Context, e types.ActionLogEntry) error {
	return k.actionLog.Set(ctx, collections.Join(e.AgentId, e.Seq), e)
}

func (k Keeper) GetActionLogEntry(ctx sdk.Context, agentID string, seq uint64) (types.ActionLogEntry, error) {
	return k.actionLog.Get(ctx, collections.Join(agentID, seq))
}

func (k Keeper) GetAgentActionsPaginated(ctx sdk.Context, agentID string, pageReq *query.PageRequest) ([]types.ActionLogEntry, *query.PageResponse, error) {
	return collcompat.CollectionPaginate(ctx, k.actionLog, pageReq,
		func(_ collections.Pair[string, uint64], e types.ActionLogEntry) (types.ActionLogEntry, error) {
			return e, nil
		},
		collcompat.WithCollectionPaginationPairPrefix[string, uint64](agentID))
}

func (k Keeper) GetAllActionLog(ctx sdk.Context) ([]types.ActionLogEntry, error) {
	var entries []types.ActionLogEntry
	err := k.actionLog.Walk(ctx, nil, func(_ collections.Pair[string, uint64], e types.ActionLogEntry) (bool, error) {
		entries = append(entries, e)
		return false, nil
	})
	return entries, err
}
