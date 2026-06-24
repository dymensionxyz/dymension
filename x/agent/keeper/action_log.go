package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k Keeper) setActionLogEntry(ctx sdk.Context, entry types.ActionLogEntry) error {
	return k.actionLog.Set(ctx, collections.Join(entry.AgentId, entry.Seq), entry)
}

// GetActionLogEntry returns the log entry for (agentID, seq) and whether it
// was found.
func (k Keeper) GetActionLogEntry(ctx sdk.Context, agentID string, seq uint64) (types.ActionLogEntry, bool) {
	entry, err := k.actionLog.Get(ctx, collections.Join(agentID, seq))
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ActionLogEntry{}, false
		}
		panic(err)
	}
	return entry, true
}
