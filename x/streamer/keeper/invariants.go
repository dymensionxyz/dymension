package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "streams-count", StreamsCountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "last-stream-id", LastStreamIdInvariant(k))
	ir.RegisterRoute(types.ModuleName, "streamer-balance", StreamerBalanceInvariant(k))
	ir.RegisterRoute(types.ModuleName, "streams", StreamsInvariant(k))
}

// AllInvariants runs all invariants of the x/streamer module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := LastStreamIdInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = StreamsCountInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = StreamerBalanceInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = StreamsInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return "", false
	}
}

func StreamerBalanceInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		toDistCoins := k.GetModuleToDistributeCoins(ctx)
		balance := k.bk.GetAllBalances(ctx, k.ak.GetModuleAddress(types.ModuleName))

		insufficient := !toDistCoins.IsAllLTE(balance)
		if insufficient {
			msg += "streamer balance < toDistCoins"
			broken = true
		}

		return sdk.FormatInvariant(
			types.ModuleName, "streamer-balance",
			msg,
		), broken
	}
}

func StreamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		streams := k.GetNotFinishedStreams(ctx)
		for _, stream := range streams {
			if stream.FilledEpochs > stream.NumEpochsPaidOver {
				msg += fmt.Sprintf("filled epochs > num epochs paid over on stream %d", stream.Id)
				broken = true
			}

			overflow := !stream.DistributedCoins.IsAllLTE(stream.Coins)
			if overflow {
				msg += fmt.Sprintf("distributed coins > coins on stream %d", stream.Id)
				broken = true
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "streams",
			msg,
		), broken
	}
}

func LastStreamIdInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		streams := k.GetStreams(ctx)
		lastStreamId := k.GetLastStreamID(ctx)

		if len(streams) == 0 {
			if lastStreamId != 0 {
				msg += fmt.Sprintf("last stream id %d != 0\n", lastStreamId)
				broken = true
			}
		} else if streams[len(streams)-1].Id != lastStreamId {
			msg += fmt.Sprintf("last stream id %d != last stream id in store %d\n", streams[len(streams)-1].Id, lastStreamId)
			broken = true
		}

		return sdk.FormatInvariant(
			types.ModuleName, "last-stream-id",
			msg,
		), broken
	}
}

func StreamsCountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)

		streams := k.GetStreams(ctx)
		if len(streams) == 0 {
			return "no streams found", false
		}

		upcomingStreams := k.GetUpcomingStreams(ctx)
		activeStreams := k.GetActiveStreams(ctx)
		finishedStreams := k.GetFinishedStreams(ctx)

		broken = len(streams) != len(upcomingStreams)+len(activeStreams)+len(finishedStreams)

		return sdk.FormatInvariant(
			types.ModuleName, "streams-count",
			msg,
		), broken
	}
}
