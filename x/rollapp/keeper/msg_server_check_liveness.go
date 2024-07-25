package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) LivenessSlash(goCtx context.Context, msg *types.MsgLivenessSlash) (*types.MsgLivenessSlashResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var rewardAddr sdk.AccAddress

	res, err := k.SlashLiveness(ctx, msg.GetRollappId(), rewardAddr)
	if err != nil {
		return nil, errorsmod.Wrap(err, "slash liveness")
	}
	// TODO: event, maybe better in keeper?
	return &types.MsgLivenessSlashResponse{
		Slashed: !res.slashed.Empty(), // TODO: maybe slashed is a bad name, maybe success is better
	}, nil
}
