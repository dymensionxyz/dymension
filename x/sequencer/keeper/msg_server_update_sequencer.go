package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UpdateSequencerInformation defines a method for creating a new sequencer
func (k msgServer) UpdateSequencerInformation(goCtx context.Context, msg *types.MsgUpdateSequencerInformation) (*types.MsgUpdateSequencerInformationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidRequest, "validate basic: %v", err)
	}

	sequencer, found := k.GetSequencer(ctx, msg.Creator)
	if !found {
		return nil, types.ErrUnknownSequencer
	}

	if sequencer.Jailed {
		return nil, types.ErrSequencerJailed
	}

	rollapp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	if rollapp.Frozen {
		return nil, types.ErrRollappJailed
	}

	sequencer.Metadata = msg.Metadata

	k.SetSequencer(ctx, sequencer)

	if err := ctx.EventManager().EmitTypedEvent(&sequencer); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateSequencerInformationResponse{}, nil
}
