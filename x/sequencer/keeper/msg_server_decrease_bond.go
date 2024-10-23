package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"
)

// DecreaseBond implements types.MsgServer.
func (k msgServer) DecreaseBond(goCtx context.Context, msg *types.MsgDecreaseBond) (*types.MsgDecreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, found := k.GetSequencer(ctx, msg.GetCreator())
	if !found {
		return nil, types.ErrSequencerNotFound
	}
	return &types.MsgDecreaseBondResponse{}, k.tryUnbond(ctx, seq, uptr.To(msg.GetDecreaseAmount()))
}
