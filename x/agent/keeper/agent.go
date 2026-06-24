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
