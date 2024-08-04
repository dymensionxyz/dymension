package keeper_test

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/stretchr/testify/require"
)

func TestKeeper_IncreaseBuyOfferCountAndGet(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Zero(t, dk.GetCountBuyOffer(ctx))

	count := dk.IncreaseBuyOfferCountAndGet(ctx)
	require.Equal(t, uint64(1), count)
	require.Equal(t, uint64(1), dk.GetCountBuyOffer(ctx))

	count = dk.IncreaseBuyOfferCountAndGet(ctx)
	require.Equal(t, uint64(2), count)
	require.Equal(t, uint64(2), dk.GetCountBuyOffer(ctx))

	count = dk.IncreaseBuyOfferCountAndGet(ctx)
	require.Equal(t, uint64(3), count)
	require.Equal(t, uint64(3), dk.GetCountBuyOffer(ctx))

	dk.SetCountBuyOffer(ctx, math.MaxUint64-1)

	count = dk.IncreaseBuyOfferCountAndGet(ctx)
	require.Equal(t, uint64(math.MaxUint64), count)
	require.Equal(t, uint64(math.MaxUint64), dk.GetCountBuyOffer(ctx))

	require.Panics(t, func() {
		dk.IncreaseBuyOfferCountAndGet(ctx)
	}, "expect panic on overflow when increasing count of all-time buy offer records greater than uint64")
}

func TestKeeper_GetSetInsertNewBuyOffer(t *testing.T) {
	// TODO DymNS: add test for Sell/Buy Alias

	buyerA := testAddr(1).bech32()

	t.Run("get non-exists offer should returns nil", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer := dk.GetBuyOffer(ctx, "non-exists")
		require.Nil(t, offer)

		offer = dk.GetBuyOffer(ctx, "183418")
		require.Nil(t, offer)
	})

	t.Run("should returns error when set empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		err := dk.SetBuyOffer(ctx, dymnstypes.BuyOffer{
			Id:         "",
			Name:       "a",
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		})
		require.Error(t, err)
	})

	t.Run("can set and can get", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer := dymnstypes.BuyOffer{
			Id:         "1",
			Name:       "a",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetBuyOffer(ctx, offer)
		require.NoError(t, err)

		offerGot := dk.GetBuyOffer(ctx, "1")
		require.NotNil(t, offerGot)

		require.Equal(t, offer, *offerGot)
	})

	t.Run("should panic when insert non-empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.Panics(t, func() {
			_, _ = dk.InsertNewBuyOffer(ctx, dymnstypes.BuyOffer{
				Id:         "1",
				Name:       "a",
				Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			})
		})
	})

	t.Run("can insert", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1 := dymnstypes.BuyOffer{
			Id:         "",
			Name:       "a",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer, err := dk.InsertNewBuyOffer(ctx, offer1)
		require.NoError(t, err)
		require.Equal(t, "1", offer.Id)

		offerGot := dk.GetBuyOffer(ctx, "1")
		require.NotNil(t, offerGot)

		offer1.Id = "1"
		require.Equal(t, offer1, *offerGot)
	})

	// TODO DymNS: add case for same id buy different type, should be allowed and not overwritten

	t.Run("can not insert duplicated", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetCountBuyOffer(ctx, 1)
		nextId := "2"

		existing := dymnstypes.BuyOffer{
			Id:         nextId,
			Name:       "a",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetBuyOffer(ctx, existing)
		require.NoError(t, err)

		offer := dymnstypes.BuyOffer{
			Id:         "",
			Name:       "a",
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		_, err = dk.InsertNewBuyOffer(ctx, offer)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Buy-Order-ID already exists")
	})

	t.Run("should automatically fill ID when insert", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1 := dymnstypes.BuyOffer{
			Id:         "",
			Name:       "a",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer, err := dk.InsertNewBuyOffer(ctx, offer1)
		require.NoError(t, err)
		require.Equal(t, "1", offer.Id)

		offerGot := dk.GetBuyOffer(ctx, "1")
		require.NotNil(t, offerGot)

		offer1.Id = "1"
		require.Equal(t, offer1, *offerGot)

		dk.SetCountBuyOffer(ctx, 99)

		offer2 := dymnstypes.BuyOffer{
			Id:         "",
			Name:       "b",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}
		offer, err = dk.InsertNewBuyOffer(ctx, offer2)
		require.NoError(t, err)
		require.Equal(t, "100", offer.Id)

		offerGot = dk.GetBuyOffer(ctx, "100")
		require.NotNil(t, offerGot)

		offer2.Id = "100"
		require.Equal(t, offer2, *offerGot)
	})

	t.Run("can delete", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		var err error

		offer1 := dymnstypes.BuyOffer{
			Id:         "1",
			Name:       "a",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}
		err = dk.SetBuyOffer(ctx, offer1)
		require.NoError(t, err)

		offer2 := dymnstypes.BuyOffer{
			Id:         "2",
			Name:       "b",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(2),
		}
		err = dk.SetBuyOffer(ctx, offer2)
		require.NoError(t, err)

		offer3 := dymnstypes.BuyOffer{
			Id:         "3",
			Name:       "c",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(3),
		}
		err = dk.SetBuyOffer(ctx, offer3)
		require.NoError(t, err)

		offer4 := dymnstypes.BuyOffer{
			Id:         "4",
			Name:       "d",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(4),
		}
		err = dk.SetBuyOffer(ctx, offer4)
		require.NoError(t, err)

		require.NotNil(t, dk.GetBuyOffer(ctx, "1"))
		require.NotNil(t, dk.GetBuyOffer(ctx, "2"))
		require.NotNil(t, dk.GetBuyOffer(ctx, "3"))
		require.NotNil(t, dk.GetBuyOffer(ctx, "4"))

		dk.DeleteBuyOffer(ctx, "2")
		require.NotNil(t, dk.GetBuyOffer(ctx, "1"))
		require.Nil(t, dk.GetBuyOffer(ctx, "2"))
		require.NotNil(t, dk.GetBuyOffer(ctx, "3"))
		require.NotNil(t, dk.GetBuyOffer(ctx, "4"))

		dk.DeleteBuyOffer(ctx, "4")
		require.NotNil(t, dk.GetBuyOffer(ctx, "1"))
		require.Nil(t, dk.GetBuyOffer(ctx, "2"))
		require.NotNil(t, dk.GetBuyOffer(ctx, "3"))
		require.Nil(t, dk.GetBuyOffer(ctx, "4"))
	})

	t.Run("delete non-existing will not error", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.DeleteBuyOffer(ctx, "99999")
	})

	t.Run("event should be fired on set/insert offer", func(t *testing.T) {
		tests := []struct {
			name    string
			offer   dymnstypes.BuyOffer
			setFunc func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOffer)
		}{
			{
				name: "set offer",
				offer: dymnstypes.BuyOffer{
					Id:         "1",
					Name:       "a",
					Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				setFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOffer) {
					err := dk.SetBuyOffer(ctx, offer)
					require.NoError(t, err)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dk, _, _, ctx := testkeeper.DymNSKeeper(t)

				tt.setFunc(ctx, dk, tt.offer)

				events := ctx.EventManager().Events()
				require.NotEmpty(t, events)

				for _, event := range events {
					if event.Type != dymnstypes.EventTypeBuyOffer {
						continue
					}

					var actionName string
					for _, attr := range event.Attributes {
						if attr.Key == dymnstypes.AttributeKeyBoActionName {
							actionName = attr.Value
						}
					}
					require.NotEmpty(t, actionName, "event attr action name could not be found")
					require.Equalf(t,
						actionName, dymnstypes.AttributeValueBoActionNameSet,
						"event attr action name should be `%s`", dymnstypes.AttributeValueBoActionNameSet,
					)
					return
				}

				t.Errorf("event %s not found", dymnstypes.EventTypeBuyOffer)
			})
		}
	})

	t.Run("event should be fired on delete offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer := dymnstypes.BuyOffer{
			Id:         "1",
			Name:       "a",
			Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetBuyOffer(ctx, offer)
		require.NoError(t, err)

		ctx = ctx.WithEventManager(sdk.NewEventManager())

		dk.DeleteBuyOffer(ctx, offer.Id)

		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

		for _, event := range events {
			if event.Type != dymnstypes.EventTypeBuyOffer {
				continue
			}

			var actionName string
			for _, attr := range event.Attributes {
				if attr.Key == dymnstypes.AttributeKeyBoActionName {
					actionName = attr.Value
				}
			}
			require.NotEmpty(t, actionName, "event attr action name could not be found")
			require.Equalf(t,
				actionName, dymnstypes.AttributeValueBoActionNameDelete,
				"event attr action name should be `%s`", dymnstypes.AttributeValueBoActionNameDelete,
			)
			return
		}

		t.Errorf("event %s not found", dymnstypes.EventTypeBuyOffer)
	})
}

func TestKeeper_GetAllBuyOffers(t *testing.T) {
	// TODO DymNS: add test for Sell/Buy Alias

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyerA := testAddr(1).bech32()

	offer1 := dymnstypes.BuyOffer{
		Id:         "1",
		Name:       "a",
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err := dk.SetBuyOffer(ctx, offer1)
	require.NoError(t, err)

	offers := dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 1)
	require.Equal(t, offer1, offers[0])

	offer2 := dymnstypes.BuyOffer{
		Id:         "2",
		Name:       "a",
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOffer(ctx, offer2)
	require.NoError(t, err)

	offers = dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 2)
	require.Equal(t, []dymnstypes.BuyOffer{offer1, offer2}, offers)

	offer3 := dymnstypes.BuyOffer{
		Id:         "3",
		Name:       "b",
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(3),
	}
	err = dk.SetBuyOffer(ctx, offer3)
	require.NoError(t, err)

	offers = dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 3)
	require.Equal(t, []dymnstypes.BuyOffer{offer1, offer2, offer3}, offers)
}
