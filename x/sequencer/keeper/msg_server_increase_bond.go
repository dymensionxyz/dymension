package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// IncreaseBond implements types.MsgServer.
func (k msgServer) IncreaseBond(goCtx context.Context, msg *types.MsgIncreaseBond) (*types.MsgIncreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sequencer, found := k.GetSequencer(ctx, msg.GetCreator())
	if !found {
		return nil, types.ErrUnknownSequencer
	}
	if !sequencer.IsBonded() {
		return nil, types.ErrInvalidSequencerStatus
	}

	// transfer the bond from the sequencer to the module account
	seqAcc := sdk.MustAccAddressFromBech32(msg.Creator)
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.AddAmount))
	if err != nil {
		return nil, err
	}

	// update the sequencers bond amount in state
	sequencer.Tokens = sequencer.Tokens.Add(msg.AddAmount)
	k.UpdateSequencer(ctx, &sequencer, sequencer.Status)

	// emit a typed event which includes the added amount and the active bond amount
	err = uevent.EmitTypedEvent(ctx,
		&types.EventIncreasedBond{
			Sequencer:   msg.Creator,
			Bond:        sequencer.Tokens,
			AddedAmount: msg.AddAmount,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgIncreaseBondResponse{}, nil
}
