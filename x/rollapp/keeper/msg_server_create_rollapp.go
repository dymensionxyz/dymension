package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Already validated chain id in ValidateBasic, so we assume it's valid
	rollappId := types.MustNewChainID(msg.RollappId)

	if err := k.CheckIfRollappExists(ctx, rollappId); err != nil {
		return nil, err
	}

	k.SetRollapp(ctx, msg.GetRollapp())

	creator := sdk.MustAccAddressFromBech32(msg.Creator)

	if err := k.hooks.RollappCreated(ctx, msg.RollappId, msg.Alias, creator); err != nil {
		return nil, fmt.Errorf("rollapp created hook: %w", err)
	}

	if err := uevent.EmitTypedEvent(ctx, msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgCreateRollappResponse{}, nil
}
