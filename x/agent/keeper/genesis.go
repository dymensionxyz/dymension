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
		if err := k.SetAgent(ctx, a); err != nil {
			panic(err)
		}
	}
	for _, e := range g.ActionLog {
		if err := k.setActionLogEntry(ctx, e); err != nil {
			panic(err)
		}
	}
}

func ExportGenesis(ctx sdk.Context, k *Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	var agents []types.Agent
	if err := k.agents.Walk(ctx, nil, func(_ string, a types.Agent) (bool, error) {
		agents = append(agents, a)
		return false, nil
	}); err != nil {
		panic(err)
	}

	var actionLog []types.ActionLogEntry
	if err := k.actionLog.Walk(ctx, nil, func(_ collections.Pair[string, uint64], e types.ActionLogEntry) (bool, error) {
		actionLog = append(actionLog, e)
		return false, nil
	}); err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:    params,
		Agents:    agents,
		ActionLog: actionLog,
	}
}
