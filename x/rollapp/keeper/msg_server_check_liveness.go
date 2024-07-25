package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) LivenessSlash(goCtx context.Context, msg *types.MsgLivenessSlash) (*types.MsgLivenessSlashResponse, error) {
}
