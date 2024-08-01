package dymns_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/dymns"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestExportThenInitGenesis(t *testing.T) {
	now := time.Now().UTC()

	oldKeeper, _, _, oldCtx := testkeeper.DymNSKeeper(t)
	oldCtx = oldCtx.WithBlockTime(now)

	// Setup genesis state
	owner1 := "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	owner2 := "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d"
	anotherAccount := "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96"

	bidder1 := "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	bidder2 := "dym1wl60kvsq5c4wa600h7rnez8dguk5lpnqp4u0y2"

	buyer1 := "dym1nxswr2xhky3k0rt65paatpzjw8mg5d5rmylu3z"
	buyer2 := "dym16vz9q7m9cxfjgf3v4tm4aqf50vde84hr39kqgd"
	buyer3 := "dym1s62euc7nqg029m9v2rl77hf66u69pkuv2sg3uv"

	dymName1 := dymnstypes.DymName{
		Name:       "my-name",
		Owner:      owner1,
		Controller: owner1,
		ExpireAt:   now.Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{
			{
				Type:  dymnstypes.DymNameConfigType_NAME,
				Path:  "pseudo",
				Value: anotherAccount,
			},
		},
	}
	require.NoError(t, oldKeeper.SetDymName(oldCtx, dymName1))

	dymName2 := dymnstypes.DymName{
		Name:       "light",
		Owner:      owner2,
		Controller: owner2,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	require.NoError(t, oldKeeper.SetDymName(oldCtx, dymName2))

	dymName3Expired := dymnstypes.DymName{
		Name:       "expired",
		Owner:      owner2,
		Controller: owner2,
		ExpireAt:   now.Add(-time.Hour).Unix(),
	}
	require.NoError(t, oldKeeper.SetDymName(oldCtx, dymName3Expired))

	so1 := dymnstypes.SellOrder{
		Name:      dymName1.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
		HighestBid: &dymnstypes.SellOrderBid{
			Bidder: bidder1,
			Price:  dymnsutils.TestCoin(200),
		},
	}
	require.NoError(t, oldKeeper.SetSellOrder(oldCtx, so1))

	so2 := dymnstypes.SellOrder{
		Name:      dymName2.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(900),
		HighestBid: &dymnstypes.SellOrderBid{
			Bidder: bidder2,
			Price:  dymnsutils.TestCoin(800),
		},
	}
	require.NoError(t, oldKeeper.SetSellOrder(oldCtx, so2))

	so3 := dymnstypes.SellOrder{
		Name:      dymName3Expired.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(200),
	}
	require.NoError(t, oldKeeper.SetSellOrder(oldCtx, so3))

	offer1 := dymnstypes.OfferToBuy{
		Id:         "1",
		Name:       dymName1.Name,
		Buyer:      buyer1,
		OfferPrice: dymnsutils.TestCoin(100),
	}
	require.NoError(t, oldKeeper.SetOfferToBuy(oldCtx, offer1))

	offer2 := dymnstypes.OfferToBuy{
		Id:         "2",
		Name:       dymName2.Name,
		Buyer:      buyer2,
		OfferPrice: dymnsutils.TestCoin(200),
	}
	require.NoError(t, oldKeeper.SetOfferToBuy(oldCtx, offer2))

	offer3OfExpired := dymnstypes.OfferToBuy{
		Id:         "3",
		Name:       dymName3Expired.Name,
		Buyer:      buyer3,
		OfferPrice: dymnsutils.TestCoin(300),
	}
	require.NoError(t, oldKeeper.SetOfferToBuy(oldCtx, offer3OfExpired))

	// Export genesis state
	genState := dymns.ExportGenesis(oldCtx, oldKeeper)

	t.Run("params should be exported correctly", func(t *testing.T) {
		require.Equal(t, oldKeeper.GetParams(oldCtx), genState.Params)
	})

	t.Run("dym-names should be exported correctly", func(t *testing.T) {
		require.Len(t, genState.DymNames, 2)
		require.Contains(t, genState.DymNames, dymName1)
		require.Contains(t, genState.DymNames, dymName2)
		// Expired dym-name should not be exported
	})

	t.Run("sell orders's non-refunded bids should be exported correctly", func(t *testing.T) {
		require.Len(t, genState.SellOrderBids, 2)
		require.Contains(t, genState.SellOrderBids, *so1.HighestBid)
		require.Contains(t, genState.SellOrderBids, *so2.HighestBid)
		// Expired sell order should not be exported
	})

	t.Run("offers-to-buy should be exported correctly", func(t *testing.T) {
		require.Len(t, genState.OffersToBuy, 3)
		require.Contains(t, genState.OffersToBuy, offer1)
		require.Contains(t, genState.OffersToBuy, offer2)
		require.Contains(
			t, genState.OffersToBuy, offer3OfExpired,
			"offer should be exported even if the dym-name is expired",
		)
	})

	// Init genesis state

	genState.Params.Misc.BeginEpochHookIdentifier = "week" // Change the epoch identifier to test if it is imported correctly
	genState.Params.Misc.SellOrderDuration = 9999 * time.Hour
	genState.Params.PreservedRegistration.ExpirationEpoch = 8888

	newDymNsKeeper, newBankKeeper, _, newCtx := testkeeper.DymNSKeeper(t)
	newCtx = newCtx.WithBlockTime(now)

	dymns.InitGenesis(newCtx, newDymNsKeeper, *genState)

	t.Run("params should be imported correctly", func(t *testing.T) {
		importedParams := newDymNsKeeper.GetParams(newCtx)
		require.Equal(t, genState.Params, importedParams)
		require.Equal(t, "week", importedParams.Misc.BeginEpochHookIdentifier)
		require.Equal(t, 9999*time.Hour, importedParams.Misc.SellOrderDuration)
		require.Equal(t, int64(8888), importedParams.PreservedRegistration.ExpirationEpoch)
	})

	t.Run("Dym-Names should be imported correctly", func(t *testing.T) {
		require.Len(t,
			newDymNsKeeper.GetAllNonExpiredDymNames(newCtx),
			2,
			"expired dym-name should not be imported",
		)

		require.Equal(t, &dymName1, newDymNsKeeper.GetDymName(newCtx, dymName1.Name))
		require.Equal(t, &dymName2, newDymNsKeeper.GetDymName(newCtx, dymName2.Name))
		require.Nil(t,
			newDymNsKeeper.GetDymName(newCtx, dymName3Expired.Name),
			"expired dym-name should not be imported",
		)
	})

	t.Run("reverse lookup should be created correctly", func(t *testing.T) {
		owned, err := newDymNsKeeper.GetDymNamesOwnedBy(newCtx, owner1)
		require.NoError(t, err)
		require.Len(t, owned, 1)

		owned, err = newDymNsKeeper.GetDymNamesOwnedBy(newCtx, owner2)
		require.NoError(t, err)
		require.Len(t, owned, 1)

		names, err := newDymNsKeeper.GetDymNamesContainsConfiguredAddress(newCtx, owner1)
		require.NoError(t, err)
		require.Len(t, names, 1)

		names, err = newDymNsKeeper.GetDymNamesContainsConfiguredAddress(newCtx, owner2)
		require.NoError(t, err)
		require.Len(t, names, 1)

		names, err = newDymNsKeeper.GetDymNamesContainsConfiguredAddress(newCtx, anotherAccount)
		require.NoError(t, err)
		require.Len(t, names, 1)

		names, err = newDymNsKeeper.GetDymNamesContainsHexAddress(newCtx, sdk.MustAccAddressFromBech32(owner1).Bytes())
		require.NoError(t, err)
		require.Len(t, names, 1)

		names, err = newDymNsKeeper.GetDymNamesContainsHexAddress(newCtx, sdk.MustAccAddressFromBech32(owner2).Bytes())
		require.NoError(t, err)
		require.Len(t, names, 1)

		names, err = newDymNsKeeper.GetDymNamesContainsHexAddress(newCtx, sdk.MustAccAddressFromBech32(anotherAccount).Bytes())
		require.NoError(t, err)
		require.Empty(t, names, 0)
	})

	t.Run("sell orders's non-refunded bids should be refunded correctly", func(t *testing.T) {
		require.Equal(t,
			dymnsutils.TestCoin(200),
			newBankKeeper.GetBalance(newCtx, sdk.MustAccAddressFromBech32(bidder1), params.BaseDenom),
		)
		require.Equal(t,
			dymnsutils.TestCoin(800),
			newBankKeeper.GetBalance(newCtx, sdk.MustAccAddressFromBech32(bidder2), params.BaseDenom),
		)
	})

	t.Run("non-refunded offers should be refunded correctly", func(t *testing.T) {
		require.Equal(t,
			dymnsutils.TestCoin(100),
			newBankKeeper.GetBalance(newCtx, sdk.MustAccAddressFromBech32(buyer1), params.BaseDenom),
		)
		require.Equal(t,
			dymnsutils.TestCoin(200),
			newBankKeeper.GetBalance(newCtx, sdk.MustAccAddressFromBech32(buyer2), params.BaseDenom),
		)
		require.Equal(t,
			dymnsutils.TestCoin(300),
			newBankKeeper.GetBalance(newCtx, sdk.MustAccAddressFromBech32(buyer3), params.BaseDenom),
		)
	})

	// Init genesis state but with invalid input
	newDymNsKeeper, newBankKeeper, _, newCtx = testkeeper.DymNSKeeper(t)

	t.Run("invalid params", func(t *testing.T) {
		require.Panics(t, func() {
			dymns.InitGenesis(newCtx, newDymNsKeeper, dymnstypes.GenesisState{
				Params: dymnstypes.Params{
					Price: dymnstypes.PriceParams{}, // empty
					Misc:  dymnstypes.MiscParams{},  // empty
				},
			})
		})
	})

	t.Run("invalid dym-name", func(t *testing.T) {
		require.Panics(t, func() {
			dymns.InitGenesis(newCtx, newDymNsKeeper, dymnstypes.GenesisState{
				Params: dymnstypes.DefaultParams(),
				DymNames: []dymnstypes.DymName{
					{}, // empty content
				},
			})
		})
	})

	t.Run("invalid highest bid", func(t *testing.T) {
		require.Panics(t, func() {
			dymns.InitGenesis(newCtx, newDymNsKeeper, dymnstypes.GenesisState{
				Params: dymnstypes.DefaultParams(),
				SellOrderBids: []dymnstypes.SellOrderBid{
					{}, // empty content
				},
			})
		})
	})

	t.Run("invalid offer", func(t *testing.T) {
		require.Panics(t, func() {
			dymns.InitGenesis(newCtx, newDymNsKeeper, dymnstypes.GenesisState{
				Params: dymnstypes.DefaultParams(),
				OffersToBuy: []dymnstypes.OfferToBuy{
					{}, // empty content
				},
			})
		})
	})
}
