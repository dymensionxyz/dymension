package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func InitGenesis(ctx sdk.Context, k *Keeper, g types.GenesisState) {
	if err := k.SetParams(ctx, g.Params); err != nil {
		panic(err)
	}
	for _, a := range g.Agents {
		if err := k.agents.Set(ctx, a.Id, a); err != nil {
			panic(err)
		}
	}
	for _, e := range g.ActionLog {
		if err := k.actionLog.Set(ctx, collections.Join(e.AgentId, e.Seq), e); err != nil {
			panic(err)
		}
	}
	// Reputation aggregates are rebuilt from the feedback set rather than
	// imported, so they cannot drift from the records.
	for _, f := range g.Feedbacks {
		if err := k.feedback.Set(ctx, collections.Join(f.AgentId, f.Client), f); err != nil {
			panic(err)
		}
		rep, found := k.GetReputation(ctx, f.AgentId)
		if !found {
			rep = types.Reputation{AgentId: f.AgentId}
		}
		rep.Count++
		rep.ScoreSum += uint64(f.Score)
		if err := k.reputation.Set(ctx, f.AgentId, rep); err != nil {
			panic(err)
		}
	}
}

func ExportGenesis(ctx sdk.Context, k *Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	g := types.GenesisState{
		Params: params,
	}

	if err := k.agents.Walk(ctx, nil, func(_ string, a types.Agent) (stop bool, err error) {
		g.Agents = append(g.Agents, a)
		return false, nil
	}); err != nil {
		panic(err)
	}

	if err := k.actionLog.Walk(ctx, nil, func(_ collections.Pair[string, uint64], e types.ActionLogEntry) (stop bool, err error) {
		g.ActionLog = append(g.ActionLog, e)
		return false, nil
	}); err != nil {
		panic(err)
	}

	// key order == (agent_id, client) order, so the export is deterministic
	if err := k.feedback.Walk(ctx, nil, func(_ collections.Pair[string, string], f types.Feedback) (stop bool, err error) {
		g.Feedbacks = append(g.Feedbacks, f)
		return false, nil
	}); err != nil {
		panic(err)
	}

	return &g
}
