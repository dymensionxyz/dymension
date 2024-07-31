package keeper_test

import (
	"testing"
	"time"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func TestGetSetParams(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	params := dymnstypes.DefaultParams()

	err := dk.SetParams(ctx, params)
	require.NoError(t, err)

	require.Equal(t, params, dk.GetParams(ctx))

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Misc.BeginEpochHookIdentifier = ""
		require.Error(t, dk.SetParams(ctx, params))
	})

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Price.PriceDenom = ""
		require.Error(t, dk.SetParams(ctx, params))
	})

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: "@",
				Aliases: nil,
			},
		}
		require.Error(t, dk.SetParams(ctx, params))
	})

	t.Run("can not set invalid params", func(t *testing.T) {
		params := dymnstypes.DefaultParams()
		params.Misc.GracePeriodDuration = -999 * time.Hour
		require.Error(t, dk.SetParams(ctx, params))
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_CheckChainIsCoinType60ByChainId(t *testing.T) {
	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

	const chainIdInjective = "injective-1"

	params := dk.GetParams(ctx)

	t.Run("roll-app is coin-type 60", func(t *testing.T) {
		rollApp1 := rollapptypes.Rollapp{
			RollappId: "ra_1-1",
			Creator:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		}
		rk.SetRollapp(ctx, rollApp1)

		require.True(t, dk.CheckChainIsCoinType60ByChainId(ctx, rollApp1.RollappId))
	})

	t.Run("chain-id in params is coin-type 60", func(t *testing.T) {
		params.Chains.CoinType60ChainIds = []string{chainIdInjective}
		err := dk.SetParams(ctx, params)
		require.NoError(t, err)

		require.True(t, dk.CheckChainIsCoinType60ByChainId(ctx, chainIdInjective))
	})

	t.Run("otherwise not coin-type 60", func(t *testing.T) {
		require.False(t, dk.CheckChainIsCoinType60ByChainId(ctx, "cosmoshub-4"))
	})

	t.Run("chain-id not in params is not coin-type 60 regardless actual", func(t *testing.T) {
		params.Chains.CoinType60ChainIds = nil
		err := dk.SetParams(ctx, params)
		require.NoError(t, err)

		require.False(t, dk.CheckChainIsCoinType60ByChainId(ctx, chainIdInjective))
	})
}
