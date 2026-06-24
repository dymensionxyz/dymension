package keeper

import (
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
		if err := k.SetActionLogEntry(ctx, e); err != nil {
			panic(err)
		}
	}
}

func ExportGenesis(ctx sdk.Context, k *Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	agents, err := k.GetAllAgents(ctx)
	if err != nil {
		panic(err)
	}
	actionLog, err := k.GetAllActionLog(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		Params:    params,
		Agents:    agents,
		ActionLog: actionLog,
	}
}
