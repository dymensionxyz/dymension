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

func TestKeeper_IncreaseBuyOrdersCountAndGet(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Zero(t, dk.GetCountBuyOrders(ctx))

	count := dk.IncreaseBuyOrdersCountAndGet(ctx)
	require.Equal(t, uint64(1), count)
	require.Equal(t, uint64(1), dk.GetCountBuyOrders(ctx))

	count = dk.IncreaseBuyOrdersCountAndGet(ctx)
	require.Equal(t, uint64(2), count)
	require.Equal(t, uint64(2), dk.GetCountBuyOrders(ctx))

	count = dk.IncreaseBuyOrdersCountAndGet(ctx)
	require.Equal(t, uint64(3), count)
	require.Equal(t, uint64(3), dk.GetCountBuyOrders(ctx))

	dk.SetCountBuyOrders(ctx, math.MaxUint64-1)

	count = dk.IncreaseBuyOrdersCountAndGet(ctx)
	require.Equal(t, uint64(math.MaxUint64), count)
	require.Equal(t, uint64(math.MaxUint64), dk.GetCountBuyOrders(ctx))

	require.Panics(t, func() {
		dk.IncreaseBuyOrdersCountAndGet(ctx)
	}, "expect panic on overflow when increasing count of all-time buy offer records greater than uint64")
}

func TestKeeper_GetSetInsertNewBuyOrder(t *testing.T) {
	buyerA := testAddr(1).bech32()

	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	t.Run("get non-exists offer should returns nil", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		offer := dk.GetBuyOrder(ctx, "10183418")
		require.Nil(t, offer)
	})

	t.Run("should returns error when set empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, orderType := range supportedOrderTypes {
			err := dk.SetBuyOrder(ctx, dymnstypes.BuyOrder{
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

		offer1 := dymnstypes.BuyOrder{
			Id:         "101",
			GoodsId:    "my-name",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetBuyOrder(ctx, offer1)
		require.NoError(t, err)

		offerGot1 := dk.GetBuyOrder(ctx, offer1.Id)
		require.NotNil(t, offerGot1)

		require.Equal(t, offer1, *offerGot1)

		offer2 := dymnstypes.BuyOrder{
			Id:         "202",
			GoodsId:    "alias",
			Type:       dymnstypes.AliasOrder,
			Params:     []string{"rollapp_1-1"},
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err = dk.SetBuyOrder(ctx, offer2)
		require.NoError(t, err)

		offerGot2 := dk.GetBuyOrder(ctx, offer2.Id)
		require.NotNil(t, offerGot2)

		require.Equal(t, offer2, *offerGot2)

		// previous record should not be effected

		offerGot1 = dk.GetBuyOrder(ctx, offer1.Id)
		require.NotNil(t, offerGot1)

		require.NotEqual(t, *offerGot1, *offerGot2)
		require.Equal(t, offer1, *offerGot1)
	})

	t.Run("set omits params if empty", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1 := dymnstypes.BuyOrder{
			Id:         "101",
			GoodsId:    "my-name",
			Type:       dymnstypes.NameOrder,
			Params:     []string{},
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetBuyOrder(ctx, offer1)
		require.NoError(t, err)

		offerGot1 := dk.GetBuyOrder(ctx, offer1.Id)
		require.NotNil(t, offerGot1)

		require.Nil(t, offerGot1.Params)
	})

	t.Run("should panic when insert non-empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, orderType := range supportedOrderTypes {
			require.Panics(t, func() {
				_, _ = dk.InsertNewBuyOrder(ctx, dymnstypes.BuyOrder{
					Id:         dymnstypes.CreateBuyOrderId(orderType, 1),
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

		offer1a := dymnstypes.BuyOrder{
			Id:         "",
			GoodsId:    "my-name",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer1b, err := dk.InsertNewBuyOrder(ctx, offer1a)
		require.NoError(t, err)
		require.Equal(t, "101", offer1b.Id)

		offerGot1 := dk.GetBuyOrder(ctx, "101")
		require.NotNil(t, offerGot1)

		offer1a.Id = offer1b.Id
		require.Equal(t, offer1a, *offerGot1)

		offer2a := dymnstypes.BuyOrder{
			Id:         "",
			GoodsId:    "alias",
			Type:       dymnstypes.AliasOrder,
			Params:     []string{"rollapp_1-1"},
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer2b, err := dk.InsertNewBuyOrder(ctx, offer2a)
		require.NoError(t, err)
		require.Equal(t, "202", offer2b.Id)

		offerGot2 := dk.GetBuyOrder(ctx, "202")
		require.NotNil(t, offerGot2)

		offer2a.Id = offer2b.Id
		require.Equal(t, offer2a, *offerGot2)

		// previous record should not be effected

		offerGot1 = dk.GetBuyOrder(ctx, "101")
		require.NotNil(t, offerGot1)

		require.NotEqual(t, *offerGot1, *offerGot2)
		require.Equal(t, offer1a, *offerGot1)
	})

	t.Run("can not insert duplicated", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			dk.SetCountBuyOrders(ctx, 1)
			const nextId uint64 = 2

			var params []string
			if orderType == dymnstypes.AliasOrder {
				params = []string{"rollapp_1-1"}
			}

			existing := dymnstypes.BuyOrder{
				Id:         dymnstypes.CreateBuyOrderId(orderType, nextId),
				GoodsId:    "goods",
				Type:       orderType,
				Params:     params,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			err := dk.SetBuyOrder(ctx, existing)
			require.NoError(t, err)

			offer := dymnstypes.BuyOrder{
				Id:         "",
				GoodsId:    "goods",
				Type:       orderType,
				Params:     params,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			_, err = dk.InsertNewBuyOrder(ctx, offer)
			require.Error(t, err)
			require.Contains(t, err.Error(), "Buy-Order ID already exists")
		}
	})

	t.Run("should automatically fill ID when insert", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			var params []string
			if orderType == dymnstypes.AliasOrder {
				params = []string{"rollapp_1-1"}
			}

			offer1 := dymnstypes.BuyOrder{
				Id:         "",
				GoodsId:    "one",
				Type:       orderType,
				Params:     params,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			offer, err := dk.InsertNewBuyOrder(ctx, offer1)
			require.NoError(t, err)

			wantId1 := dymnstypes.CreateBuyOrderId(orderType, 1)
			require.Equal(t, wantId1, offer.Id)

			offerGot := dk.GetBuyOrder(ctx, wantId1)
			require.NotNil(t, offerGot)

			offer1.Id = wantId1
			require.Equal(t, offer1, *offerGot)

			dk.SetCountBuyOrders(ctx, 99)

			offer2 := dymnstypes.BuyOrder{
				Id:         "",
				GoodsId:    "two",
				Type:       orderType,
				Params:     params,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}
			offer, err = dk.InsertNewBuyOrder(ctx, offer2)
			require.NoError(t, err)

			wantId2 := dymnstypes.CreateBuyOrderId(orderType, 100)
			require.Equal(t, wantId2, offer.Id)

			offerGot = dk.GetBuyOrder(ctx, wantId2)
			require.NotNil(t, offerGot)

			offer2.Id = wantId2
			require.Equal(t, offer2, *offerGot)
		}
	})

	t.Run("can delete", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		var err error

		offer1 := dymnstypes.BuyOrder{
			Id:         "101",
			GoodsId:    "a",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(1),
		}
		err = dk.SetBuyOrder(ctx, offer1)
		require.NoError(t, err)

		offer2 := dymnstypes.BuyOrder{
			Id:         "202",
			GoodsId:    "b",
			Type:       dymnstypes.AliasOrder,
			Params:     []string{"rollapp_1-1"},
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(2),
		}
		err = dk.SetBuyOrder(ctx, offer2)
		require.NoError(t, err)

		offer3 := dymnstypes.BuyOrder{
			Id:         "103",
			GoodsId:    "c",
			Type:       dymnstypes.NameOrder,
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(3),
		}
		err = dk.SetBuyOrder(ctx, offer3)
		require.NoError(t, err)

		offer4 := dymnstypes.BuyOrder{
			Id:         "204",
			GoodsId:    "d",
			Type:       dymnstypes.AliasOrder,
			Params:     []string{"rollapp_2-2"},
			Buyer:      buyerA,
			OfferPrice: dymnsutils.TestCoin(4),
		}
		err = dk.SetBuyOrder(ctx, offer4)
		require.NoError(t, err)

		require.NotNil(t, dk.GetBuyOrder(ctx, offer1.Id))
		require.NotNil(t, dk.GetBuyOrder(ctx, offer2.Id))
		require.NotNil(t, dk.GetBuyOrder(ctx, offer3.Id))
		require.NotNil(t, dk.GetBuyOrder(ctx, offer4.Id))

		dk.DeleteBuyOrder(ctx, offer2.Id)
		require.NotNil(t, dk.GetBuyOrder(ctx, offer1.Id))
		require.Nil(t, dk.GetBuyOrder(ctx, offer2.Id))
		require.NotNil(t, dk.GetBuyOrder(ctx, offer3.Id))
		require.NotNil(t, dk.GetBuyOrder(ctx, offer4.Id))

		dk.DeleteBuyOrder(ctx, offer4.Id)
		require.NotNil(t, dk.GetBuyOrder(ctx, offer1.Id))
		require.Nil(t, dk.GetBuyOrder(ctx, offer2.Id))
		require.NotNil(t, dk.GetBuyOrder(ctx, offer3.Id))
		require.Nil(t, dk.GetBuyOrder(ctx, offer4.Id))

		dk.DeleteBuyOrder(ctx, offer3.Id)
		require.NotNil(t, dk.GetBuyOrder(ctx, offer1.Id))
		require.Nil(t, dk.GetBuyOrder(ctx, offer2.Id))
		require.Nil(t, dk.GetBuyOrder(ctx, offer3.Id))
		require.Nil(t, dk.GetBuyOrder(ctx, offer4.Id))
	})

	t.Run("delete non-existing will not panics", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.DeleteBuyOrder(ctx, "1099999")
		dk.DeleteBuyOrder(ctx, "2099999")
	})

	t.Run("event should be fired on set/insert offer", func(t *testing.T) {
		tests := []struct {
			name    string
			offer   dymnstypes.BuyOrder
			setFunc func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOrder)
		}{
			{
				name: "set offer type Dym-Name",
				offer: dymnstypes.BuyOrder{
					Id:         "101",
					GoodsId:    "my-name",
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				setFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOrder) {
					err := dk.SetBuyOrder(ctx, offer)
					require.NoError(t, err)
				},
			},
			{
				name: "set offer type Alias",
				offer: dymnstypes.BuyOrder{
					Id:         "201",
					GoodsId:    "alias",
					Type:       dymnstypes.AliasOrder,
					Params:     []string{"rollapp_1-1"},
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(1),
				},
				setFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.BuyOrder) {
					err := dk.SetBuyOrder(ctx, offer)
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
					if event.Type != dymnstypes.EventTypeBuyOrder {
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

				t.Errorf("event %s not found", dymnstypes.EventTypeBuyOrder)
			})
		}
	})

	t.Run("event should be fired on delete offer", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			offer := dymnstypes.BuyOrder{
				Id:         dymnstypes.CreateBuyOrderId(orderType, 1),
				GoodsId:    "goods",
				Type:       orderType,
				Buyer:      buyerA,
				OfferPrice: dymnsutils.TestCoin(1),
			}

			err := dk.SetBuyOrder(ctx, offer)
			require.NoError(t, err)

			ctx = ctx.WithEventManager(sdk.NewEventManager())

			dk.DeleteBuyOrder(ctx, offer.Id)

			events := ctx.EventManager().Events()
			require.NotEmpty(t, events)

			for _, event := range events {
				if event.Type != dymnstypes.EventTypeBuyOrder {
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

			t.Errorf("event %s not found", dymnstypes.EventTypeBuyOrder)
		}
	})
}

func TestKeeper_GetAllBuyOrders(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyerA := testAddr(1).bech32()

	offer1 := dymnstypes.BuyOrder{
		Id:         "101",
		GoodsId:    "a",
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err := dk.SetBuyOrder(ctx, offer1)
	require.NoError(t, err)

	offers := dk.GetAllBuyOrders(ctx)
	require.Len(t, offers, 1)
	require.Equal(t, offer1, offers[0])

	offer2 := dymnstypes.BuyOrder{
		Id:         "202",
		GoodsId:    "b",
		Type:       dymnstypes.AliasOrder,
		Params:     []string{"rollapp_1-1"},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOrder(ctx, offer2)
	require.NoError(t, err)

	offers = dk.GetAllBuyOrders(ctx)
	require.Len(t, offers, 2)
	require.Equal(t, []dymnstypes.BuyOrder{offer1, offer2}, offers)

	offer3 := dymnstypes.BuyOrder{
		Id:         "103",
		GoodsId:    "a",
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOrder(ctx, offer3)
	require.NoError(t, err)

	offers = dk.GetAllBuyOrders(ctx)
	require.Len(t, offers, 3)
	require.Equal(t, []dymnstypes.BuyOrder{
		offer1, offer3, // <= Dym-Name Buy-Order should be sorted first
		offer2, // <= Alias Buy-Order should be sorted second
		// because of store branched by order type
	}, offers)

	offer4 := dymnstypes.BuyOrder{
		Id:         "204",
		GoodsId:    "b",
		Type:       dymnstypes.AliasOrder,
		Params:     []string{"rollapp_2-2"},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetBuyOrder(ctx, offer4)
	require.NoError(t, err)

	offers = dk.GetAllBuyOrders(ctx)
	require.Len(t, offers, 4)
	require.Equal(t, []dymnstypes.BuyOrder{offer1, offer3, offer2, offer4}, offers)

	offer5 := dymnstypes.BuyOrder{
		Id:         "105",
		GoodsId:    "b",
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(3),
	}
	err = dk.SetBuyOrder(ctx, offer5)
	require.NoError(t, err)

	offers = dk.GetAllBuyOrders(ctx)
	require.Len(t, offers, 5)
	require.Equal(t, []dymnstypes.BuyOrder{offer1, offer3, offer5, offer2, offer4}, offers)
}
