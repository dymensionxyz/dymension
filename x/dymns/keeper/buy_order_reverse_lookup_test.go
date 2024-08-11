package keeper_test

import (
	"time"

	"golang.org/x/exp/slices"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

func (s *KeeperTestSuite) TestKeeper_GetAddReverseMappingBuyerToPlacedBuyOrder() {
	buyer1a := testAddr(1).bech32()
	buyer2a := testAddr(2).bech32()
	buyer3a := testAddr(3).bech32()
	someoneA := testAddr(4).bech32()

	s.Require().Error(
		s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, "0x", "101"),
		"should not allow invalid buyer address",
	)

	s.Require().Error(
		s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer1a, "@"),
		"should not allow invalid offer ID",
	)

	_, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, "0x")
	s.Require().Error(err, "should not allow invalid buyer address")

	offer1 := dymnstypes.BuyOrder{
		Id:                     "101",
		AssetId:                "a",
		AssetType:              dymnstypes.TypeName,
		Buyer:                  buyer1a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer1))
	err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer1a, offer1.Id)
	s.Require().NoError(err)

	offer2 := dymnstypes.BuyOrder{
		Id:                     "202",
		AssetId:                "alias",
		AssetType:              dymnstypes.TypeAlias,
		Params:                 []string{"rollapp_1-1"},
		Buyer:                  buyer2a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer2))
	err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer2a, offer2.Id)
	s.Require().NoError(err)

	offer3 := dymnstypes.BuyOrder{
		Id:                     "103",
		AssetId:                "c",
		AssetType:              dymnstypes.TypeName,
		Buyer:                  buyer2a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer3))
	err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer2a, offer3.Id)
	s.Require().NoError(err)

	offer4 := dymnstypes.BuyOrder{
		Id:                     "204",
		AssetId:                "salas",
		AssetType:              dymnstypes.TypeAlias,
		Params:                 []string{"rollapp_2-2"},
		Buyer:                  buyer3a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer4))
	err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer3a, offer4.Id)
	s.Require().NoError(err)

	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer2a, "103721461"),
		"no check non-existing offer record",
	)

	s.Run("no error if duplicated ID", func() {
		for i := 0; i < 3; i++ {
			s.Require().NoError(s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer2a, offer2.Id))
		}
	})

	placedBy1, err1 := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyer1a)
	s.Require().NoError(err1)
	s.Require().Len(placedBy1, 1)

	placedBy2, err2 := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyer2a)
	s.Require().NoError(err2)
	s.Require().NotEqual(3, len(placedBy2), "should not include non-existing offers")
	s.Require().Len(placedBy2, 2)

	placedBy3, err3 := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyer3a)
	s.Require().NoError(err3)
	s.Require().Len(placedBy3, 1)

	placedByNonExists, err3 := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, someoneA)
	s.Require().NoError(err3)
	s.Require().Len(placedByNonExists, 0)

	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer2a, offer1.Id),
		"no error if offer placed by another buyer",
	)
	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyer2a, offer4.Id),
		"no error if offer placed by another buyer",
	)
	placedBy2, err2 = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyer2a)
	s.Require().NoError(err2)
	s.Require().Len(placedBy2, 2, "should not include offers placed by another buyer")
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingBuyerToPlacedBuyOrder() {
	buyerA := testAddr(1).bech32()
	someoneA := testAddr(2).bech32()

	s.Require().Error(
		s.dymNsKeeper.RemoveReverseMappingBuyerToBuyOrder(s.ctx, "0x", "101"),
		"should not allow invalid buyer address",
	)

	s.Require().Error(
		s.dymNsKeeper.RemoveReverseMappingBuyerToBuyOrder(s.ctx, buyerA, "@"),
		"should not allow invalid offer ID",
	)

	offer1 := dymnstypes.BuyOrder{
		Id:                     "101",
		AssetId:                "my-name",
		AssetType:              dymnstypes.TypeName,
		Buyer:                  buyerA,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer1))
	s.Require().NoError(s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyerA, offer1.Id))

	offer2 := dymnstypes.BuyOrder{
		Id:                     "202",
		AssetId:                "alias",
		AssetType:              dymnstypes.TypeAlias,
		Params:                 []string{"rollapp_1-1"},
		Buyer:                  buyerA,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer2))
	s.Require().NoError(s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyerA, offer2.Id))

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingBuyerToBuyOrder(s.ctx, someoneA, offer1.Id),
		"no error if buyer non-exists",
	)

	placedByBuyer, err := s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyerA)
	s.Require().NoError(err)
	s.Require().Len(placedByBuyer, 2, "existing data must be kept")

	s.Require().NoError(
		s.dymNsKeeper.RemoveReverseMappingBuyerToBuyOrder(s.ctx, buyerA, "10138132187"),
		"no error if not placed order",
	)

	placedByBuyer, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyerA)
	s.Require().NoError(err)
	s.Require().Len(placedByBuyer, 2, "existing data must be kept")

	s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingBuyerToBuyOrder(s.ctx, buyerA, offer1.Id))
	placedByBuyer, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyerA)
	s.Require().NoError(err)
	s.Require().Len(placedByBuyer, 1)
	s.Require().Equal(offer2.Id, placedByBuyer[0].Id)

	s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingBuyerToBuyOrder(s.ctx, buyerA, offer2.Id))
	placedByBuyer, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyerA)
	s.Require().NoError(err)
	s.Require().Len(placedByBuyer, 0)
}

func (s *KeeperTestSuite) TestKeeper_AddReverseMappingAssetIdToBuyOrder_Generic() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("pass - can add", func() {
		s.SetupTest()

		const assetId = "asset"

		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, assetId, assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			s.Require().NoError(err)
		}

		s.Require().NotEmpty(s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId)).OrderIds)
		s.Require().NotEmpty(s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, dymnstypes.AliasToBuyOrderIdsRvlKey(assetId)).OrderIds)
	})

	s.Run("pass - can add without collision across asset types", func() {
		s.SetupTest()

		const assetId = "asset"

		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, assetId, assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			s.Require().NoError(err)
		}

		err := s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, assetId, dymnstypes.TypeName, dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 2))
		s.Require().NoError(err)

		err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, assetId, dymnstypes.TypeAlias, dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 3))
		s.Require().NoError(err)

		nameOffers := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId))
		s.Require().Len(nameOffers.OrderIds, 2)
		aliasOffers := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, dymnstypes.AliasToBuyOrderIdsRvlKey(assetId))
		s.Require().Len(aliasOffers.OrderIds, 2)

		s.Require().NotEqual(nameOffers.OrderIds, aliasOffers.OrderIds, "data must be persisted separately")

		s.Require().Equal(dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 1), nameOffers.OrderIds[0])
		s.Require().Equal(dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 2), nameOffers.OrderIds[1])
		s.Require().Equal(dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 1), aliasOffers.OrderIds[0])
		s.Require().Equal(dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 3), aliasOffers.OrderIds[1])
	})

	s.Run("fail - should reject invalid offer id", func() {
		s.SetupTest()

		for _, assetType := range supportedAssetTypes {
			s.Require().ErrorContains(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, "asset", assetType, "@"),
				"invalid Buy-Order ID",
			)
		}
	})

	s.Run("fail - should reject invalid asset id", func() {
		s.SetupTest()

		for _, assetType := range supportedAssetTypes {
			var wantErrContains string
			switch assetType {
			case dymnstypes.TypeName:
				wantErrContains = "invalid Dym-Name"
			case dymnstypes.TypeAlias:
				wantErrContains = "invalid Alias"
			default:
				s.T().Fatalf("unsupported asset type: %s", assetType)
			}
			s.Require().ErrorContains(
				s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, "@", assetType, dymnstypes.CreateBuyOrderId(assetType, 1)),
				wantErrContains,
			)
		}
	})

	s.Run("fail - should reject invalid asset type", func() {
		s.SetupTest()

		s.Require().ErrorContains(
			s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, "@", dymnstypes.AssetType_AT_UNKNOWN, "101"),
			"invalid asset type",
		)

		for i := int32(0); i < 99; i++ {
			assetType := dymnstypes.AssetType(i)

			if slices.Contains(supportedAssetTypes, dymnstypes.AssetType(i)) {
				continue
			}

			s.Require().ErrorContains(
				s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, "@", assetType, "101"),
				"invalid asset type",
			)
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAddReverseMappingAssetIdToBuyOrder_Type_DymName() {
	_, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, "@")
	s.Require().Error(
		err,
		"fail - should reject invalid Dym-Name",
	)

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))

	offer11 := dymnstypes.BuyOrder{
		Id:         "1011",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         "1012",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer12))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer11.Id))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer12.Id))

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName2))

	offer2 := dymnstypes.BuyOrder{
		Id:         "102",
		AssetId:    dymName2.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer2))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName2.Name, dymnstypes.TypeName, offer2.Id))

	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, "1012356215631"),
		"no check non-existing offer id",
	)

	s.Run("no error if duplicated name", func() {
		for i := 0; i < 3; i++ {
			s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName2.Name, dymnstypes.TypeName, offer2.Id))
		}
	})

	linked1, err1 := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
	s.Require().NoError(err1)
	s.Require().Len(linked1, 2)
	s.Require().Equal(offer11.Id, linked1[0].Id)
	s.Require().Equal(offer12.Id, linked1[1].Id)

	linked2, err2 := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName2.Name)
	s.Require().NoError(err2)
	s.Require().NotEqual(2, len(linked2), "should not include non-existing offers")
	s.Require().Len(linked2, 1)
	s.Require().Equal(offer2.Id, linked2[0].Id)

	linkedByNotExists, err3 := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, "non-exists")
	s.Require().NoError(err3)
	s.Require().Len(linkedByNotExists, 0)

	s.Run("should be linked regardless of expired Dym-Name", func() {
		dymName1.ExpireAt = time.Now().UTC().Add(-time.Hour).Unix()
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))

		linked1, err1 = s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
		s.Require().NoError(err1)
		s.Require().Len(linked1, 2)
		s.Require().Equal(offer11.Id, linked1[0].Id)
		s.Require().Equal(offer12.Id, linked1[1].Id)
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAddReverseMappingAssetIdToBuyOrder_Type_Alias() {
	_, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, "@")
	s.Require().Error(
		err,
		"fail - should reject invalid Alias",
	)

	const alias1 = "one"

	const dstRollAppId = "rollapp_3-2"

	buyerA := testAddr(1).bech32()

	offer11 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 11),
		AssetId:    alias1,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 12),
		AssetId:    alias1,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer12))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer11.Id))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer12.Id))

	const alias2 = "two"

	offer2 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    alias2,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer2))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias2, dymnstypes.TypeAlias, offer2.Id))

	s.Require().NoError(
		s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias2, dymnstypes.TypeAlias, "2012356215631"),
		"no check non-existing offer id",
	)

	s.Run("no error if duplicated name", func() {
		for i := 0; i < 3; i++ {
			s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias2, dymnstypes.TypeAlias, offer2.Id))
		}
	})

	linked1, err1 := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias1)
	s.Require().NoError(err1)
	s.Require().Len(linked1, 2)
	s.Require().Equal(offer11.Id, linked1[0].Id)
	s.Require().Equal(offer12.Id, linked1[1].Id)

	linked2, err2 := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias2)
	s.Require().NoError(err2)
	s.Require().NotEqual(2, len(linked2), "should not include non-existing offers")
	s.Require().Len(linked2, 1)
	s.Require().Equal(offer2.Id, linked2[0].Id)

	linkedByNotExists, err3 := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, "nah")
	s.Require().NoError(err3)
	s.Require().Len(linkedByNotExists, 0)
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingAssetIdToBuyOrder_Generic() {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	s.Run("pass - can remove", func() {
		s.SetupTest()

		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, "asset", assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			s.Require().NoError(err)
		}

		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, "asset", assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			s.Require().NoError(err)
		}
	})

	s.Run("pass - can remove of non-exists", func() {
		s.SetupTest()

		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, "asset", assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			s.Require().NoError(err)
		}
	})

	s.Run("pass - can remove without collision across asset types", func() {
		s.SetupTest()

		const assetId = "asset"

		for _, assetType := range supportedAssetTypes {
			err := s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, assetId, assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			s.Require().NoError(err)
		}

		err := s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, assetId, dymnstypes.TypeName, dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 1))
		s.Require().NoError(err)

		s.Require().Empty(s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId)).OrderIds)
		s.Require().NotEmpty(s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, dymnstypes.AliasToBuyOrderIdsRvlKey(assetId)).OrderIds)
	})

	s.Run("fail - should reject invalid offer id", func() {
		s.SetupTest()

		for _, assetType := range supportedAssetTypes {
			s.Require().ErrorContains(
				s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, "asset", assetType, "@"),
				"invalid Buy-Order ID",
			)
		}
	})

	s.Run("fail - should reject invalid asset id", func() {
		s.SetupTest()

		for _, assetType := range supportedAssetTypes {
			var wantErrContains string
			switch assetType {
			case dymnstypes.TypeName:
				wantErrContains = "invalid Dym-Name"
			case dymnstypes.TypeAlias:
				wantErrContains = "invalid Alias"
			default:
				s.T().Fatalf("unsupported asset type: %s", assetType)
			}
			s.Require().ErrorContains(
				s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, "@", assetType, dymnstypes.CreateBuyOrderId(assetType, 1)),
				wantErrContains,
			)
		}
	})

	s.Run("fail - should reject invalid asset type", func() {
		s.SetupTest()

		s.Require().ErrorContains(
			s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, "@", dymnstypes.AssetType_AT_UNKNOWN, "101"),
			"invalid asset type",
		)

		for i := int32(0); i < 99; i++ {
			assetType := dymnstypes.AssetType(i)

			if slices.Contains(supportedAssetTypes, dymnstypes.AssetType(i)) {
				continue
			}

			s.Require().ErrorContains(
				s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, "@", assetType, "101"),
				"invalid asset type",
			)
		}
	})
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingAssetIdToBuyOrder_Type_DymName() {
	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))

	offer11 := dymnstypes.BuyOrder{
		Id:         "1011",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         "1012",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer12))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer11.Id))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer12.Id))

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName2))

	offer2 := dymnstypes.BuyOrder{
		Id:         "102",
		AssetId:    dymName2.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer2))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName2.Name, dymnstypes.TypeName, offer2.Id))

	s.Run("no error if remove a record that not linked", func() {
		linked, _ := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
		s.Require().Len(linked, 2)

		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer2.Id))

		linked, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
		s.Require().NoError(err)
		s.Require().Len(linked, 2, "existing data must be kept")
	})

	s.Run("no error if element is not in the list", func() {
		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, "10218362184621"))

		linked, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
		s.Require().NoError(err)
		s.Require().Len(linked, 2, "existing data must be kept")
	})

	s.Run("remove correctly", func() {
		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer11.Id))

		linked, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
		s.Require().NoError(err)
		s.Require().Len(linked, 1)
		s.Require().Equal(offer12.Id, linked[0].Id)

		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, dymName1.Name, dymnstypes.TypeName, offer12.Id))

		linked, err = s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName1.Name)
		s.Require().NoError(err)
		s.Require().Empty(linked)
	})

	linked, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName2.Name)
	s.Require().NoError(err)
	s.Require().Len(linked, 1)
}

func (s *KeeperTestSuite) TestKeeper_RemoveReverseMappingAssetIdToBuyOrder_Type_Alias() {
	buyerA := testAddr(1).bech32()

	const alias1 = "one"

	const dstRollAppId = "rollapp_3-2"

	offer11 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 11),
		AssetId:    alias1,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 12),
		AssetId:    alias1,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer12))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer11.Id))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer12.Id))

	const alias2 = "two"

	offer2 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    alias2,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	s.Require().NoError(s.dymNsKeeper.SetBuyOrder(s.ctx, offer2))

	s.Require().NoError(s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, alias2, dymnstypes.TypeAlias, offer2.Id))

	s.Run("no error if remove a record that not linked", func() {
		linked, _ := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias1)
		s.Require().Len(linked, 2)

		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer2.Id))

		linked, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias1)
		s.Require().NoError(err)
		s.Require().Len(linked, 2, "existing data must be kept")
	})

	s.Run("no error if element is not in the list", func() {
		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, "20218362184621"))

		linked, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias1)
		s.Require().NoError(err)
		s.Require().Len(linked, 2, "existing data must be kept")
	})

	s.Run("remove correctly", func() {
		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer11.Id))

		linked, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias1)
		s.Require().NoError(err)
		s.Require().Len(linked, 1)
		s.Require().Equal(offer12.Id, linked[0].Id)

		s.Require().NoError(s.dymNsKeeper.RemoveReverseMappingAssetIdToBuyOrder(s.ctx, alias1, dymnstypes.TypeAlias, offer12.Id))

		linked, err = s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias1)
		s.Require().NoError(err)
		s.Require().Empty(linked)
	})

	linked, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, alias2)
	s.Require().NoError(err)
	s.Require().Len(linked, 1)
}
