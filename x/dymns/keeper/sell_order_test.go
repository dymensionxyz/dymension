package keeper_test

import (
	cryptorand "crypto/rand"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
	"testing"
	"time"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetSetSellOrder(t *testing.T) {
	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	t.Run("reject invalid SO", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{})
		require.Error(t, err)
	})

	t.Run("can set", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				GoodsId:  "goods",
				Type:     orderType,
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(100),
			}
			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)
		}
	})

	t.Run("event should be fired on set", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				GoodsId:  "goods",
				Type:     orderType,
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(100),
			}
			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			events := ctx.EventManager().Events()
			require.NotEmpty(t, events)

			func(events sdk.Events) {
				for _, event := range events {
					if event.Type != dymnstypes.EventTypeSellOrder {
						continue
					}

					var actionName string
					for _, attr := range event.Attributes {
						if attr.Key == dymnstypes.AttributeKeySoActionName {
							actionName = attr.Value
						}
					}
					require.NotEmpty(t, actionName, "event attr action name could not be found")
					require.Equalf(t,
						actionName, dymnstypes.AttributeValueSoActionNameSet,
						"event attr action name should be `%s`", dymnstypes.AttributeValueSoActionNameSet,
					)
					return
				}

				t.Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
			}(events)
		}
	})

	t.Run("can not set Sell-Order with unknown type", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{
			GoodsId:  "goods",
			Type:     dymnstypes.OrderType_OT_UNKNOWN,
			ExpireAt: 1,
			MinPrice: dymnsutils.TestCoin(100),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid SO type")
	})

	t.Run("non-exists returns nil", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, orderType := range supportedOrderTypes {
			require.Nil(t, dk.GetSellOrder(ctx, "goods", orderType))
		}
	})

	t.Run("omit Sell Price if zero", func(t *testing.T) {
		for _, sellPrice := range []*sdk.Coin{nil, dymnsutils.TestCoinP(0), {}} {
			for _, orderType := range supportedOrderTypes {
				dk, _, _, ctx := testkeeper.DymNSKeeper(t)

				err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{
					GoodsId:   "goods",
					Type:      orderType,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(100),
					SellPrice: sellPrice,
				})
				require.NoError(t, err)

				require.Nil(t, dk.GetSellOrder(ctx, "goods", orderType).SellPrice)
			}
		}
	})

	t.Run("get returns correct inserted record, regardless type", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		var sellOrders []dymnstypes.SellOrder

		const seed int = 100

		for i := 0; i < seed; i++ {
			for _, orderType := range supportedOrderTypes {

				so := dymnstypes.SellOrder{
					GoodsId:  fmt.Sprintf("dog%d", i), // same goods id for all types
					Type:     orderType,
					ExpireAt: 1 + int64(i),
					MinPrice: dymnsutils.TestCoin(int64(seed + i)),
				}
				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)

				sellOrders = append(sellOrders, so)
			}
		}

		for _, so := range sellOrders {
			got := dk.GetSellOrder(ctx, so.GoodsId, so.Type)
			require.NotNil(t, got)
			require.Equal(t, so, *got)
		}
	})
}

func TestKeeper_DeleteSellOrder(t *testing.T) {
	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	t.Run("can delete", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				GoodsId:  "goods",
				Type:     orderType,
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(1),
			}

			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			require.NotNil(t, dk.GetSellOrder(ctx, so.GoodsId, so.Type))

			dk.DeleteSellOrder(ctx, so.GoodsId, so.Type)

			require.Nil(t, dk.GetSellOrder(ctx, so.GoodsId, so.Type))
		}
	})

	t.Run("event should be fired upon deletion", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				GoodsId:  "goods",
				Type:     orderType,
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(1),
			}

			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			dk.DeleteSellOrder(ctx, so.GoodsId, so.Type)

			events := ctx.EventManager().Events()
			require.NotEmpty(t, events)

			func(events sdk.Events) {
				for _, event := range events {
					if event.Type != dymnstypes.EventTypeSellOrder {
						continue
					}

					var actionName string
					for _, attr := range event.Attributes {
						if attr.Key == dymnstypes.AttributeKeySoActionName {
							actionName = attr.Value
						}
					}
					require.NotEmpty(t, actionName, "event attr action name could not be found")
					require.Equalf(t,
						actionName, dymnstypes.AttributeValueSoActionNameSet,
						"event attr action name should be `%s`", dymnstypes.AttributeValueSoActionNameDelete,
					)
					return
				}

				t.Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
			}(events)
		}
	})

	t.Run("delete remove the correct record regardless type", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		type testCase struct {
			so      dymnstypes.SellOrder
			deleted bool
		}

		var testCases []*testCase

		// build testcases
		const seed int = 100
		for i := 0; i < seed; i++ {
			for j, orderType := range supportedOrderTypes {
				so := dymnstypes.SellOrder{
					GoodsId:  fmt.Sprintf("dog%03d%d", i, j),
					Type:     orderType,
					ExpireAt: 1,
					MinPrice: dymnsutils.TestCoin(1),
				}

				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)

				require.NotNil(t, dk.GetSellOrder(ctx, so.GoodsId, so.Type))

				testCases = append(testCases, &testCase{so: so, deleted: false})
			}
		}

		require.Len(t, testCases, seed*len(supportedOrderTypes))
		require.Len(t, dk.GetAllSellOrders(ctx), len(testCases))

		// test delete
		for i, tc := range testCases {
			dk.DeleteSellOrder(ctx, tc.so.GoodsId, tc.so.Type)
			tc.deleted = true
			require.Nil(t, dk.GetSellOrder(ctx, tc.so.GoodsId, tc.so.Type))

			require.Len(t, dk.GetAllSellOrders(ctx), len(testCases)-(i+1))

			for _, tc2 := range testCases {
				if tc2.deleted {
					require.Nil(t, dk.GetSellOrder(ctx, tc2.so.GoodsId, tc2.so.Type))
				} else {
					require.NotNil(t, dk.GetSellOrder(ctx, tc2.so.GoodsId, tc2.so.Type))
				}
			}
		}
	})
}

func TestKeeper_MoveSellOrderToHistorical(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	ownerA := testAddr(1).bech32()
	bidderA := testAddr(2).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}
	err := dk.SetDymName(ctx, dymName1)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}
	err = dk.SetDymName(ctx, dymName2)
	require.NoError(t, err)

	dymNames := dk.GetAllNonExpiredDymNames(ctx)
	require.Len(t, dymNames, 2)

	so11 := dymnstypes.SellOrder{
		GoodsId:   dymName1.Name,
		Type:      dymnstypes.NameOrder,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so11)
	require.NoError(t, err)

	alias2 := "alias"
	so21 := dymnstypes.SellOrder{
		GoodsId:  alias2,
		Type:     dymnstypes.AliasOrder,
		ExpireAt: 2,
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so21)
	require.NoError(t, err)

	alias3 := "salas"
	so3 := dymnstypes.SellOrder{
		GoodsId:  alias3,
		Type:     dymnstypes.AliasOrder,
		ExpireAt: 3,
		MinPrice: dymnsutils.TestCoin(222),
	}
	err = dk.SetSellOrder(ctx, so3)
	require.NoError(t, err)

	t.Run("should able to move", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so11.GoodsId, so11.Type)
		require.NoError(t, err)

		err = dk.MoveSellOrderToHistorical(ctx, so21.GoodsId, so21.Type)
		require.NoError(t, err)
	})

	t.Run("moved SO should be removed from active", func(t *testing.T) {
		require.Nil(t, dk.GetSellOrder(ctx, so11.GoodsId, so11.Type))
		require.Nil(t, dk.GetSellOrder(ctx, so21.GoodsId, so21.Type))
	})

	t.Run("has min expiry mapping", func(t *testing.T) {
		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, so11.GoodsId, so11.Type)
		require.True(t, found)
		require.Equal(t, so11.ExpireAt, minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, so21.GoodsId, so21.Type)
		require.True(t, found)
		require.Equal(t, so21.ExpireAt, minExpiry)
	})

	t.Run("should not move non-exists", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, "non-exists", dymnstypes.NameOrder)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order: Dym-Name: non-exists")

		err = dk.MoveSellOrderToHistorical(ctx, "non-exists", dymnstypes.AliasOrder)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order: Alias: non-exists")
	})

	t.Run("should able to move a duplicated without error", func(t *testing.T) {
		defer func() {
			dk.DeleteSellOrder(ctx, so11.GoodsId, so11.Type)
			dk.DeleteSellOrder(ctx, so21.GoodsId, so21.Type)
		}()

		for _, so := range []dymnstypes.SellOrder{so11, so21} {
			err = dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			err = dk.MoveSellOrderToHistorical(ctx, so.GoodsId, so.Type)
			require.NoError(t, err)

			list := dk.GetHistoricalSellOrders(ctx, so.GoodsId, so.Type)
			require.Len(t, list, 1, "do not persist duplicated historical SO")
		}
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Empty(t, dk.GetSellOrder(ctx, dymName2.Name, dymnstypes.NameOrder))
		require.NotNil(t, dk.GetSellOrder(ctx, alias3, dymnstypes.AliasOrder))
	})

	so4 := dymnstypes.SellOrder{
		GoodsId:  dymName2.Name,
		Type:     dymnstypes.NameOrder,
		ExpireAt: 1,
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so4)
	require.NoError(t, err)

	t.Run("should able to move", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so4.GoodsId, so4.Type)
		require.NoError(t, err)
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.NameOrder), 1)
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder), 1)
		require.Len(t, dk.GetHistoricalSellOrders(ctx, alias2, dymnstypes.AliasOrder), 1)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder))
	})

	so12 := dymnstypes.SellOrder{
		GoodsId:   dymName1.Name,
		Type:      dymnstypes.NameOrder,
		ExpireAt:  now.Unix() + 1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so12)
	require.NoError(t, err)

	so22 := dymnstypes.SellOrder{
		GoodsId:   alias2,
		Type:      dymnstypes.AliasOrder,
		ExpireAt:  now.Unix() + 1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so22)
	require.NoError(t, err)

	t.Run("should not move yet finished SO", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so12.GoodsId, so12.Type)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order not yet expired")

		err = dk.MoveSellOrderToHistorical(ctx, so22.GoodsId, so22.Type)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order not yet expired")
	})

	for _, so := range []dymnstypes.SellOrder{so12, so22} {
		so.HighestBid = &dymnstypes.SellOrderBid{
			Bidder: bidderA,
			Price:  dymnsutils.TestCoin(300),
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		t.Run("should able to move finished SO", func(t *testing.T) {
			err := dk.MoveSellOrderToHistorical(ctx, so.GoodsId, so.Type)
			require.NoError(t, err)

			list := dk.GetHistoricalSellOrders(ctx, so.GoodsId, so.Type)
			require.Len(t, list, 2, "should appended to historical")
		})
	}

	minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, so12.GoodsId, so12.Type)
	require.True(t, found)
	require.Equal(t, so11.ExpireAt, minExpiry, "should keep the minimum")
	require.NotEqual(t, so12.ExpireAt, minExpiry, "should keep the minimum")

	minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, so22.GoodsId, so22.Type)
	require.True(t, found)
	require.Equal(t, so21.ExpireAt, minExpiry, "should keep the minimum")
	require.NotEqual(t, so22.ExpireAt, minExpiry, "should keep the minimum")

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder), 1)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder))
	})
}

func TestKeeper_GetAndDeleteHistoricalSellOrders(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	ownerA := testAddr(1).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}
	err := dk.SetDymName(ctx, dymName1)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
	}
	err = dk.SetDymName(ctx, dymName2)
	require.NoError(t, err)

	alias3 := "alias"
	alias4 := "salas"

	t.Run("getting non-exists should returns empty", func(t *testing.T) {
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.NameOrder))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias4, dymnstypes.AliasOrder))
	})

	so11 := dymnstypes.SellOrder{
		GoodsId:   dymName1.Name,
		Type:      dymnstypes.NameOrder,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so11)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so11.GoodsId, so11.Type)
	require.NoError(t, err)

	so2 := dymnstypes.SellOrder{
		GoodsId:  alias3,
		Type:     dymnstypes.AliasOrder,
		ExpireAt: 7,
		MinPrice: dymnsutils.TestCoin(200),
	}
	err = dk.SetSellOrder(ctx, so2)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so2.GoodsId, so2.Type)
	require.NoError(t, err)

	so3 := dymnstypes.SellOrder{
		GoodsId:  dymName2.Name,
		Type:     dymnstypes.NameOrder,
		ExpireAt: 1,
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so3)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so3.GoodsId, so3.Type)
	require.NoError(t, err)

	so3.ExpireAt++
	err = dk.SetSellOrder(ctx, so3)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so3.GoodsId, so3.Type)
	require.NoError(t, err)

	so4 := dymnstypes.SellOrder{
		GoodsId:  alias4,
		Type:     dymnstypes.AliasOrder,
		ExpireAt: 5,
		MinPrice: dymnsutils.TestCoin(500),
	}
	err = dk.SetSellOrder(ctx, so4)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so4.GoodsId, so4.Type)
	require.NoError(t, err)

	t.Run("fetch correctly", func(t *testing.T) {
		list1 := dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.NameOrder)
		require.Len(t, list1, 1)
		list2 := dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder)
		require.Len(t, list2, 2)
		list3 := dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder)
		require.Len(t, list3, 1)
		list4 := dk.GetHistoricalSellOrders(ctx, alias4, dymnstypes.AliasOrder)
		require.Len(t, list4, 1)

		require.Equal(t, so3.GoodsId, list2[0].GoodsId)
		require.Equal(t, so3.GoodsId, list2[1].GoodsId)

		require.Equal(t, int64(1), list2[0].ExpireAt)
		require.Equal(t, int64(2), list2[1].ExpireAt)

		require.Equal(t, int64(7), list3[0].ExpireAt)
		require.Equal(t, int64(5), list4[0].ExpireAt)
	})

	t.Run("delete", func(t *testing.T) {
		dk.DeleteHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.NameOrder)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.NameOrder))

		list2 := dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder)
		require.Len(t, list2, 2)

		dk.DeleteHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.NameOrder))

		list3 := dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder)
		require.Len(t, list3, 1)

		dk.DeleteHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.AliasOrder))

		list4 := dk.GetHistoricalSellOrders(ctx, alias4, dymnstypes.AliasOrder)
		require.Len(t, list4, 1)
	})
}

func TestKeeper_GetSetActiveSellOrdersExpiration(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	t.Run("get", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			aSoe := dk.GetActiveSellOrdersExpiration(ctx, orderType)
			require.Empty(t, aSoe.Records, "default list must be empty")
			require.NotNil(t, aSoe.Records, "list must be initialized")
		}
	})

	t.Run("set", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			aSoe := &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						GoodsId:  "hello",
						ExpireAt: 123,
					},
					{
						GoodsId:  "world",
						ExpireAt: 456,
					},
				},
			}
			err := dk.SetActiveSellOrdersExpiration(ctx, aSoe, orderType)
			require.NoError(t, err)

			aSoe = dk.GetActiveSellOrdersExpiration(ctx, orderType)
			require.Len(t, aSoe.Records, 2)
			require.Equal(t, "hello", aSoe.Records[0].GoodsId)
			require.Equal(t, int64(123), aSoe.Records[0].ExpireAt)
			require.Equal(t, "world", aSoe.Records[1].GoodsId)
			require.Equal(t, int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	t.Run("must automatically sort when set", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						GoodsId:  "bbb",
						ExpireAt: 456,
					},
					{
						GoodsId:  "aaa",
						ExpireAt: 123,
					},
				},
			}, orderType)
			require.NoError(t, err)

			aSoe := dk.GetActiveSellOrdersExpiration(ctx, orderType)
			require.Len(t, aSoe.Records, 2)

			require.Equal(t, "aaa", aSoe.Records[0].GoodsId)
			require.Equal(t, int64(123), aSoe.Records[0].ExpireAt)
			require.Equal(t, "bbb", aSoe.Records[1].GoodsId)
			require.Equal(t, int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	t.Run("can not set if set is not valid", func(t *testing.T) {
		for _, orderType := range supportedOrderTypes {
			// not unique
			err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						GoodsId:  "dup",
						ExpireAt: 456,
					},
					{
						GoodsId:  "dup",
						ExpireAt: 123,
					},
				},
			}, orderType)
			require.Error(t, err)

			// zero expiry
			err = dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						GoodsId:  "alice",
						ExpireAt: -1,
					},
					{
						GoodsId:  "bob",
						ExpireAt: 0,
					},
				},
			}, orderType)
			require.Error(t, err)
		}
	})

	t.Run("each order type persists separately", func(t *testing.T) {
		err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					GoodsId:  "goods",
					ExpireAt: 1,
				},
			},
		}, dymnstypes.NameOrder)
		require.NoError(t, err)

		err = dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					GoodsId:  "goods",
					ExpireAt: 2,
				},
			},
		}, dymnstypes.AliasOrder)
		require.NoError(t, err)

		listName := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.NameOrder)
		require.Len(t, listName.Records, 1)
		require.Equal(t, "goods", listName.Records[0].GoodsId)
		require.Equal(t, int64(1), listName.Records[0].ExpireAt)

		listAlias := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.AliasOrder)
		require.Len(t, listAlias.Records, 1)
		require.Equal(t, "goods", listAlias.Records[0].GoodsId)
		require.Equal(t, int64(2), listAlias.Records[0].ExpireAt)
	})
}

func TestKeeper_GetSetMinExpiryHistoricalSellOrder(t *testing.T) {
	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	for _, orderType := range supportedOrderTypes {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetMinExpiryHistoricalSellOrder(ctx, "hello", orderType, 123)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "world", orderType, 456)

		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, "hello", orderType)
		require.True(t, found)
		require.Equal(t, int64(123), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "world", orderType)
		require.True(t, found)
		require.Equal(t, int64(456), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "non-exists", orderType)
		require.False(t, found)
		require.Zero(t, minExpiry)

		t.Run("set zero means delete", func(t *testing.T) {
			dk.SetMinExpiryHistoricalSellOrder(ctx, "hello", orderType, 0)

			minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "hello", orderType)
			require.False(t, found)
			require.Zero(t, minExpiry)
		})
	}

	t.Run("each order type persists separately", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetMinExpiryHistoricalSellOrder(ctx, "goods", dymnstypes.NameOrder, 1)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "goods", dymnstypes.AliasOrder, 2)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "pool", dymnstypes.AliasOrder, 2)

		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, "goods", dymnstypes.NameOrder)
		require.True(t, found)
		require.Equal(t, int64(1), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "goods", dymnstypes.AliasOrder)
		require.True(t, found)
		require.Equal(t, int64(2), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "pool", dymnstypes.NameOrder)
		require.False(t, found)
		require.Zero(t, minExpiry)
	})
}

func TestKeeper_GetAllSellOrders(t *testing.T) {
	supportedOrderTypes := []dymnstypes.OrderType{
		dymnstypes.NameOrder, dymnstypes.AliasOrder,
	}

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	var sellOrders []dymnstypes.SellOrder

	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1000))
	require.NoError(t, err)

	seed := 200 + int(n.Int64())

	for i := 0; i < seed; i++ {
		for j, orderType := range supportedOrderTypes {
			so := dymnstypes.SellOrder{
				GoodsId:  fmt.Sprintf("dog%03d%d", i, j),
				Type:     orderType,
				ExpireAt: 1,
				MinPrice: dymnsutils.TestCoin(1),
			}
			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			sellOrders = append(sellOrders, so)
		}
	}

	require.Len(t, sellOrders, seed*len(supportedOrderTypes))

	allSellOrders := dk.GetAllSellOrders(ctx)

	require.Len(t, allSellOrders, len(sellOrders), "should returns all inserted records")

	for _, so := range sellOrders {
		require.Contains(t, allSellOrders, so)
	}
}
