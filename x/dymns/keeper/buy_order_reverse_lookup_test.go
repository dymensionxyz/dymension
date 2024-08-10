package keeper_test

import (
	"testing"
	"time"

	"golang.org/x/exp/slices"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetAddReverseMappingBuyerToPlacedBuyOrder(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyer1a := testAddr(1).bech32()
	buyer2a := testAddr(2).bech32()
	buyer3a := testAddr(3).bech32()
	someoneA := testAddr(4).bech32()

	require.Error(
		t,
		dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, "0x", "101"),
		"should not allow invalid buyer address",
	)

	require.Error(
		t,
		dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer1a, "@"),
		"should not allow invalid offer ID",
	)

	_, err := dk.GetBuyOrdersByBuyer(ctx, "0x")
	require.Error(
		t,
		err,
		"should not allow invalid buyer address",
	)

	offer1 := dymnstypes.BuyOrder{
		Id:                     "101",
		AssetId:                "a",
		AssetType:              dymnstypes.TypeName,
		Buyer:                  buyer1a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer1))
	err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer1a, offer1.Id)
	require.NoError(t, err)

	offer2 := dymnstypes.BuyOrder{
		Id:                     "202",
		AssetId:                "alias",
		AssetType:              dymnstypes.TypeAlias,
		Params:                 []string{"rollapp_1-1"},
		Buyer:                  buyer2a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer2))
	err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer2a, offer2.Id)
	require.NoError(t, err)

	offer3 := dymnstypes.BuyOrder{
		Id:                     "103",
		AssetId:                "c",
		AssetType:              dymnstypes.TypeName,
		Buyer:                  buyer2a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer3))
	err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer2a, offer3.Id)
	require.NoError(t, err)

	offer4 := dymnstypes.BuyOrder{
		Id:                     "204",
		AssetId:                "salas",
		AssetType:              dymnstypes.TypeAlias,
		Params:                 []string{"rollapp_2-2"},
		Buyer:                  buyer3a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer4))
	err = dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer3a, offer4.Id)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer2a, "103721461"),
		"no check non-existing offer record",
	)

	t.Run("no error if duplicated ID", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer2a, offer2.Id),
			)
		}
	})

	placedBy1, err1 := dk.GetBuyOrdersByBuyer(ctx, buyer1a)
	require.NoError(t, err1)
	require.Len(t, placedBy1, 1)

	placedBy2, err2 := dk.GetBuyOrdersByBuyer(ctx, buyer2a)
	require.NoError(t, err2)
	require.NotEqual(t, 3, len(placedBy2), "should not include non-existing offers")
	require.Len(t, placedBy2, 2)

	placedBy3, err3 := dk.GetBuyOrdersByBuyer(ctx, buyer3a)
	require.NoError(t, err3)
	require.Len(t, placedBy3, 1)

	placedByNonExists, err3 := dk.GetDymNamesOwnedBy(ctx, someoneA)
	require.NoError(t, err3)
	require.Len(t, placedByNonExists, 0)

	require.NoError(
		t,
		dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer2a, offer1.Id),
		"no error if offer placed by another buyer",
	)
	require.NoError(
		t,
		dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyer2a, offer4.Id),
		"no error if offer placed by another buyer",
	)
	placedBy2, err2 = dk.GetBuyOrdersByBuyer(ctx, buyer2a)
	require.NoError(t, err2)
	require.Len(t, placedBy2, 2, "should not include offers placed by another buyer")
}

func TestKeeper_RemoveReverseMappingBuyerToPlacedBuyOrder(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyerA := testAddr(1).bech32()
	someoneA := testAddr(2).bech32()

	require.Error(
		t,
		dk.RemoveReverseMappingBuyerToBuyOrder(ctx, "0x", "101"),
		"should not allow invalid buyer address",
	)

	require.Error(
		t,
		dk.RemoveReverseMappingBuyerToBuyOrder(ctx, buyerA, "@"),
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
	require.NoError(t, dk.SetBuyOrder(ctx, offer1))
	require.NoError(t, dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyerA, offer1.Id))

	offer2 := dymnstypes.BuyOrder{
		Id:                     "202",
		AssetId:                "alias",
		AssetType:              dymnstypes.TypeAlias,
		Params:                 []string{"rollapp_1-1"},
		Buyer:                  buyerA,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer2))
	require.NoError(t, dk.AddReverseMappingBuyerToBuyOrderRecord(ctx, buyerA, offer2.Id))

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOrder(ctx, someoneA, offer1.Id),
		"no error if buyer non-exists",
	)

	placedByBuyer, err := dk.GetBuyOrdersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 2, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOrder(ctx, buyerA, "10138132187"),
		"no error if not placed order",
	)

	placedByBuyer, err = dk.GetBuyOrdersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 2, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOrder(ctx, buyerA, offer1.Id),
	)
	placedByBuyer, err = dk.GetBuyOrdersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 1)
	require.Equal(t, offer2.Id, placedByBuyer[0].Id)

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOrder(ctx, buyerA, offer2.Id),
	)
	placedByBuyer, err = dk.GetBuyOrdersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 0)
}

func TestKeeper_AddReverseMappingAssetIdToBuyOrder_Generic(t *testing.T) {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	t.Run("pass - can add", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		const assetId = "asset"

		for _, assetType := range supportedAssetTypes {
			err := dk.AddReverseMappingAssetIdToBuyOrder(ctx, assetId, assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			require.NoError(t, err)
		}

		require.NotEmpty(t, dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId)).OrderIds)
		require.NotEmpty(t, dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, dymnstypes.AliasToBuyOrderIdsRvlKey(assetId)).OrderIds)
	})

	t.Run("pass - can add without collision across asset types", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		const assetId = "asset"

		for _, assetType := range supportedAssetTypes {
			err := dk.AddReverseMappingAssetIdToBuyOrder(ctx, assetId, assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			require.NoError(t, err)
		}

		err := dk.AddReverseMappingAssetIdToBuyOrder(ctx, assetId, dymnstypes.TypeName, dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 2))
		require.NoError(t, err)

		err = dk.AddReverseMappingAssetIdToBuyOrder(ctx, assetId, dymnstypes.TypeAlias, dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 3))
		require.NoError(t, err)

		nameOffers := dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId))
		require.Len(t, nameOffers.OrderIds, 2)
		aliasOffers := dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, dymnstypes.AliasToBuyOrderIdsRvlKey(assetId))
		require.Len(t, aliasOffers.OrderIds, 2)

		require.NotEqual(t, nameOffers.OrderIds, aliasOffers.OrderIds, "data must be persisted separately")

		require.Equal(t, dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 1), nameOffers.OrderIds[0])
		require.Equal(t, dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 2), nameOffers.OrderIds[1])
		require.Equal(t, dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 1), aliasOffers.OrderIds[0])
		require.Equal(t, dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 3), aliasOffers.OrderIds[1])
	})

	t.Run("fail - should reject invalid offer id", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			require.ErrorContains(t,
				dk.AddReverseMappingAssetIdToBuyOrder(ctx, "asset", assetType, "@"),
				"invalid Buy-Order ID",
			)
		}
	})

	t.Run("fail - should reject invalid asset id", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			var wantErrContains string
			switch assetType {
			case dymnstypes.TypeName:
				wantErrContains = "invalid Dym-Name"
			case dymnstypes.TypeAlias:
				wantErrContains = "invalid Alias"
			default:
				t.Fatalf("unsupported asset type: %s", assetType)
			}
			require.ErrorContains(
				t,
				dk.AddReverseMappingAssetIdToBuyOrder(ctx, "@", assetType, dymnstypes.CreateBuyOrderId(assetType, 1)),
				wantErrContains,
			)
		}
	})

	t.Run("fail - should reject invalid asset type", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.ErrorContains(t,
			dk.AddReverseMappingAssetIdToBuyOrder(ctx, "@", dymnstypes.AssetType_AT_UNKNOWN, "101"),
			"invalid asset type",
		)

		for i := int32(0); i < 99; i++ {
			assetType := dymnstypes.AssetType(i)

			if slices.Contains(supportedAssetTypes, dymnstypes.AssetType(i)) {
				continue
			}

			require.ErrorContains(t,
				dk.AddReverseMappingAssetIdToBuyOrder(ctx, "@", assetType, "101"),
				"invalid asset type",
			)
		}
	})
}

func TestKeeper_GetAddReverseMappingAssetIdToBuyOrder_Type_DymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	_, err := dk.GetBuyOrdersOfDymName(ctx, "@")
	require.Error(
		t,
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
	require.NoError(t, dk.SetDymName(ctx, dymName1))

	offer11 := dymnstypes.BuyOrder{
		Id:         "1011",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         "1012",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer12.Id),
	)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	offer2 := dymnstypes.BuyOrder{
		Id:         "102",
		AssetId:    dymName2.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName2.Name, dymnstypes.TypeName, offer2.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, "1012356215631"),
		"no check non-existing offer id",
	)

	t.Run("no error if duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName2.Name, dymnstypes.TypeName, offer2.Id),
			)
		}
	})

	linked1, err1 := dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
	require.NoError(t, err1)
	require.Len(t, linked1, 2)
	require.Equal(t, offer11.Id, linked1[0].Id)
	require.Equal(t, offer12.Id, linked1[1].Id)

	linked2, err2 := dk.GetBuyOrdersOfDymName(ctx, dymName2.Name)
	require.NoError(t, err2)
	require.NotEqual(t, 2, len(linked2), "should not include non-existing offers")
	require.Len(t, linked2, 1)
	require.Equal(t, offer2.Id, linked2[0].Id)

	linkedByNotExists, err3 := dk.GetBuyOrdersOfDymName(ctx, "non-exists")
	require.NoError(t, err3)
	require.Len(t, linkedByNotExists, 0)

	t.Run("should be linked regardless of expired Dym-Name", func(t *testing.T) {
		dymName1.ExpireAt = time.Now().UTC().Add(-time.Hour).Unix()
		require.NoError(t, dk.SetDymName(ctx, dymName1))

		linked1, err1 = dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err1)
		require.Len(t, linked1, 2)
		require.Equal(t, offer11.Id, linked1[0].Id)
		require.Equal(t, offer12.Id, linked1[1].Id)
	})
}

func TestKeeper_GetAddReverseMappingAssetIdToBuyOrder_Type_Alias(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	_, err := dk.GetBuyOrdersOfAlias(ctx, "@")
	require.Error(
		t,
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
	require.NoError(t, dk.SetBuyOrder(ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 12),
		AssetId:    alias1,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer12.Id),
	)

	const alias2 = "two"

	offer2 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    alias2,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias2, dymnstypes.TypeAlias, offer2.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias2, dymnstypes.TypeAlias, "2012356215631"),
		"no check non-existing offer id",
	)

	t.Run("no error if duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias2, dymnstypes.TypeAlias, offer2.Id),
			)
		}
	})

	linked1, err1 := dk.GetBuyOrdersOfAlias(ctx, alias1)
	require.NoError(t, err1)
	require.Len(t, linked1, 2)
	require.Equal(t, offer11.Id, linked1[0].Id)
	require.Equal(t, offer12.Id, linked1[1].Id)

	linked2, err2 := dk.GetBuyOrdersOfAlias(ctx, alias2)
	require.NoError(t, err2)
	require.NotEqual(t, 2, len(linked2), "should not include non-existing offers")
	require.Len(t, linked2, 1)
	require.Equal(t, offer2.Id, linked2[0].Id)

	linkedByNotExists, err3 := dk.GetBuyOrdersOfAlias(ctx, "nah")
	require.NoError(t, err3)
	require.Len(t, linkedByNotExists, 0)
}

func TestKeeper_RemoveReverseMappingAssetIdToBuyOrder_Generic(t *testing.T) {
	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	t.Run("pass - can remove", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			err := dk.AddReverseMappingAssetIdToBuyOrder(ctx, "asset", assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			require.NoError(t, err)
		}

		for _, assetType := range supportedAssetTypes {
			err := dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, "asset", assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			require.NoError(t, err)
		}
	})

	t.Run("pass - can remove of non-exists", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			err := dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, "asset", assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			require.NoError(t, err)
		}
	})

	t.Run("pass - can remove without collision across asset types", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		const assetId = "asset"

		for _, assetType := range supportedAssetTypes {
			err := dk.AddReverseMappingAssetIdToBuyOrder(ctx, assetId, assetType, dymnstypes.CreateBuyOrderId(assetType, 1))
			require.NoError(t, err)
		}

		err := dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, assetId, dymnstypes.TypeName, dymnstypes.CreateBuyOrderId(dymnstypes.TypeName, 1))
		require.NoError(t, err)

		require.Empty(t, dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, dymnstypes.DymNameToBuyOrderIdsRvlKey(assetId)).OrderIds)
		require.NotEmpty(t, dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, dymnstypes.AliasToBuyOrderIdsRvlKey(assetId)).OrderIds)
	})

	t.Run("fail - should reject invalid offer id", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			require.ErrorContains(t,
				dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, "asset", assetType, "@"),
				"invalid Buy-Order ID",
			)
		}
	})

	t.Run("fail - should reject invalid asset id", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		for _, assetType := range supportedAssetTypes {
			var wantErrContains string
			switch assetType {
			case dymnstypes.TypeName:
				wantErrContains = "invalid Dym-Name"
			case dymnstypes.TypeAlias:
				wantErrContains = "invalid Alias"
			default:
				t.Fatalf("unsupported asset type: %s", assetType)
			}
			require.ErrorContains(
				t,
				dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, "@", assetType, dymnstypes.CreateBuyOrderId(assetType, 1)),
				wantErrContains,
			)
		}
	})

	t.Run("fail - should reject invalid asset type", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.ErrorContains(t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, "@", dymnstypes.AssetType_AT_UNKNOWN, "101"),
			"invalid asset type",
		)

		for i := int32(0); i < 99; i++ {
			assetType := dymnstypes.AssetType(i)

			if slices.Contains(supportedAssetTypes, dymnstypes.AssetType(i)) {
				continue
			}

			require.ErrorContains(t,
				dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, "@", assetType, "101"),
				"invalid asset type",
			)
		}
	})
}

func TestKeeper_RemoveReverseMappingAssetIdToBuyOrder_Type_DymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))

	offer11 := dymnstypes.BuyOrder{
		Id:         "1011",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         "1012",
		AssetId:    dymName1.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer12.Id),
	)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	offer2 := dymnstypes.BuyOrder{
		Id:         "102",
		AssetId:    dymName2.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, dymName2.Name, dymnstypes.TypeName, offer2.Id),
	)

	t.Run("no error if remove a record that not linked", func(t *testing.T) {
		linked, _ := dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
		require.Len(t, linked, 2)

		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer2.Id),
		)

		linked, err := dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("no error if element is not in the list", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, "10218362184621"),
		)

		linked, err := dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("remove correctly", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer11.Id),
		)

		linked, err := dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 1)
		require.Equal(t, offer12.Id, linked[0].Id)

		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, dymName1.Name, dymnstypes.TypeName, offer12.Id),
		)

		linked, err = dk.GetBuyOrdersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	linked, err := dk.GetBuyOrdersOfDymName(ctx, dymName2.Name)
	require.NoError(t, err)
	require.Len(t, linked, 1)
}

func TestKeeper_RemoveReverseMappingAssetIdToBuyOrder_Type_Alias(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

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
	require.NoError(t, dk.SetBuyOrder(ctx, offer11))

	offer12 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 12),
		AssetId:    alias1,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer12.Id),
	)

	const alias2 = "two"

	offer2 := dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    alias2,
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{dstRollAppId},
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOrder(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingAssetIdToBuyOrder(ctx, alias2, dymnstypes.TypeAlias, offer2.Id),
	)

	t.Run("no error if remove a record that not linked", func(t *testing.T) {
		linked, _ := dk.GetBuyOrdersOfAlias(ctx, alias1)
		require.Len(t, linked, 2)

		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer2.Id),
		)

		linked, err := dk.GetBuyOrdersOfAlias(ctx, alias1)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("no error if element is not in the list", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, "20218362184621"),
		)

		linked, err := dk.GetBuyOrdersOfAlias(ctx, alias1)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("remove correctly", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer11.Id),
		)

		linked, err := dk.GetBuyOrdersOfAlias(ctx, alias1)
		require.NoError(t, err)
		require.Len(t, linked, 1)
		require.Equal(t, offer12.Id, linked[0].Id)

		require.NoError(
			t,
			dk.RemoveReverseMappingAssetIdToBuyOrder(ctx, alias1, dymnstypes.TypeAlias, offer12.Id),
		)

		linked, err = dk.GetBuyOrdersOfAlias(ctx, alias1)
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	linked, err := dk.GetBuyOrdersOfAlias(ctx, alias2)
	require.NoError(t, err)
	require.Len(t, linked, 1)
}
