package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) UpdateRollappInformation(goCtx context.Context, msg *types.MsgUpdateRollappInformation) (*types.MsgUpdateRollappInformationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate update: %w", err)
	}

	updated, err := k.CanUpdateRollapp(ctx, msg)
	if err != nil {
		return nil, err
	}

	k.SetRollapp(ctx, updated)

	if err = ctx.EventManager().EmitTypedEvent(msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateRollappInformationResponse{}, nil
}
