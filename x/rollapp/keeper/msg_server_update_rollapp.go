package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) UpdateRollappInformation(goCtx context.Context, msg *types.MsgUpdateRollappInformation) (*types.MsgUpdateRollappInformationResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.UpdateRollapp(ctx, msg); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvent(msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateRollappInformationResponse{}, nil
}
