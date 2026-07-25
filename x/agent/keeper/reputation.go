package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k Keeper) SetFeedback(ctx sdk.Context, fb types.Feedback) error {
	return k.feedback.Set(ctx, collections.Join(fb.AgentId, fb.Client), fb)
}

// GetFeedback returns the feedback for (agentID, client) and whether it was
// found.
func (k Keeper) GetFeedback(ctx sdk.Context, agentID, client string) (types.Feedback, bool) {
	fb, err := k.feedback.Get(ctx, collections.Join(agentID, client))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Feedback{}, false
		}
		panic(err)
	}
	return fb, true
}

func (k Keeper) SetReputation(ctx sdk.Context, rep types.Reputation) error {
	return k.reputation.Set(ctx, rep.AgentId, rep)
}

// GetReputation returns the reputation aggregate for agentID and whether it
// was found.
func (k Keeper) GetReputation(ctx sdk.Context, agentID string) (types.Reputation, bool) {
	rep, err := k.reputation.Get(ctx, agentID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Reputation{}, false
		}
		panic(err)
	}
	return rep, true
}
