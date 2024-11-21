package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k msgServer) UpdateSequencerInformation(
	goCtx context.Context,
	msg *types.MsgUpdateSequencerInformation,
) (*types.MsgUpdateSequencerInformationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.RealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}
	defer func() {
		k.SetSequencer(ctx, seq)
	}()

	rollapp := k.rollappKeeper.MustGetRollapp(ctx, seq.RollappId)

	if err := msg.VMSpecificValidate(rollapp.VmType); err != nil {
		return nil, err
	}

	seq.Metadata = msg.Metadata

	if err := uevent.EmitTypedEvent(ctx, &seq); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateSequencerInformationResponse{}, nil
}

// UpdateOptInStatus : if false, then the sequencer will not be chosen as proposer or successor.
// If already chosen as proposer or successor, the change has no effect.
func (k msgServer) UpdateOptInStatus(goCtx context.Context,
	msg *types.MsgUpdateOptInStatus,
) (*types.MsgUpdateOptInStatus, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	seq, err := k.RealSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	if seq.NoticeStarted() {
		// prevent a sequencer who proposed in the past from becoming chosen as proposer again
		return nil, gerrc.ErrFailedPrecondition.Wrap(`tried to change opt in status after rotating: not allowed because
sequencers can only be proposer at most once`)
	}

	if err := seq.SetOptedIn(ctx, msg.OptedIn); err != nil {
		return nil, err
	}
	k.SetSequencer(ctx, seq)

	// maybe set as proposer if one is needed
	proposer := k.GetProposer(ctx, seq.RollappId)
	if proposer.Sentinel() {
		if err := k.RecoverFromSentinel(ctx, seq.RollappId); err != nil {
			return nil, err
		}
	}
	return &types.MsgUpdateOptInStatus{}, nil
}
