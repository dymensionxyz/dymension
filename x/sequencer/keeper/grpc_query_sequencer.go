package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) Sequencers(c context.Context, req *types.QuerySequencersRequest) (*types.QuerySequencersResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	var seqs []types.Sequencer

	store := ctx.KVStore(k.storeKey)
	sequencerStore := prefix.NewStore(store, types.SequencersKey())

	pageRes, err := query.Paginate(sequencerStore, req.Pagination, func(key []byte, value []byte) error {
		var sequencer types.Sequencer
		if err := k.cdc.Unmarshal(value, &sequencer); err != nil {
			return err
		}
		seqs = append(seqs, sequencer)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QuerySequencersResponse{Sequencers: seqs, Pagination: pageRes}, nil
}

func (k Keeper) Sequencer(c context.Context, req *types.QueryGetSequencerRequest) (*types.QueryGetSequencerResponse, error) {
	if req == nil {
		return nil, gerrc.ErrInvalidArgument
	}
	ctx := sdk.UnwrapSDKContext(c)

	seq, err := k.RealSequencer(ctx, req.SequencerAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryGetSequencerResponse{Sequencer: seq}, nil
}
