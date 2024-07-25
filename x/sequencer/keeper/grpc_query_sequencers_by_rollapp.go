package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) SequencersByRollapp(c context.Context, req *types.QueryGetSequencersByRollappRequest) (*types.QueryGetSequencersByRollappResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := k.rollappKeeper.GetRollapp(ctx, req.RollappId); !ok {
		return nil, types.ErrUnknownRollappID
	}

	sequencers := k.GetSequencersByRollapp(ctx, req.RollappId)
	return &types.QueryGetSequencersByRollappResponse{
		Sequencers: sequencers,
	}, nil
}

func (k Keeper) SequencersByRollappByStatus(c context.Context, req *types.QueryGetSequencersByRollappByStatusRequest) (*types.QueryGetSequencersByRollappByStatusResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := k.rollappKeeper.GetRollapp(ctx, req.RollappId); !ok {
		return nil, types.ErrUnknownRollappID
	}

	sequencers := k.GetSequencersByRollappByStatus(
		ctx,
		req.RollappId,
		req.Status,
	)

	return &types.QueryGetSequencersByRollappByStatusResponse{
		Sequencers: sequencers,
	}, nil
}

// GetProposerByRollapp implements types.QueryServer.
func (k Keeper) GetProposerByRollapp(c context.Context, req *types.QueryGetProposerByRollappRequest) (*types.QueryGetProposerByRollappResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	seq, ok := k.GetProposer(ctx, req.RollappId)
	if !ok {
		return nil, types.ErrNoProposer
	}

	return &types.QueryGetProposerByRollappResponse{
		ProposerAddr: seq.SequencerAddress,
	}, nil
}

// GetNextProposerByRollapp implements types.QueryServer.
func (k Keeper) GetNextProposerByRollapp(c context.Context, req *types.QueryGetNextProposerByRollappRequest) (*types.QueryGetNextProposerByRollappResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	seq, ok := k.GetNextProposer(ctx, req.RollappId)
	if ok {
		return &types.QueryGetNextProposerByRollappResponse{
			NextProposerAddr:   seq.SequencerAddress,
			RotationInProgress: true,
		}, nil
	}

	// if rotation is not in progress, we return the expected next proposer in case for the next rotation
	expectedNext := k.ExpectedNextProposer(ctx, req.RollappId)
	return &types.QueryGetNextProposerByRollappResponse{
		NextProposerAddr:   expectedNext.SequencerAddress,
		RotationInProgress: false,
	}, nil
}
