package keeper_test

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_GetSetSellOrder() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("reject invalid SO", func() {
		err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{})
		s.Require().Error(err)
	})

	s.Run("can set", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(100),
				}
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)
			})
		}
	})

	s.Run("event should be fired on set", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(100),
				}
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

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
						s.Require().NotEmpty(actionName, "event attr action name could not be found")
						s.Require().Equalf(
							actionName, dymnstypes.AttributeValueSoActionNameSet,
							"event attr action name should be `%s`", dymnstypes.AttributeValueSoActionNameSet,
						)
						return
					}

					s.T().Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
				}(events)
			})
		}
	})

	s.Run("can not set Sell-Order with unknown type", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.AssetType_AT_UNKNOWN,
			ExpireAt:  1,
			MinPrice:  s.coin(100),
		})
		s.Require().ErrorContains(err, "invalid SO type")
	})

	s.Run("non-exists returns nil", func() {
		s.RefreshContext()

		for _, assetType := range supportedAssetTypes {
			s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, "asset", assetType))
		}
	})

	s.Run("omit Sell Price if zero", func() {
		for _, sellPrice := range []*sdk.Coin{nil, uptr.To(s.coin(0)), {}} {
			for _, assetType := range supportedAssetTypes {
				s.Run(assetType.PrettyName(), func() {
					s.RefreshContext()

					err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{
						AssetId:   "asset",
						AssetType: assetType,
						ExpireAt:  1,
						MinPrice:  s.coin(100),
						SellPrice: sellPrice,
					})
					s.Require().NoError(err)

					s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, "asset", assetType).SellPrice)
				})
			}
		}
	})

	s.Run("omit Bid params if empty", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetSellOrder(s.ctx, dymnstypes.SellOrder{
			AssetId:   "asset",
			AssetType: dymnstypes.TypeName,
			ExpireAt:  1,
			MinPrice:  s.coin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: testAddr(1).bech32(),
				Price:  s.coin(200),
				Params: []string{},
			},
		})
		s.Require().NoError(err)

		s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, "asset", dymnstypes.TypeName).HighestBid.Params)
	})

	s.Run("get returns correct inserted record, regardless type", func() {
		s.RefreshContext()

		var sellOrders []dymnstypes.SellOrder

		const seed int = 100

		for i := 0; i < seed; i++ {
			for _, assetType := range supportedAssetTypes {

				so := dymnstypes.SellOrder{
					AssetId:   fmt.Sprintf("dog%d", i), // same asset id for all types
					AssetType: assetType,
					ExpireAt:  1 + int64(i),
					MinPrice:  s.coin(int64(seed + i)),
				}
				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				sellOrders = append(sellOrders, so)
			}
		}

		for _, so := range sellOrders {
			got := s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType)
			s.Require().NotNil(got)
			s.Require().Equal(so, *got)
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_DeleteSellOrder() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("can delete", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(1),
				}

				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType))

				s.dymNsKeeper.DeleteSellOrder(s.ctx, so.AssetId, so.AssetType)

				s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType))
			})
		}
	})

	s.Run("event should be fired upon deletion", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				so := dymnstypes.SellOrder{
					AssetId:   "asset",
					AssetType: assetType,
					ExpireAt:  1,
					MinPrice:  s.coin(1),
				}

				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				s.dymNsKeeper.DeleteSellOrder(s.ctx, so.AssetId, so.AssetType)

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

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
						s.Require().NotEmpty(actionName, "event attr action name could not be found")
						s.Require().Equalf(
							actionName, dymnstypes.AttributeValueSoActionNameSet,
							"event attr action name should be `%s`", dymnstypes.AttributeValueSoActionNameDelete,
						)
						return
					}

					s.T().Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
				}(events)
			})
		}
	})

	s.Run("delete remove the correct record regardless type", func() {
		s.RefreshContext()

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
					MinPrice:  s.coin(1),
				}

				err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
				s.Require().NoError(err)

				s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, so.AssetId, so.AssetType))

				testCases = append(testCases, &testCase{so: so, deleted: false})
			}
		}

		s.Require().Len(testCases, seed*len(supportedAssetTypes))
		s.Require().Len(s.dymNsKeeper.GetAllSellOrders(s.ctx), len(testCases))

		// test delete
		for i, tc := range testCases {
			s.dymNsKeeper.DeleteSellOrder(s.ctx, tc.so.AssetId, tc.so.AssetType)
			tc.deleted = true
			s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, tc.so.AssetId, tc.so.AssetType))

			s.Require().Len(s.dymNsKeeper.GetAllSellOrders(s.ctx), len(testCases)-(i+1))

			for _, tc2 := range testCases {
				if tc2.deleted {
					s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, tc2.so.AssetId, tc2.so.AssetType))
				} else {
					s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, tc2.so.AssetId, tc2.so.AssetType))
				}
			}
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_GetSetActiveSellOrdersExpiration() {
	s.RefreshContext()

	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("get", func() {
		for _, assetType := range supportedAssetTypes {
			aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, assetType)
			s.Require().Empty(aSoe.Records, "default list must be empty")
			s.Require().NotNil(aSoe.Records, "list must be initialized")
		}
	})

	s.Run("set", func() {
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
			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, aSoe, assetType)
			s.Require().NoError(err)

			aSoe = s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, assetType)
			s.Require().Len(aSoe.Records, 2)
			s.Require().Equal("hello", aSoe.Records[0].AssetId)
			s.Require().Equal(int64(123), aSoe.Records[0].ExpireAt)
			s.Require().Equal("world", aSoe.Records[1].AssetId)
			s.Require().Equal(int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	s.Run("must automatically sort when set", func() {
		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
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
			s.Require().NoError(err)

			aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, assetType)
			s.Require().Len(aSoe.Records, 2)

			s.Require().Equal("aaa", aSoe.Records[0].AssetId)
			s.Require().Equal(int64(123), aSoe.Records[0].ExpireAt)
			s.Require().Equal("bbb", aSoe.Records[1].AssetId)
			s.Require().Equal(int64(456), aSoe.Records[1].ExpireAt)
		}
	})

	s.Run("can not set if set is not valid", func() {
		for _, assetType := range supportedAssetTypes {
			// not unique
			err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
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
			s.Require().Error(err)

			// zero expiry
			err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
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
			s.Require().Error(err)
		}
	})

	s.Run("each asset type persists separately", func() {
		err := s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "asset",
					ExpireAt: 1,
				},
			},
		}, dymnstypes.TypeName)
		s.Require().NoError(err)

		err = s.dymNsKeeper.SetActiveSellOrdersExpiration(s.ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					AssetId:  "asset",
					ExpireAt: 2,
				},
			},
		}, dymnstypes.TypeAlias)
		s.Require().NoError(err)

		listName := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeName)
		s.Require().Len(listName.Records, 1)
		s.Require().Equal("asset", listName.Records[0].AssetId)
		s.Require().Equal(int64(1), listName.Records[0].ExpireAt)

		listAlias := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)
		s.Require().Len(listAlias.Records, 1)
		s.Require().Equal("asset", listAlias.Records[0].AssetId)
		s.Require().Equal(int64(2), listAlias.Records[0].ExpireAt)
	})
}

func Benchmark_SetActiveSellOrdersExpiration(b *testing.B) {
	b.StopTimer()
	b.ReportAllocs()

	// 2024-09-26: 0.356s for setting a list of 777600 elements
	// Benchmark_SetActiveSellOrdersExpiration-8 | 2s563ms | 3 | 356593625 ns/op | 85929664 B/op | 29837 allocs/op

	for r := 1; r <= b.N; r++ {
		now := time.Now().Unix()
		dk, _, _, ctx := testkeeper.DymNSKeeper(b)
		aSoe := dk.GetActiveSellOrdersExpiration(ctx, dymnstypes.TypeName)

		const sampleCount = 3 * 86400 /*max live SO*/ / 5 /*block time*/ * (400_000_000/int(dymnstypes.OpGasPlaceSellOrder) - 1) /*max num of txs per block*/
		fmt.Println("Elements count", sampleCount)

		aSoe.Records = make([]dymnstypes.ActiveSellOrdersExpirationRecord, sampleCount)

		for j := 0; j < sampleCount; j++ {
			dymName := fmt.Sprintf("name-%d", j+rand.Int())
			aSoe.Records[j] = dymnstypes.ActiveSellOrdersExpirationRecord{
				AssetId:  dymName,
				ExpireAt: now + int64(j)*2,
			}
		}

		err := func() error {
			b.StartTimer()
			defer b.StopTimer()
			return dk.SetActiveSellOrdersExpiration(ctx, aSoe, dymnstypes.TypeName)
		}()
		require.NoError(b, err)
	}
}

func (s *KeeperTestSuite) TestKeeper_GetAllSellOrders() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.RefreshContext()

	var sellOrders []dymnstypes.SellOrder

	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(1000))
	s.Require().NoError(err)

	seed := 200 + int(n.Int64())

	for i := 0; i < seed; i++ {
		for j, assetType := range supportedAssetTypes {
			so := dymnstypes.SellOrder{
				AssetId:   fmt.Sprintf("dog%03d%d", i, j),
				AssetType: assetType,
				ExpireAt:  1,
				MinPrice:  s.coin(1),
			}
			err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
			s.Require().NoError(err)

			sellOrders = append(sellOrders, so)
		}
	}

	s.Require().Len(sellOrders, seed*len(supportedAssetTypes))

	allSellOrders := s.dymNsKeeper.GetAllSellOrders(s.ctx)

	s.Require().Len(allSellOrders, len(sellOrders), "should returns all inserted records")

	for _, so := range sellOrders {
		s.Require().Contains(allSellOrders, so)
	}
}
