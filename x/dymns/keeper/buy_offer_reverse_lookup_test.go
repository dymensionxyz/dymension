package keeper_test

import (
	"testing"
	"time"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetAddReverseMappingBuyerToPlacedBuyOffer(t *testing.T) {
	// TODO DymNS: add test for Sell/Buy Alias

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyer1a := testAddr(1).bech32()
	buyer2a := testAddr(2).bech32()
	someoneA := testAddr(3).bech32()

	require.Error(
		t,
		dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, "0x", "101"),
		"should not allow invalid buyer address",
	)

	require.Error(
		t,
		dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer1a, "@"),
		"should not allow invalid offer ID",
	)

	_, err := dk.GetBuyOffersByBuyer(ctx, "0x")
	require.Error(
		t,
		err,
		"should not allow invalid buyer address",
	)

	offer1 := dymnstypes.BuyOffer{
		Id:                     "101",
		Name:                   "a",
		Type:                   dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:                  buyer1a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer1))
	err = dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer1a, offer1.Id)
	require.NoError(t, err)

	offer2 := dymnstypes.BuyOffer{
		Id:                     "102",
		Name:                   "b",
		Type:                   dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:                  buyer2a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer2))
	err = dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer2a, offer2.Id)
	require.NoError(t, err)

	offer3 := dymnstypes.BuyOffer{
		Id:                     "103",
		Name:                   "c",
		Type:                   dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:                  buyer2a,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer3))
	err = dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer2a, offer3.Id)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer2a, "103721461"),
		"no check non-existing offer record",
	)

	t.Run("no error if duplicated ID", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer2a, offer2.Id),
			)
		}
	})

	placedBy1, err1 := dk.GetBuyOffersByBuyer(ctx, buyer1a)
	require.NoError(t, err1)
	require.Len(t, placedBy1, 1)

	placedBy2, err2 := dk.GetBuyOffersByBuyer(ctx, buyer2a)
	require.NoError(t, err2)
	require.NotEqual(t, 3, len(placedBy2), "should not include non-existing offers")
	require.Len(t, placedBy2, 2)

	placedByNonExists, err3 := dk.GetDymNamesOwnedBy(ctx, someoneA)
	require.NoError(t, err3)
	require.Len(t, placedByNonExists, 0)

	require.NoError(
		t,
		dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyer2a, offer1.Id),
		"no error if offer placed by another buyer",
	)
	placedBy2, err2 = dk.GetBuyOffersByBuyer(ctx, buyer2a)
	require.NoError(t, err2)
	require.Len(t, placedBy2, 2, "should not include offers placed by another buyer")
}

func TestKeeper_RemoveReverseMappingBuyerToPlacedBuyOffer(t *testing.T) {
	// TODO DymNS: add test for Sell/Buy Alias

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	buyerA := testAddr(1).bech32()
	someoneA := testAddr(2).bech32()

	require.Error(
		t,
		dk.RemoveReverseMappingBuyerToBuyOffer(ctx, "0x", "101"),
		"should not allow invalid buyer address",
	)

	require.Error(
		t,
		dk.RemoveReverseMappingBuyerToBuyOffer(ctx, buyerA, "@"),
		"should not allow invalid offer ID",
	)

	offer1 := dymnstypes.BuyOffer{
		Id:                     "101",
		Name:                   "a",
		Type:                   dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:                  buyerA,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer1))
	require.NoError(t, dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyerA, offer1.Id))

	offer2 := dymnstypes.BuyOffer{
		Id:                     "102",
		Name:                   "b",
		Type:                   dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:                  buyerA,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer2))
	require.NoError(t, dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, buyerA, offer2.Id))

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOffer(ctx, someoneA, offer1.Id),
		"no error if buyer non-exists",
	)

	placedByBuyer, err := dk.GetBuyOffersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 2, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOffer(ctx, buyerA, "10138132187"),
		"no error if not placed order",
	)

	placedByBuyer, err = dk.GetBuyOffersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 2, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOffer(ctx, buyerA, offer1.Id),
	)
	placedByBuyer, err = dk.GetBuyOffersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 1)
	require.Equal(t, offer2.Id, placedByBuyer[0].Id)

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToBuyOffer(ctx, buyerA, offer2.Id),
	)
	placedByBuyer, err = dk.GetBuyOffersByBuyer(ctx, buyerA)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 0)
}

func TestKeeper_GetAddReverseMappingDymNameToBuyOffer(t *testing.T) {
	// TODO DymNS: add test for Sell/Buy Alias

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, "@", "101"),
		"fail - should reject invalid Dym-Name",
	)
	require.Error(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, "a", "@"),
		"fail - should reject invalid offer-id",
	)

	_, err := dk.GetBuyOffersOfDymName(ctx, "@")
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

	offer11 := dymnstypes.BuyOffer{
		Id:         "1011",
		Name:       dymName1.Name,
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer11))

	offer12 := dymnstypes.BuyOffer{
		Id:         "1012",
		Name:       dymName1.Name,
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer12.Id),
	)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	offer2 := dymnstypes.BuyOffer{
		Id:         "102",
		Name:       dymName2.Name,
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName2.Name, offer2.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, "1012356215631"),
		"no check non-existing offer id",
	)

	t.Run("no error if duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName2.Name, offer2.Id),
			)
		}
	})

	linked1, err1 := dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
	require.NoError(t, err1)
	require.Len(t, linked1, 2)
	require.Equal(t, offer11.Id, linked1[0].Id)
	require.Equal(t, offer12.Id, linked1[1].Id)

	linked2, err2 := dk.GetBuyOffersOfDymName(ctx, dymName2.Name)
	require.NoError(t, err2)
	require.NotEqual(t, 2, len(linked2), "should not include non-existing offers")
	require.Len(t, linked2, 1)
	require.Equal(t, offer2.Id, linked2[0].Id)

	linkedByNotExists, err3 := dk.GetBuyOffersOfDymName(ctx, "non-exists")
	require.NoError(t, err3)
	require.Len(t, linkedByNotExists, 0)

	t.Run("should be linked regardless of expired Dym-Name", func(t *testing.T) {
		dymName1.ExpireAt = time.Now().UTC().Add(-time.Hour).Unix()
		require.NoError(t, dk.SetDymName(ctx, dymName1))

		linked1, err1 = dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err1)
		require.Len(t, linked1, 2)
		require.Equal(t, offer11.Id, linked1[0].Id)
		require.Equal(t, offer12.Id, linked1[1].Id)
	})
}

func TestKeeper_RemoveReverseMappingDymNameToBuyOffer(t *testing.T) {
	// TODO DymNS: add test for Sell/Buy Alias

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.RemoveReverseMappingDymNameToBuyOffer(ctx, "@", "101"),
		"fail - should reject invalid Dym-Name",
	)
	require.Error(
		t,
		dk.RemoveReverseMappingDymNameToBuyOffer(ctx, "a", "@"),
		"fail - should reject invalid offer-id",
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

	offer11 := dymnstypes.BuyOffer{
		Id:         "1011",
		Name:       dymName1.Name,
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer11))

	offer12 := dymnstypes.BuyOffer{
		Id:         "1012",
		Name:       dymName1.Name,
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer12.Id),
	)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	offer2 := dymnstypes.BuyOffer{
		Id:         "102",
		Name:       dymName2.Name,
		Type:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetBuyOffer(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToBuyOffer(ctx, dymName2.Name, offer2.Id),
	)

	t.Run("no error if remove a record that not linked", func(t *testing.T) {
		linked, _ := dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
		require.Len(t, linked, 2)

		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer2.Id),
		)

		linked, err := dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("no error if element is not in the list", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, "10218362184621"),
		)

		linked, err := dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("remove correctly", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer11.Id),
		)

		linked, err := dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 1)
		require.Equal(t, offer12.Id, linked[0].Id)

		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToBuyOffer(ctx, dymName1.Name, offer12.Id),
		)

		linked, err = dk.GetBuyOffersOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	linked, err := dk.GetBuyOffersOfDymName(ctx, dymName2.Name)
	require.NoError(t, err)
	require.Len(t, linked, 1)
}
