package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	db "github.com/tendermint/tm-db"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getStreamsFromIterator iterates over everything in a stream's iterator, until it reaches the end. Return all streams iterated over.
func (k Keeper) getStreamsFromIterator(ctx sdk.Context, iterator db.Iterator) []types.Stream {
	streams := []types.Stream{}
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		streamIDs := []uint64{}
		err := json.Unmarshal(iterator.Value(), &streamIDs)
		if err != nil {
			panic(err)
		}
		for _, streamID := range streamIDs {
			stream, err := k.GetStreamByID(ctx, streamID)
			if err != nil {
				panic(err)
			}
			streams = append(streams, *stream)
		}
	}
	return streams
}

// setStream set the stream inside store.
func (k Keeper) setStream(ctx sdk.Context, stream *types.Stream) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(stream)
	if err != nil {
		return err
	}
	store.Set(streamStoreKey(stream.Id), bz)
	return nil
}

// CreateStream creates a stream and sends coins to the stream.
func (k Keeper) CreateStream(ctx sdk.Context, coins sdk.Coins, distrTo sdk.AccAddress, startTime time.Time, epochIdentifier string, numEpochsPaidOver uint64) (uint64, error) {
	_, err := sdk.AccAddressFromBech32(distrTo.String())
	if err != nil {
		return 0, err
	}

	if !coins.IsAllPositive() {
		return 0, fmt.Errorf("all coins %s must be positive", coins)
	}

	moduleBalance := k.bk.GetAllBalances(ctx, authtypes.NewModuleAddress(types.ModuleName))
	alreadyAllocatedCoins := k.GetModuleToDistributeCoins(ctx)

	if !coins.IsAllLTE(moduleBalance.Sub(spendedCoins...)) {
		return 0, fmt.Errorf("insufficient module balance to distribute coins")
	}

	if (k.ek.GetEpochInfo(ctx, epochIdentifier) == epochstypes.EpochInfo{}) {
		return 0, fmt.Errorf("epoch identifier does not exist: %s", epochIdentifier)
	}

	if numEpochsPaidOver <= 0 {
		return 0, fmt.Errorf("numEpochsPaidOver must be greater than 0")
	}

	stream := types.Stream{
		Id:                   k.GetLastStreamID(ctx) + 1,
		DistributeTo:         distrTo.String(),
		Coins:                coins.Sort(),
		StartTime:            startTime,
		DistrEpochIdentifier: epochIdentifier,
		NumEpochsPaidOver:    numEpochsPaidOver,
	}

	err = k.setStream(ctx, &stream)
	if err != nil {
		return 0, err
	}
	k.SetLastStreamID(ctx, stream.Id)

	combinedKeys := combineKeys(types.KeyPrefixUpcomingStreams, getTimeKey(stream.StartTime))
	err = k.CreateStreamRefKeys(ctx, &stream, combinedKeys)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateStream,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(stream.Id)),
		),
	})

	return stream.Id, nil
}

// GetStreamByID returns stream from stream ID.
func (k Keeper) GetStreamByID(ctx sdk.Context, streamID uint64) (*types.Stream, error) {
	stream := types.Stream{}
	store := ctx.KVStore(k.storeKey)
	streamKey := streamStoreKey(streamID)
	if !store.Has(streamKey) {
		return nil, fmt.Errorf("stream with ID %d does not exist", streamID)
	}
	bz := store.Get(streamKey)
	if err := proto.Unmarshal(bz, &stream); err != nil {
		return nil, err
	}
	return &stream, nil
}

// GetStreamFromIDs returns multiple streams from a streamIDs array.
func (k Keeper) GetStreamFromIDs(ctx sdk.Context, streamIDs []uint64) ([]types.Stream, error) {
	streams := []types.Stream{}
	for _, streamID := range streamIDs {
		stream, err := k.GetStreamByID(ctx, streamID)
		if err != nil {
			return []types.Stream{}, err
		}
		streams = append(streams, *stream)
	}
	return streams, nil
}

// GetStreams returns upcoming, active, and finished streams.
func (k Keeper) GetStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.StreamsIterator(ctx))
}

// GetNotFinishedStreams returns both upcoming and active streams.
func (k Keeper) GetNotFinishedStreams(ctx sdk.Context) []types.Stream {
	return append(k.GetActiveStreams(ctx), k.GetUpcomingStreams(ctx)...)
}

// GetActiveStreams returns active streams.
func (k Keeper) GetActiveStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.ActiveStreamsIterator(ctx))
}

// GetUpcomingStreams returns upcoming streams.
func (k Keeper) GetUpcomingStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.UpcomingStreamsIterator(ctx))
}

// GetFinishedStreams returns finished streams.
func (k Keeper) GetFinishedStreams(ctx sdk.Context) []types.Stream {
	return k.getStreamsFromIterator(ctx, k.FinishedStreamsIterator(ctx))
}
