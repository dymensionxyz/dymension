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
	buyerA := testAddr(1).bech32()

	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	t.Run("get non-exists offer should returns nil", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		offer := dk.GetBuyOffer(ctx, "10183418")
		require.Nil(t, offer)
	})

	t.Run("should returns error when set empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, orderType := range supportedOrderTypes {
			err := dk.SetBuyOffer(ctx, dymnstypes.BuyOffer{
				Id:         "",
				GoodsId:    "goods",
				Type:       orderType,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), "ID of offer is empty")
		}
	})

	t.Run("can set and can get", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1 := dymnstypes.BuyOffer{
			Id:         "101",
			GoodsId:    "my-name",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetBuyOffer(ctx, offer1)
		require.NoError(t, err)

		offerGot1 := dk.GetBuyOffer(ctx, "101")
		require.NotNil(t, offerGot1)

		require.Equal(t, offer1, *offerGot1)

		offer2 := dymnstypes.BuyOffer{
			Id:         "202",
			GoodsId:    "alias",
			Type:       dymnstypes.AliasOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err = dk.SetBuyOffer(ctx, offer2)
		require.NoError(t, err)

		offerGot2 := dk.GetBuyOffer(ctx, "202")
		require.NotNil(t, offerGot2)

		require.Equal(t, offer2, *offerGot2)

		// previous record should not be effected

		offerGot1 = dk.GetBuyOffer(ctx, "101")
		require.NotNil(t, offerGot1)

		require.NotEqual(t, *offerGot1, *offerGot2)
		require.Equal(t, offer1, *offerGot1)
	})

	t.Run("should panic when insert non-empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, orderType := range supportedOrderTypes {
			require.Panics(t, func() {
				_, _ = dk.InsertNewBuyOffer(ctx, dymnstypes.BuyOffer{
					Id:         dymnstypes.CreateBuyOfferId(orderType, 1),
					GoodsId:    "goods",
					Type:       orderType,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				})
			})
		}
	})

	t.Run("can insert", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1a := dymnstypes.BuyOffer{
			Id:         "",
			GoodsId:    "my-name",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer1b, err := dk.InsertNewBuyOffer(ctx, offer1a)
		require.NoError(t, err)
		require.Equal(t, "101", offer1b.Id)

		offerGot1 := dk.GetBuyOffer(ctx, "101")
		require.NotNil(t, offerGot1)

		offer1a.Id = "101"
		require.Equal(t, offer1a, *offerGot1)

		offer2a := dymnstypes.BuyOffer{
			Id:         "",
			GoodsId:    "alias",
			Type:       dymnstypes.AliasOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer2b, err := dk.InsertNewBuyOffer(ctx, offer2a)
		require.NoError(t, err)
		require.Equal(t, "202", offer2b.Id)

		offerGot2 := dk.GetBuyOffer(ctx, "202")
		require.NotNil(t, offerGot2)

		offer2a.Id = "202"
		require.Equal(t, offer2a, *offerGot2)

		// previous record should not be effected

		offerGot1 = dk.GetBuyOffer(ctx, "101")
		require.NotNil(t, offerGot1)

		require.NotEqual(t, *offerGot1, *offerGot2)
		require.Equal(t, offer1b, *offerGot1)
	})

	t.Run("can not insert duplicated", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			dk.SetCountBuyOffer(ctx, 1)
			const nextId uint64 = 2

			existing := dymnstypes.BuyOffer{
				Id:         dymnstypes.CreateBuyOfferId(orderType, nextId),
				GoodsId:    "goods",
				Type:       orderType,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			err := dk.SetBuyOffer(ctx, existing)
			require.NoError(t, err)

			offer := dymnstypes.BuyOffer{
				Id:         "",
				GoodsId:    "goods",
				Type:       orderType,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			_, err = dk.InsertNewBuyOffer(ctx, offer)
			require.Error(t, err)
			require.Contains(t, err.Error(), "Buy-Order-ID already exists")
		}
	})

	t.Run("should automatically fill ID when insert", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			offer1 := dymnstypes.BuyOffer{
				Id:         "",
				GoodsId:    "one",
				Type:       orderType,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			offer, err := dk.InsertNewBuyOffer(ctx, offer1)
			require.NoError(t, err)

			wantId1 := dymnstypes.CreateBuyOfferId(orderType, 1)
			require.Equal(t, wantId1, offer.Id)

			offerGot := dk.GetBuyOffer(ctx, wantId1)
			require.NotNil(t, offerGot)

			offer1.Id = wantId1
			require.Equal(t, offer1, *offerGot)

			dk.SetCountBuyOffer(ctx, 99)

			offer2 := dymnstypes.BuyOffer{
				Id:         "",
				GoodsId:    "two",
				Type:       orderType,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}
			offer, err = dk.InsertNewBuyOffer(ctx, offer2)
			require.NoError(t, err)

			wantId2 := dymnstypes.CreateBuyOfferId(orderType, 100)
			require.Equal(t, wantId2, offer.Id)

			offerGot = dk.GetBuyOffer(ctx, wantId2)
			require.NotNil(t, offerGot)

			offer2.Id = wantId2
			require.Equal(t, offer2, *offerGot)
		}
	})

	t.Run("can delete", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		var err error

		offer1 := dymnstypes.BuyOffer{
			Id:         "101",
			GoodsId:    "a",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}
		err = dk.SetBuyOffer(ctx, offer1)
		require.NoError(t, err)

		offer2 := dymnstypes.BuyOffer{
			Id:         "202",
			GoodsId:    "b",
			Type:       dymnstypes.AliasOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(2),
		}
		err = dk.SetBuyOffer(ctx, offer2)
		require.NoError(t, err)

		offer3 := dymnstypes.BuyOffer{
			Id:         "103",
			GoodsId:    "c",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(3),
		}
		err = dk.SetBuyOffer(ctx, offer3)
		require.NoError(t, err)

		offer4 := dymnstypes.BuyOffer{
			Id:         "204",
			GoodsId:    "d",
			Type:       dymnstypes.AliasOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(4),
		}
		err = dk.SetBuyOffer(ctx, offer4)
		require.NoError(t, err)

		require.NotNil(t, dk.GetBuyOffer(ctx, offer1.Id))
		require.NotNil(t, dk.GetBuyOffer(ctx, offer2.Id))
		require.NotNil(t, dk.GetBuyOffer(ctx, offer3.Id))
		require.NotNil(t, dk.GetBuyOffer(ctx, offer4.Id))

		dk.DeleteBuyOffer(ctx, offer2.Id)
		require.NotNil(t, dk.GetBuyOffer(ctx, offer1.Id))
		require.Nil(t, dk.GetBuyOffer(ctx, offer2.Id))
		require.NotNil(t, dk.GetBuyOffer(ctx, offer3.Id))
		require.NotNil(t, dk.GetBuyOffer(ctx, offer4.Id))

		dk.DeleteBuyOffer(ctx, offer4.Id)
		require.NotNil(t, dk.GetBuyOffer(ctx, offer1.Id))
		require.Nil(t, dk.GetBuyOffer(ctx, offer2.Id))
		require.NotNil(t, dk.GetBuyOffer(ctx, offer3.Id))
		require.Nil(t, dk.GetBuyOffer(ctx, offer4.Id))

		dk.DeleteBuyOffer(ctx, offer3.Id)
		require.NotNil(t, dk.GetBuyOffer(ctx, offer1.Id))
		require.Nil(t, dk.GetBuyOffer(ctx, offer2.Id))
		require.Nil(t, dk.GetBuyOffer(ctx, offer3.Id))
		require.Nil(t, dk.GetBuyOffer(ctx, offer4.Id))
	})

	t.Run("delete non-existing will not panics", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.DeleteBuyOffer(ctx, "1099999")
		dk.DeleteBuyOffer(ctx, "2099999")
	})

	t.Run("event should be fired on set/insert offer", func(t *testing.T) {
		tests := []struct {
			name    string
			offer   dymnstypes.BuyOffer
			setFunc func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOffer)
		}{
			{
				name: "set offer type Dym-Name",
				offer: dymnstypes.BuyOffer{
					Id:         "101",
					GoodsId:    "my-name",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				setFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOffer) {
					err := dk.SetBuyOffer(ctx, offer)
					require.NoError(t, err)
				},
			},
			{
				name: "set offer type Alias",
				offer: dymnstypes.BuyOffer{
					Id:         "201",
					GoodsId:    "alias",
					Type:       dymnstypes.AliasOrder,
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
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			offer := dymnstypes.BuyOffer{
				Id:         dymnstypes.CreateBuyOfferId(orderType, 1),
				GoodsId:    "goods",
				Type:       orderType,
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
		}
	})
}

func TestKeeper_GetAllBuyOffers(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyerA := testAddr(1).bech32()

	offer1 := dymnstypes.BuyOffer{
		Id:         "101",
		GoodsId:    "a",
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err := dk.SetBuyOffer(ctx, offer1)
	require.NoError(t, err)

	offers := dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 1)
	require.Equal(t, offer1, offers[0])

	offer2 := dymnstypes.BuyOffer{
		Id:         "202",
		GoodsId:    "b",
		Type:       dymnstypes.AliasOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOffer(ctx, offer2)
	require.NoError(t, err)

	offers = dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 2)
	require.Equal(t, []dymnstypes.BuyOffer{offer1, offer2}, offers)

	offer3 := dymnstypes.BuyOffer{
		Id:         "103",
		GoodsId:    "a",
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOffer(ctx, offer3)
	require.NoError(t, err)

	offers = dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 3)
	require.Equal(t, []dymnstypes.BuyOffer{
		offer1, offer3, // <= Dym-Name Buy-Order should be sorted first
		offer2, // <= Alias Buy-Order should be sorted second
		// because of store branched by order type
	}, offers)

	offer4 := dymnstypes.BuyOffer{
		Id:         "204",
		GoodsId:    "b",
		Type:       dymnstypes.AliasOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOffer(ctx, offer4)
	require.NoError(t, err)

	offers = dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 4)
	require.Equal(t, []dymnstypes.BuyOffer{offer1, offer3, offer2, offer4}, offers)

	offer5 := dymnstypes.BuyOffer{
		Id:         "105",
		GoodsId:    "b",
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(3),
	}
	err = dk.SetBuyOffer(ctx, offer5)
	require.NoError(t, err)

	offers = dk.GetAllBuyOffers(ctx)
	require.Len(t, offers, 5)
	require.Equal(t, []dymnstypes.BuyOffer{offer1, offer3, offer5, offer2, offer4}, offers)
}
