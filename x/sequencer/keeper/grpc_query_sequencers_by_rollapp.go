package keeper

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) SequencersByRollapp(c context.Context, req *types.QueryGetSequencersByRollappRequest) (*types.QueryGetSequencersByRollappResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	if _, ok := k.rollappKeeper.GetRollapp(ctx, req.RollappId); !ok {
		return nil, errors.Join(gerrc.ErrNotFound, types.ErrUnknownRollappID)
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
		return nil, errors.Join(gerrc.ErrNotFound, types.ErrUnknownRollappID)
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
		ProposerAddr: seq.Address,
	}, nil
}

// GetNextProposerByRollapp implements types.QueryServer.
func (k Keeper) GetNextProposerByRollapp(c context.Context, req *types.QueryGetNextProposerByRollappRequest) (*types.QueryGetNextProposerByRollappResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	// if rotation is not in progress, we return the expected next proposer in case for the next rotation
	expectedNext := k.ExpectedNextProposer(ctx, req.RollappId)
	return &types.QueryGetNextProposerByRollappResponse{
		NextProposerAddr:   expectedNext.Address,
		RotationInProgress: k.IsRotating(ctx, req.RollappId),
	}, nil
}

func (k Keeper) Proposers(c context.Context, req *types.QueryProposersRequest) (*types.QueryProposersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var proposers []types.Sequencer
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	sequencerStore := prefix.NewStore(store, types.ProposerByRollappKey(""))

	pageRes, err := query.Paginate(sequencerStore, req.Pagination, func(key []byte, value []byte) error {
		proposer, ok := k.GetSequencer(ctx, string(value))
		if ok {
			proposers = append(proposers, proposer)
		}
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryProposersResponse{Proposers: proposers, Pagination: pageRes}, nil
}
