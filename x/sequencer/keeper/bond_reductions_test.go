package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

const (
	seq1 = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	seq2 = "dym1d0wlmz987qlurs6e3kc6zd25z6wsdmnwx8tafy"
)

func TestGetMatureDecreasingBondIDs(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)

	t.Run("No mature bonds", func(t *testing.T) {
		ids := keeper.GetMatureDecreasingBondIDs(ctx, time.Now())
		require.Len(t, ids, 0)
	})

	t.Run("Mature bonds of multiple sequencers", func(t *testing.T) {
		bondReductionTime := time.Now()
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq1,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		// Not mature
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime.Add(time.Hour),
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})

		ids := keeper.GetMatureDecreasingBondIDs(ctx, bondReductionTime)
		require.Len(t, ids, 2)
	})
}

func TestGetBondReductionsBySequencer(t *testing.T) {
	keeper, ctx := keepertest.SequencerKeeper(t)

	t.Run("No bond reductions", func(t *testing.T) {
		ids := keeper.GetBondReductionsBySequencer(ctx, seq1)
		require.Len(t, ids, 0)
	})

	t.Run("Bond reductions of multiple sequencers", func(t *testing.T) {
		bondReductionTime := time.Now()
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq1,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime,
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})
		keeper.SetDecreasingBondQueue(ctx, types.BondReduction{
			SequencerAddress:   seq2,
			DecreaseBondTime:   bondReductionTime.Add(time.Hour),
			DecreaseBondAmount: sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		})

		ids := keeper.GetBondReductionsBySequencer(ctx, seq1)
		require.Len(t, ids, 1)

		ids = keeper.GetBondReductionsBySequencer(ctx, seq2)
		require.Len(t, ids, 2)
	})
}
