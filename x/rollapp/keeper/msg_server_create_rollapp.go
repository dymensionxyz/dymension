package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("validate rollapp: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	rollappId, _ := types.NewChainID(msg.RollappId)
	if err := k.CheckIfRollappExists(ctx, rollappId); err != nil {
		return nil, err
	}

	k.SetRollapp(ctx, msg.GetRollapp())

	creator := sdk.MustAccAddressFromBech32(msg.Creator)

	if err := k.hooks.RollappCreated(ctx, msg.RollappId, msg.Alias, creator); err != nil {
		return nil, fmt.Errorf("rollapp created hook: %w", err)
	}

	if err := ctx.EventManager().EmitTypedEvent(msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgCreateRollappResponse{}, nil
}
