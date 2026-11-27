package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k Keeper) SetAgent(ctx sdk.Context, agent types.Agent) error {
	return k.agents.Set(ctx, agent.Id, agent)
}

// GetAgent returns the agent and whether it was found.
func (k Keeper) GetAgent(ctx sdk.Context, id string) (types.Agent, bool) {
	agent, err := k.agents.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Agent{}, false
		}
		panic(err)
	}
	return agent, true
}

// GetAgentOwner returns the owner of a registered agent and whether it exists.
func (k Keeper) GetAgentOwner(ctx sdk.Context, agentID string) (string, bool) {
	agent, found := k.GetAgent(ctx, agentID)
	if !found {
		return "", false
	}
	return agent.Owner, true
}

// RebaseLastActionHeights re-expresses every agent's last action height
// relative to height for a zero-height export. The restarted chain's block 1
// follows height, so each agent keeps exactly the heartbeat age it had: a
// stale agent stays stale and a fresh one expires on its original schedule.
func (k Keeper) RebaseLastActionHeights(ctx sdk.Context, height int64) error {
	return k.updateAllAgents(ctx, func(a *types.Agent) { a.LastActionHeight -= height })
}

// updateAllAgents collects all agents before writing so the store is never
// written while an iterator is open.
func (k Keeper) updateAllAgents(ctx sdk.Context, update func(*types.Agent)) error {
	iter, err := k.agents.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	agents, err := iter.Values()
	if err != nil {
		return err
	}
	for _, a := range agents {
		update(&a)
		if err := k.SetAgent(ctx, a); err != nil {
			return err
		}
	}
	return nil
}
