package streamer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/dymensionxyz/dymension/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/x/streamer/types"
)

func NewStreamerProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CreateStreamProposal:
			return HandleCreateStreamProposal(ctx, k, c)
		case *types.StopStreamProposal:
			return HandleStopStreamProposal(ctx, k, c)
		case *types.ReplaceStreamDistributionProposal:
			return HandleReplaceStreamDistributionProposal(ctx, k, c)
		case *types.UpdateStreamDistributionProposal:
			return HandleUpdateStreamDistributionProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized streamer proposal content type: %T", c)
		}
	}
}

// HandleCreateStreamProposal is a handler for executing a passed community spend proposal
func HandleCreateStreamProposal(ctx sdk.Context, k keeper.Keeper, p *types.CreateStreamProposal) error {
	distrInfo, err := k.NewDistrInfo(ctx, p.DistributeToRecords)
	if err != nil {
		return err
	}

	_, err = k.CreateStream(ctx, p.Coins, distrInfo, p.StartTime, p.DistrEpochIdentifier, p.NumEpochsPaidOver)
	if err != nil {
		return err
	}
	return nil
}

// HandleStopStreamProposal is a handler for executing a passed community spend proposal
func HandleStopStreamProposal(ctx sdk.Context, k keeper.Keeper, p *types.StopStreamProposal) error {
	stream, err := k.GetStreamByID(ctx, p.StreamId)
	if err != nil {
		return err
	}

	if stream.IsFinishedStream(ctx.BlockTime()) {
		return sdkerrors.Wrapf(types.ErrInvalidStreamStatus, "stream %d is already finished", p.StreamId)
	}

	return k.StopStream(ctx, p.StreamId)
}

// HandleReplaceStreamDistributionProposal is a handler for executing a passed community spend proposal
func HandleReplaceStreamDistributionProposal(ctx sdk.Context, k keeper.Keeper, p *types.ReplaceStreamDistributionProposal) error {
	stream, err := k.GetStreamByID(ctx, p.StreamId)
	if err != nil {
		return err
	}

	if stream.IsFinishedStream(ctx.BlockTime()) {
		return sdkerrors.Wrapf(types.ErrInvalidStreamStatus, "stream %d is already finished", p.StreamId)
	}

	return k.ReplaceDistrRecords(ctx, p.StreamId, p.Records)
}

// HandleUpdateStreamDistributionProposal is a handler for executing a passed community spend proposal
func HandleUpdateStreamDistributionProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateStreamDistributionProposal) error {
	stream, err := k.GetStreamByID(ctx, p.StreamId)
	if err != nil {
		return err
	}

	if stream.IsFinishedStream(ctx.BlockTime()) {
		return sdkerrors.Wrapf(types.ErrInvalidStreamStatus, "stream %d is already finished", p.StreamId)
	}

	return k.UpdateDistrRecords(ctx, p.StreamId, p.Records)
}
