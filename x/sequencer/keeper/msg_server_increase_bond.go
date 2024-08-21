package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// IncreaseBond implements types.MsgServer.
func (k msgServer) IncreaseBond(goCtx context.Context, msg *types.MsgIncreaseBond) (*types.MsgIncreaseBondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sequencer, err := k.bondUpdateAllowed(ctx, msg.GetCreator())
	if err != nil {
		return nil, err
	}

	// transfer the bond from the sequencer to the module account
	seqAcc := sdk.MustAccAddressFromBech32(msg.Creator)
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, seqAcc, types.ModuleName, sdk.NewCoins(msg.AddAmount))
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

	return &types.MsgIncreaseBondResponse{}, err
}

func (k msgServer) bondUpdateAllowed(ctx sdk.Context, senderAddress string) (types.Sequencer, error) {
	// check if the sequencer already exists
	sequencer, found := k.GetSequencer(ctx, senderAddress)
	if !found {
		return types.Sequencer{}, types.ErrUnknownSequencer
	}

	// check if the sequencer is bonded
	if !sequencer.IsBonded() {
		return types.Sequencer{}, types.ErrInvalidSequencerStatus
	}

	// check if sequencer is currently jailed
	if sequencer.Jailed {
		return types.Sequencer{}, types.ErrSequencerJailed
	}
	return sequencer, nil
}
