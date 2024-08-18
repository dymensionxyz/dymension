package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// UpdateSequencerInformation defines a method for updating a sequencer
func (k msgServer) UpdateSequencerInformation(goCtx context.Context, msg *types.MsgUpdateSequencerInformation) (*types.MsgUpdateSequencerInformationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sequencer, found := k.GetSequencer(ctx, msg.Creator)
	if !found {
		return nil, types.ErrUnknownSequencer
	}

	if sequencer.Jailed {
		return nil, types.ErrSequencerJailed
	}

	rollapp, found := k.rollappKeeper.GetRollapp(ctx, sequencer.RollappId)
	if !found {
		return nil, types.ErrUnknownRollappID
	}

	if rollapp.Frozen {
		return nil, types.ErrRollappJailed
	}

	if err := msg.VMSpecificValidate(rollapp.VmType); err != nil {
		return nil, err
	}

	sequencer.Metadata = msg.Metadata

	k.SetSequencer(ctx, sequencer)

	if err := ctx.EventManager().EmitTypedEvent(&sequencer); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateSequencerInformationResponse{}, nil
}
