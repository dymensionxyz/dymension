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
	for _, e := range g.Escrows {
		if err := k.escrows.Set(ctx, e.AgentId, e); err != nil {
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

	if err := k.escrows.Walk(ctx, nil, func(_ string, e types.AgentEscrow) (stop bool, err error) {
		g.Escrows = append(g.Escrows, e)
		return false, nil
	}); err != nil {
		panic(err)
	}

	return &g
}
