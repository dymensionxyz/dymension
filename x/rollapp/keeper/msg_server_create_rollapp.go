package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Already validated chain id in ValidateBasic, so we assume it's valid
	rollappId := types.MustNewChainID(msg.RollappId)

	// when creating a new Rollapp, the chainID revision number should always be 1
	// As we manage rollapp revision in the RollappKeeper, we don't allow increasing chainID revision number.
	// (IBC has it's own concept of revision which we assume is always 1)
	if rollappId.GetRevisionNumber() != 1 {
		return nil, errorsmod.Wrapf(types.ErrInvalidRollappID, "revision number should be 1, got: %d", rollappId.GetRevisionNumber())
	}

	if err := k.CheckIfRollappExists(ctx, rollappId); err != nil {
		return nil, err
	}
	if err := k.validMinBond(ctx, msg.MinSequencerBond); err != nil {
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
