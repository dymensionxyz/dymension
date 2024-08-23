package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) UpdateRollappInformation(goCtx context.Context, msg *types.MsgUpdateRollappInformation) (*types.MsgUpdateRollappInformationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	updated, err := k.CheckAndUpdateRollappFields(ctx, msg)
	if err != nil {
		return nil, err
	}

	k.SetRollapp(ctx, updated)

	if err = uevent.EmitTypedEvent(ctx, msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateRollappInformationResponse{}, nil
}
