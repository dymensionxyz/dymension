package keeper_test

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetSetSellOrder(t *testing.T) {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	t.Run("reject invalid SO", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{})
		require.Error(t, err)
	})

	t.Run("can set", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				AssetId:   "asset",
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  dymnsutils.TestCoin(100),
			}
			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)
		}
	})

	t.Run("event should be fired on set", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				AssetId:   "asset",
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  dymnsutils.TestCoin(100),
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
			AssetId:   "asset",
			AssetType: dymnstypes.AssetType_AT_UNKNOWN,
			ExpireAt:  1,
			MinPrice:  dymnsutils.TestCoin(100),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid SO type")
	})

	t.Run("non-exists returns nil", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			require.Nil(t, dk.GetSellOrder(ctx, "asset", assetType))
		}
	})

	t.Run("omit Sell Price if zero", func(t *testing.T) {
		for _, sellPrice := range []*sdk.Coin{nil, dymnsutils.TestCoinP(0), {}} {
			for _, assetType := range supportedAssetTypes {
				dk, _, _, ctx := testkeeper.DymNSKeeper(t)

				err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(100),
					SellPrice: sellPrice,
				})
				require.NoError(t, err)

				require.Nil(t, dk.GetSellOrder(ctx, "asset", assetType).SellPrice)
			}
		}
	})

	t.Run("omit Bid params if empty", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.TypeName,
			ExpireAt:  1,
			MinPrice:  dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: testAddr(1).bech32(),
				Price:  dymnsutils.TestCoin(200),
				Params: []string{},
			},
		})
		require.NoError(t, err)

		require.Nil(t, dk.GetSellOrder(ctx, "asset", dymnstypes.TypeName).HighestBid.Params)
	})

	t.Run("get returns correct inserted record, regardless type", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		var sellOrders []dymnstypes.SellOrder

		const seed int = 100

		for i := 0; i < seed; i++ {
			for _, assetType := range supportedAssetTypes {

				so := dymnstypes.SellOrder{
					AssetId:   fmt.Sprintf("dog%d", i), // same asset id for all types
					AssetType: assetType,
					ExpireAt:  1 + int64(i),
					MinPrice:  dymnsutils.TestCoin(int64(seed + i)),
				}
				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)

				sellOrders = append(sellOrders, so)
			}
		}

		for _, so := range sellOrders {
			got := dk.GetSellOrder(ctx, so.AssetId, so.AssetType)
			require.NotNil(t, got)
			require.Equal(t, so, *got)
		}
	})
}

func TestKeeper_DeleteSellOrder(t *testing.T) {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	t.Run("can delete", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				AssetId:   "asset",
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  dymnsutils.TestCoin(1),
			}

			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			require.NotNil(t, dk.GetSellOrder(ctx, so.AssetId, so.AssetType))

			dk.DeleteSellOrder(ctx, so.AssetId, so.AssetType)

			require.Nil(t, dk.GetSellOrder(ctx, so.AssetId, so.AssetType))
		}
	})

	t.Run("event should be fired upon deletion", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			dk, _, _, ctx := testkeeper.DymNSKeeper(t)

			so := dymnstypes.SellOrder{
				AssetId:   "asset",
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  dymnsutils.TestCoin(1),
			}

			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			dk.DeleteSellOrder(ctx, so.AssetId, so.AssetType)

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
			for j, assetType := range supportedAssetTypes {
				so := dymnstypes.SellOrder{
					AssetId:   fmt.Sprintf("dog%03d%d", i, j),
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  dymnsutils.TestCoin(1),
				}

				err := dk.SetSellOrder(ctx, so)
				require.NoError(t, err)

				require.NotNil(t, dk.GetSellOrder(ctx, so.AssetId, so.AssetType))

				testCases = append(testCases, &testCase{so: so, deleted: false})
			}
		}

		require.Len(t, testCases, seed*len(supportedAssetTypes))
		require.Len(t, dk.GetAllSellOrders(ctx), len(testCases))

		// test delete
		for i, tc := range testCases {
			dk.DeleteSellOrder(ctx, tc.so.AssetId, tc.so.AssetType)
			tc.deleted = true
			require.Nil(t, dk.GetSellOrder(ctx, tc.so.AssetId, tc.so.AssetType))

			require.Len(t, dk.GetAllSellOrders(ctx), len(testCases)-(i+1))

			for _, tc2 := range testCases {
				if tc2.deleted {
					require.Nil(t, dk.GetSellOrder(ctx, tc2.so.AssetId, tc2.so.AssetType))
				} else {
					require.NotNil(t, dk.GetSellOrder(ctx, tc2.so.AssetId, tc2.so.AssetType))
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
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so11)
	require.NoError(t, err)

	alias2 := "alias"
	so21 := dymnstypes.SellOrder{
		AssetId:   alias2,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  2,
		MinPrice:  dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so21)
	require.NoError(t, err)

	alias3 := "salas"
	so3 := dymnstypes.SellOrder{
		AssetId:   alias3,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  3,
		MinPrice:  dymnsutils.TestCoin(222),
	}
	err = dk.SetSellOrder(ctx, so3)
	require.NoError(t, err)

	t.Run("should able to move", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so11.AssetId, so11.AssetType)
		require.NoError(t, err)

		err = dk.MoveSellOrderToHistorical(ctx, so21.AssetId, so21.AssetType)
		require.NoError(t, err)
	})

	t.Run("moved SO should be removed from active", func(t *testing.T) {
		require.Nil(t, dk.GetSellOrder(ctx, so11.AssetId, so11.AssetType))
		require.Nil(t, dk.GetSellOrder(ctx, so21.AssetId, so21.AssetType))
	})

	t.Run("has min expiry mapping", func(t *testing.T) {
		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, so11.AssetId, so11.AssetType)
		require.True(t, found)
		require.Equal(t, so11.ExpireAt, minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, so21.AssetId, so21.AssetType)
		require.True(t, found)
		require.Equal(t, so21.ExpireAt, minExpiry)
	})

	t.Run("should not move non-exists", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, "non-exists", dymnstypes.TypeName)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order: Dym-Name: non-exists")

		err = dk.MoveSellOrderToHistorical(ctx, "non-exists", dymnstypes.TypeAlias)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order: Alias: non-exists")
	})

	t.Run("should able to move a duplicated without error", func(t *testing.T) {
		defer func() {
			dk.DeleteSellOrder(ctx, so11.AssetId, so11.AssetType)
			dk.DeleteSellOrder(ctx, so21.AssetId, so21.AssetType)
		}()

		for _, so := range []dymnstypes.SellOrder{so11, so21} {
			err = dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			err = dk.MoveSellOrderToHistorical(ctx, so.AssetId, so.AssetType)
			require.NoError(t, err)

			list := dk.GetHistoricalSellOrders(ctx, so.AssetId, so.AssetType)
			require.Len(t, list, 1, "do not persist duplicated historical SO")
		}
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Empty(t, dk.GetSellOrder(ctx, dymName2.Name, dymnstypes.TypeName))
		require.NotNil(t, dk.GetSellOrder(ctx, alias3, dymnstypes.TypeAlias))
	})

	so4 := dymnstypes.SellOrder{
		AssetId:   dymName2.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so4)
	require.NoError(t, err)

	t.Run("should able to move", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so4.AssetId, so4.AssetType)
		require.NoError(t, err)
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName), 1)
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName), 1)
		require.Len(t, dk.GetHistoricalSellOrders(ctx, alias2, dymnstypes.TypeAlias), 1)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias))
	})

	so12 := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  now.Unix() + 1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so12)
	require.NoError(t, err)

	so22 := dymnstypes.SellOrder{
		AssetId:   alias2,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  now.Unix() + 1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so22)
	require.NoError(t, err)

	t.Run("should not move yet finished SO", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so12.AssetId, so12.AssetType)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order not yet expired")

		err = dk.MoveSellOrderToHistorical(ctx, so22.AssetId, so22.AssetType)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Sell-Order not yet expired")
	})

	for _, so := range []dymnstypes.SellOrder{so12, so22} {
		so.HighestBid = &dymnstypes.SellOrderBid{
			Bidder: bidderA,
			Price:  dymnsutils.TestCoin(300),
		}
		if so.AssetType == dymnstypes.TypeAlias {
			so.HighestBid.Params = []string{"rollapp_1-1"}
		}

		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		t.Run("should able to move finished SO", func(t *testing.T) {
			err := dk.MoveSellOrderToHistorical(ctx, so.AssetId, so.AssetType)
			require.NoError(t, err)

			list := dk.GetHistoricalSellOrders(ctx, so.AssetId, so.AssetType)
			require.Len(t, list, 2, "should appended to historical")
		})
	}

	minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, so12.AssetId, so12.AssetType)
	require.True(t, found)
	require.Equal(t, so11.ExpireAt, minExpiry, "should keep the minimum")
	require.NotEqual(t, so12.ExpireAt, minExpiry, "should keep the minimum")

	minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, so22.AssetId, so22.AssetType)
	require.True(t, found)
	require.Equal(t, so21.ExpireAt, minExpiry, "should keep the minimum")
	require.NotEqual(t, so22.ExpireAt, minExpiry, "should keep the minimum")

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName), 1)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias))
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
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias4, dymnstypes.TypeAlias))
	})

	so11 := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so11)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so11.AssetId, so11.AssetType)
	require.NoError(t, err)

	so2 := dymnstypes.SellOrder{
		AssetId:   alias3,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  7,
		MinPrice:  dymnsutils.TestCoin(200),
	}
	err = dk.SetSellOrder(ctx, so2)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so2.AssetId, so2.AssetType)
	require.NoError(t, err)

	so3 := dymnstypes.SellOrder{
		AssetId:   dymName2.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so3)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so3.AssetId, so3.AssetType)
	require.NoError(t, err)

	so3.ExpireAt++
	err = dk.SetSellOrder(ctx, so3)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so3.AssetId, so3.AssetType)
	require.NoError(t, err)

	so4 := dymnstypes.SellOrder{
		AssetId:   alias4,
		AssetType: dymnstypes.TypeAlias,
		ExpireAt:  5,
		MinPrice:  dymnsutils.TestCoin(500),
	}
	err = dk.SetSellOrder(ctx, so4)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so4.AssetId, so4.AssetType)
	require.NoError(t, err)

	t.Run("fetch correctly", func(t *testing.T) {
		list1 := dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName)
		require.Len(t, list1, 1)
		list2 := dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName)
		require.Len(t, list2, 2)
		list3 := dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias)
		require.Len(t, list3, 1)
		list4 := dk.GetHistoricalSellOrders(ctx, alias4, dymnstypes.TypeAlias)
		require.Len(t, list4, 1)

		require.Equal(t, so3.AssetId, list2[0].AssetId)
		require.Equal(t, so3.AssetId, list2[1].AssetId)

		require.Equal(t, int64(1), list2[0].ExpireAt)
		require.Equal(t, int64(2), list2[1].ExpireAt)

		require.Equal(t, int64(7), list3[0].ExpireAt)
		require.Equal(t, int64(5), list4[0].ExpireAt)
	})

	t.Run("delete", func(t *testing.T) {
		dk.DeleteHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName))

		list2 := dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName)
		require.Len(t, list2, 2)

		dk.DeleteHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name, dymnstypes.TypeName))

		list3 := dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias)
		require.Len(t, list3, 1)

		dk.DeleteHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, alias3, dymnstypes.TypeAlias))

		list4 := dk.GetHistoricalSellOrders(ctx, alias4, dymnstypes.TypeAlias)
		require.Len(t, list4, 1)
	})
}

func TestKeeper_GetSetActiveSellOrdersExpiration(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	t.Run("get", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			aSoe := dk.GetActiveSellOrdersExpiration(ctx, assetType)
			require.Empty(t, aSoe.Records, "default list must be empty")
			require.NotNil(t, aSoe.Records, "list must be initialized")
		}
	})

	t.Run("set", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			aSoe := &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "hello",
						ExpireAt: 123,
					},
					{
						AssetId:  "world",
						ExpireAt: 456,
					},
				},
			}
			err := dk.SetActiveSellOrdersExpiration(ctx, aSoe, assetType)
			require.NoError(t, err)

			aSoe = dk.GetActiveSellOrdersExpiration(ctx, assetType)
			require.Len(t, aSoe.Records, 2)
			require.Equal(t, "hello", aSoe.Records[0].AssetId)
			require.Equal(t, int64(123), aSoe.Records[0].ExpireAt)
			require.Equal(t, "world", aSoe.Records[1].AssetId)
			require.Equal(t, int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	t.Run("must automatically sort when set", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "bbb",
						ExpireAt: 456,
					},
					{
						AssetId:  "aaa",
						ExpireAt: 123,
					},
				},
			}, assetType)
			require.NoError(t, err)

			aSoe := dk.GetActiveSellOrdersExpiration(ctx, assetType)
			require.Len(t, aSoe.Records, 2)

			require.Equal(t, "aaa", aSoe.Records[0].AssetId)
			require.Equal(t, int64(123), aSoe.Records[0].ExpireAt)
			require.Equal(t, "bbb", aSoe.Records[1].AssetId)
			require.Equal(t, int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	t.Run("can not set if set is not valid", func(t *testing.T) {
		for _, assetType := range supportedAssetTypes {
			// not unique
			err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "dup",
						ExpireAt: 456,
					},
					{
						AssetId:  "dup",
						ExpireAt: 123,
					},
				},
			}, assetType)
			require.Error(t, err)

			// zero expiry
			err = dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
				Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
					{
						AssetId:  "alice",
						ExpireAt: -1,
					},
					{
						AssetId:  "bob",
						ExpireAt: 0,
					},
				},
			}, assetType)
			require.Error(t, err)
		}
	})

	t.Run("each asset type persists separately", func(t *testing.T) {
		err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "asset",
					ExpireAt: 1,
				},
			},
		}, dymnstypes.TypeName)
		require.NoError(t, err)

		err = dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "asset",
					ExpireAt: 2,
				},
			},
		}, dymnstypes.TypeAlias)
		require.NoError(t, err)

		listName := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeName)
		require.Len(t, listName.Records, 1)
		require.Equal(t, "asset", listName.Records[0].AssetId)
		require.Equal(t, int64(1), listName.Records[0].ExpireAt)

		listAlias := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeAlias)
		require.Len(t, listAlias.Records, 1)
		require.Equal(t, "asset", listAlias.Records[0].AssetId)
		require.Equal(t, int64(2), listAlias.Records[0].ExpireAt)
	})
}

func TestKeeper_GetSetMinExpiryHistoricalSellOrder(t *testing.T) {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	for _, assetType := range supportedAssetTypes {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetMinExpiryHistoricalSellOrder(ctx, "hello", assetType, 123)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "world", assetType, 456)

		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, "hello", assetType)
		require.True(t, found)
		require.Equal(t, int64(123), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "world", assetType)
		require.True(t, found)
		require.Equal(t, int64(456), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "non-exists", assetType)
		require.False(t, found)
		require.Zero(t, minExpiry)

		t.Run("set zero means delete", func(t *testing.T) {
			dk.SetMinExpiryHistoricalSellOrder(ctx, "hello", assetType, 0)

			minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "hello", assetType)
			require.False(t, found)
			require.Zero(t, minExpiry)
		})
	}

	t.Run("each asset type persists separately", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetMinExpiryHistoricalSellOrder(ctx, "asset", dymnstypes.TypeName, 1)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "asset", dymnstypes.TypeAlias, 2)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "pool", dymnstypes.TypeAlias, 2)

		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, "asset", dymnstypes.TypeName)
		require.True(t, found)
		require.Equal(t, int64(1), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "asset", dymnstypes.TypeAlias)
		require.True(t, found)
		require.Equal(t, int64(2), minExpiry)

		minExpiry, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "pool", dymnstypes.TypeName)
		require.False(t, found)
		require.Zero(t, minExpiry)
	})
}

func TestKeeper_GetAllSellOrders(t *testing.T) {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	var sellOrders []dymnstypes.SellOrder

	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1000))
	require.NoError(t, err)

	seed := 200 + int(n.Int64())

	for i := 0; i < seed; i++ {
		for j, assetType := range supportedAssetTypes {
			so := dymnstypes.SellOrder{
				AssetId:   fmt.Sprintf("dog%03d%d", i, j),
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  dymnsutils.TestCoin(1),
			}
			err := dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			sellOrders = append(sellOrders, so)
		}
	}

	require.Len(t, sellOrders, seed*len(supportedAssetTypes))

	allSellOrders := dk.GetAllSellOrders(ctx)

	require.Len(t, allSellOrders, len(sellOrders), "should returns all inserted records")

	for _, so := range sellOrders {
		require.Contains(t, allSellOrders, so)
	}
}
