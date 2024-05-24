package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// TriggerGenesisEvent noops
// Deprecated
func (k msgServer) TriggerGenesisEvent(context.Context, *types.MsgRollappGenesisEvent) (*types.MsgRollappGenesisEventResponse, error) {
	// TODO: return not implemented or deprecated error
	return &types.MsgRollappGenesisEventResponse{}, nil
}
