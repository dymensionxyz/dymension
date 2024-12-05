package streamer

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

func NewStreamerProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CreateStreamProposal:
			return HandleCreateStreamProposal(ctx, k, c)
		case *types.TerminateStreamProposal:
			return HandleTerminateStreamProposal(ctx, k, c)
		case *types.ReplaceStreamDistributionProposal:
			return HandleReplaceStreamDistributionProposal(ctx, k, c)
		case *types.UpdateStreamDistributionProposal:
			return HandleUpdateStreamDistributionProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(types.ErrUnknownRequest, "unrecognized streamer proposal content type: %T", c)
		}
	}
}

// HandleCreateStreamProposal is a handler for executing a passed community spend proposal
func HandleCreateStreamProposal(ctx sdk.Context, k keeper.Keeper, p *types.CreateStreamProposal) (err error) {
	defer func() {
		if err != nil {
			k.Logger(ctx).Error("Create stream proposal.", "error", err)
		}
	}()
	_, err = k.CreateStream(ctx, p.Coins, p.DistributeToRecords, p.StartTime, p.DistrEpochIdentifier, p.NumEpochsPaidOver, p.Sponsored)
	return
}

// HandleTerminateStreamProposal is a handler for executing a passed community spend proposal
func HandleTerminateStreamProposal(ctx sdk.Context, k keeper.Keeper, p *types.TerminateStreamProposal) (err error) {
	defer func() {
		if err != nil {
			k.Logger(ctx).Error("Terminate stream proposal.", "error", err)
		}
	}()
	err = k.TerminateStream(ctx, p.StreamId)
	return
}

// HandleReplaceStreamDistributionProposal is a handler for executing a passed community spend proposal
func HandleReplaceStreamDistributionProposal(ctx sdk.Context, k keeper.Keeper, p *types.ReplaceStreamDistributionProposal) (err error) {
	defer func() {
		if err != nil {
			k.Logger(ctx).Error("Replace stream proposal.", "error", err)
		}
	}()
	err = k.ReplaceDistrRecords(ctx, p.StreamId, p.Records)
	return
}

// HandleUpdateStreamDistributionProposal is a handler for executing a passed community spend proposal
func HandleUpdateStreamDistributionProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateStreamDistributionProposal) (err error) {
	defer func() {
		if err != nil {
			k.Logger(ctx).Error("Update stream proposal.", "error", err)
		}
	}()
	err = k.UpdateDistrRecords(ctx, p.StreamId, p.Records)
	return
}
