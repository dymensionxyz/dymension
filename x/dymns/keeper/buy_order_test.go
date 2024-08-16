package keeper_test

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_IncreaseBuyOrdersCountAndGet() {
	s.Require().Zero(s.dymNsKeeper.GetCountBuyOrders(s.ctx))

	count := s.dymNsKeeper.IncreaseBuyOrdersCountAndGet(s.ctx)
	s.Require().Equal(uint64(1), count)
	s.Require().Equal(uint64(1), s.dymNsKeeper.GetCountBuyOrders(s.ctx))

	count = s.dymNsKeeper.IncreaseBuyOrdersCountAndGet(s.ctx)
	s.Require().Equal(uint64(2), count)
	s.Require().Equal(uint64(2), s.dymNsKeeper.GetCountBuyOrders(s.ctx))

	count = s.dymNsKeeper.IncreaseBuyOrdersCountAndGet(s.ctx)
	s.Require().Equal(uint64(3), count)
	s.Require().Equal(uint64(3), s.dymNsKeeper.GetCountBuyOrders(s.ctx))

	s.dymNsKeeper.SetCountBuyOrders(s.ctx, math.MaxUint64-1)

	count = s.dymNsKeeper.IncreaseBuyOrdersCountAndGet(s.ctx)
	s.Require().Equal(uint64(math.MaxUint64), count)
	s.Require().Equal(uint64(math.MaxUint64), s.dymNsKeeper.GetCountBuyOrders(s.ctx))

	s.Require().Panics(func() {
		s.dymNsKeeper.IncreaseBuyOrdersCountAndGet(s.ctx)
	}, "expect panic on overflow when increasing count of all-time buy offer records greater than uint64")
}

func (s *KeeperTestSuite) TestKeeper_GetSetInsertNewBuyOrder() {
	buyerA := testAddr(1).bech32()

	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("get non-exists offer should returns nil", func() {
		offer := s.dymNsKeeper.GetBuyOrder(s.ctx, "10183418")
		s.Require().Nil(offer)
	})

	s.Run("should returns error when set empty ID offer", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				var params []string
				if assetType == dymnstypes.TypeAlias {
					params = []string{"rollapp_1-1"}
				}

				err := s.dymNsKeeper.SetBuyOrder(s.ctx, dymnstypes.BuyOrder{
					Id:         "",
					AssetId:    "asset",
					AssetType:  assetType,
					Params:     params,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				})
				s.Require().ErrorContains(err, "ID of offer is empty")
			})
		}
	})

	s.Run("can set and can get", func() {
		s.RefreshContext()

		offer1 := dymnstypes.BuyOrder{
			Id:         "101",
			AssetId:    "my-name",
			AssetType:  dymnstypes.TypeName,
			Buyer:      buyerA,
			OfferPrice: s.coin(1),
		}

		err := s.dymNsKeeper.SetBuyOrder(s.ctx, offer1)
		s.Require().NoError(err)

		offerGot1 := s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id)
		s.Require().NotNil(offerGot1)

		s.Require().Equal(offer1, *offerGot1)

		offer2 := dymnstypes.BuyOrder{
			Id:         "202",
			AssetId:    "alias",
			AssetType:  dymnstypes.TypeAlias,
			Params:     []string{"rollapp_1-1"},
			Buyer:      buyerA,
			OfferPrice: s.coin(1),
		}

		err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer2)
		s.Require().NoError(err)

		offerGot2 := s.dymNsKeeper.GetBuyOrder(s.ctx, offer2.Id)
		s.Require().NotNil(offerGot2)

		s.Require().Equal(offer2, *offerGot2)

		// previous record should not be effected

		offerGot1 = s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id)
		s.Require().NotNil(offerGot1)

		s.Require().NotEqual(*offerGot1, *offerGot2)
		s.Require().Equal(offer1, *offerGot1)
	})

	s.Run("set omits params if empty", func() {
		s.RefreshContext()

		offer1 := dymnstypes.BuyOrder{
			Id:         "101",
			AssetId:    "my-name",
			AssetType:  dymnstypes.TypeName,
			Params:     []string{},
			Buyer:      buyerA,
			OfferPrice: s.coin(1),
		}

		err := s.dymNsKeeper.SetBuyOrder(s.ctx, offer1)
		s.Require().NoError(err)

		offerGot1 := s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id)
		s.Require().NotNil(offerGot1)

		s.Require().Nil(offerGot1.Params)
	})

	s.Run("should panic when insert non-empty ID offer", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				var params []string
				if assetType == dymnstypes.TypeAlias {
					params = []string{"rollapp_1-1"}
				}

				s.Require().Panics(func() {
					_, _ = s.dymNsKeeper.InsertNewBuyOrder(s.ctx, dymnstypes.BuyOrder{
						Id:         dymnstypes.CreateBuyOrderId(assetType, 1),
						AssetId:    "asset",
						AssetType:  assetType,
						Params:     params,
						Buyer:      buyerA,
						OfferPrice: s.coin(1),
					})
				})
			})
		}
	})

	s.Run("can insert", func() {
		s.RefreshContext()

		offer1a := dymnstypes.BuyOrder{
			Id:         "",
			AssetId:    "my-name",
			AssetType:  dymnstypes.TypeName,
			Buyer:      buyerA,
			OfferPrice: s.coin(1),
		}

		offer1b, err := s.dymNsKeeper.InsertNewBuyOrder(s.ctx, offer1a)
		s.Require().NoError(err)
		s.Require().Equal("101", offer1b.Id)

		offerGot1 := s.dymNsKeeper.GetBuyOrder(s.ctx, "101")
		s.Require().NotNil(offerGot1)

		offer1a.Id = offer1b.Id
		s.Require().Equal(offer1a, *offerGot1)

		offer2a := dymnstypes.BuyOrder{
			Id:         "",
			AssetId:    "alias",
			AssetType:  dymnstypes.TypeAlias,
			Params:     []string{"rollapp_1-1"},
			Buyer:      buyerA,
			OfferPrice: s.coin(1),
		}

		offer2b, err := s.dymNsKeeper.InsertNewBuyOrder(s.ctx, offer2a)
		s.Require().NoError(err)
		s.Require().Equal("202", offer2b.Id)

		offerGot2 := s.dymNsKeeper.GetBuyOrder(s.ctx, "202")
		s.Require().NotNil(offerGot2)

		offer2a.Id = offer2b.Id
		s.Require().Equal(offer2a, *offerGot2)

		// previous record should not be effected

		offerGot1 = s.dymNsKeeper.GetBuyOrder(s.ctx, "101")
		s.Require().NotNil(offerGot1)

		s.Require().NotEqual(*offerGot1, *offerGot2)
		s.Require().Equal(offer1a, *offerGot1)
	})

	s.Run("can not insert duplicated", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
				const nextId uint64 = 2

				var params []string
				if assetType == dymnstypes.TypeAlias {
					params = []string{"rollapp_1-1"}
				}

				existing := dymnstypes.BuyOrder{
					Id:         dymnstypes.CreateBuyOrderId(assetType, nextId),
					AssetId:    "asset",
					AssetType:  assetType,
					Params:     params,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				}

				err := s.dymNsKeeper.SetBuyOrder(s.ctx, existing)
				s.Require().NoError(err)

				offer := dymnstypes.BuyOrder{
					Id:         "",
					AssetId:    "asset",
					AssetType:  assetType,
					Params:     params,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				}

				_, err = s.dymNsKeeper.InsertNewBuyOrder(s.ctx, offer)
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "Buy-Order ID already exists")
			})
		}
	})

	s.Run("should automatically fill ID when insert", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				var params []string
				if assetType == dymnstypes.TypeAlias {
					params = []string{"rollapp_1-1"}
				}

				offer1 := dymnstypes.BuyOrder{
					Id:         "",
					AssetId:    "one",
					AssetType:  assetType,
					Params:     params,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				}

				offer, err := s.dymNsKeeper.InsertNewBuyOrder(s.ctx, offer1)
				s.Require().NoError(err)

				wantId1 := dymnstypes.CreateBuyOrderId(assetType, 1)
				s.Require().Equal(wantId1, offer.Id)

				offerGot := s.dymNsKeeper.GetBuyOrder(s.ctx, wantId1)
				s.Require().NotNil(offerGot)

				offer1.Id = wantId1
				s.Require().Equal(offer1, *offerGot)

				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 99)

				offer2 := dymnstypes.BuyOrder{
					Id:         "",
					AssetId:    "two",
					AssetType:  assetType,
					Params:     params,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				}
				offer, err = s.dymNsKeeper.InsertNewBuyOrder(s.ctx, offer2)
				s.Require().NoError(err)

				wantId2 := dymnstypes.CreateBuyOrderId(assetType, 100)
				s.Require().Equal(wantId2, offer.Id)

				offerGot = s.dymNsKeeper.GetBuyOrder(s.ctx, wantId2)
				s.Require().NotNil(offerGot)

				offer2.Id = wantId2
				s.Require().Equal(offer2, *offerGot)
			})
		}
	})

	s.Run("can delete", func() {
		s.RefreshContext()

		var err error

		offer1 := dymnstypes.BuyOrder{
			Id:         "101",
			AssetId:    "a",
			AssetType:  dymnstypes.TypeName,
			Buyer:      buyerA,
			OfferPrice: s.coin(1),
		}
		err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer1)
		s.Require().NoError(err)

		offer2 := dymnstypes.BuyOrder{
			Id:         "202",
			AssetId:    "b",
			AssetType:  dymnstypes.TypeAlias,
			Params:     []string{"rollapp_1-1"},
			Buyer:      buyerA,
			OfferPrice: s.coin(2),
		}
		err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer2)
		s.Require().NoError(err)

		offer3 := dymnstypes.BuyOrder{
			Id:         "103",
			AssetId:    "c",
			AssetType:  dymnstypes.TypeName,
			Buyer:      buyerA,
			OfferPrice: s.coin(3),
		}
		err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer3)
		s.Require().NoError(err)

		offer4 := dymnstypes.BuyOrder{
			Id:         "204",
			AssetId:    "d",
			AssetType:  dymnstypes.TypeAlias,
			Params:     []string{"rollapp_2-2"},
			Buyer:      buyerA,
			OfferPrice: s.coin(4),
		}
		err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer4)
		s.Require().NoError(err)

		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id))
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer2.Id))
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer3.Id))
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer4.Id))

		s.dymNsKeeper.DeleteBuyOrder(s.ctx, offer2.Id)
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id))
		s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer2.Id))
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer3.Id))
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer4.Id))

		s.dymNsKeeper.DeleteBuyOrder(s.ctx, offer4.Id)
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id))
		s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer2.Id))
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer3.Id))
		s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer4.Id))

		s.dymNsKeeper.DeleteBuyOrder(s.ctx, offer3.Id)
		s.Require().NotNil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer1.Id))
		s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer2.Id))
		s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer3.Id))
		s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, offer4.Id))
	})

	s.Run("delete non-existing will not panics", func() {
		s.RefreshContext()

		s.dymNsKeeper.DeleteBuyOrder(s.ctx, "1099999")
		s.dymNsKeeper.DeleteBuyOrder(s.ctx, "2099999")
	})

	s.Run("event should be fired on set/insert offer", func() {
		tests := []struct {
			name    string
			offer   dymnstypes.BuyOrder
			setFunc func(offer dymnstypes.BuyOrder, s *KeeperTestSuite)
		}{
			{
				name: "set offer type Dym-Name",
				offer: dymnstypes.BuyOrder{
					Id:         "101",
					AssetId:    "my-name",
					AssetType:  dymnstypes.TypeName,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				},
				setFunc: func(offer dymnstypes.BuyOrder, s *KeeperTestSuite) {
					err := s.dymNsKeeper.SetBuyOrder(s.ctx, offer)
					s.Require().NoError(err)
				},
			},
			{
				name: "set offer type Alias",
				offer: dymnstypes.BuyOrder{
					Id:         "201",
					AssetId:    "alias",
					AssetType:  dymnstypes.TypeAlias,
					Params:     []string{"rollapp_1-1"},
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				},
				setFunc: func(offer dymnstypes.BuyOrder, s *KeeperTestSuite) {
					err := s.dymNsKeeper.SetBuyOrder(s.ctx, offer)
					s.Require().NoError(err)
				},
			},
		}
		for _, tt := range tests {
			s.Run(tt.name, func() {
				s.RefreshContext()

				tt.setFunc(tt.offer, s)

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

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
					s.Require().NotEmpty(actionName, "event attr action name could not be found")
					s.Require().Equalf(
						actionName, dymnstypes.AttributeValueBoActionNameSet,
						"event attr action name should be `%s`", dymnstypes.AttributeValueBoActionNameSet,
					)
					return
				}

				s.T().Errorf("event %s not found", dymnstypes.EventTypeBuyOrder)
			})
		}
	})

	s.Run("event should be fired on delete offer", func() {
		for _, assetType := range supportedAssetTypes {
			s.Run(assetType.PrettyName(), func() {
				s.RefreshContext()

				var params []string
				if assetType == dymnstypes.TypeAlias {
					params = []string{"rollapp_1-1"}
				}

				offer := dymnstypes.BuyOrder{
					Id:         dymnstypes.CreateBuyOrderId(assetType, 1),
					AssetId:    "asset",
					AssetType:  assetType,
					Params:     params,
					Buyer:      buyerA,
					OfferPrice: s.coin(1),
				}

				err := s.dymNsKeeper.SetBuyOrder(s.ctx, offer)
				s.Require().NoError(err)

				s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())

				s.dymNsKeeper.DeleteBuyOrder(s.ctx, offer.Id)

				events := s.ctx.EventManager().Events()
				s.Require().NotEmpty(events)

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
					s.Require().NotEmpty(actionName, "event attr action name could not be found")
					s.Require().Equalf(
						actionName, dymnstypes.AttributeValueBoActionNameDelete,
						"event attr action name should be `%s`", dymnstypes.AttributeValueBoActionNameDelete,
					)
					return
				}

				s.T().Errorf("event %s not found", dymnstypes.EventTypeBuyOrder)
			})
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAllBuyOrders() {
	buyerA := testAddr(1).bech32()

	offer1 := dymnstypes.BuyOrder{
		Id:         "101",
		AssetId:    "a",
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: s.coin(1),
	}
	err := s.dymNsKeeper.SetBuyOrder(s.ctx, offer1)
	s.Require().NoError(err)

	offers := s.dymNsKeeper.GetAllBuyOrders(s.ctx)
	s.Require().Len(offers, 1)
	s.Require().Equal(offer1, offers[0])

	offer2 := dymnstypes.BuyOrder{
		Id:         "202",
		AssetId:    "b",
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{"rollapp_1-1"},
		Buyer:      buyerA,
		OfferPrice: s.coin(1),
	}
	err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer2)
	s.Require().NoError(err)

	offers = s.dymNsKeeper.GetAllBuyOrders(s.ctx)
	s.Require().Len(offers, 2)
	s.Require().Equal([]dymnstypes.BuyOrder{offer1, offer2}, offers)

	offer3 := dymnstypes.BuyOrder{
		Id:         "103",
		AssetId:    "a",
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: s.coin(1),
	}
	err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer3)
	s.Require().NoError(err)

	offers = s.dymNsKeeper.GetAllBuyOrders(s.ctx)
	s.Require().Len(offers, 3)
	s.Require().Equal([]dymnstypes.BuyOrder{
		offer1, offer3, // <= Dym-Name Buy-Order should be sorted first
		offer2, // <= Alias Buy-Order should be sorted second
		// because of store branched by asset type
	}, offers)

	offer4 := dymnstypes.BuyOrder{
		Id:         "204",
		AssetId:    "b",
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{"rollapp_2-2"},
		Buyer:      buyerA,
		OfferPrice: s.coin(1),
	}
	err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer4)
	s.Require().NoError(err)

	offers = s.dymNsKeeper.GetAllBuyOrders(s.ctx)
	s.Require().Len(offers, 4)
	s.Require().Equal([]dymnstypes.BuyOrder{offer1, offer3, offer2, offer4}, offers)

	offer5 := dymnstypes.BuyOrder{
		Id:         "105",
		AssetId:    "b",
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: s.coin(3),
	}
	err = s.dymNsKeeper.SetBuyOrder(s.ctx, offer5)
	s.Require().NoError(err)

	offers = s.dymNsKeeper.GetAllBuyOrders(s.ctx)
	s.Require().Len(offers, 5)
	s.Require().Equal([]dymnstypes.BuyOrder{offer1, offer3, offer5, offer2, offer4}, offers)
}
