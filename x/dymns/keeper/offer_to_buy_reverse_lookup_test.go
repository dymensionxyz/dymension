package keeper_test

import (
	"testing"
	"time"

	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetAddReverseMappingBuyerToPlacedOfferToBuy(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, "0x", "1"),
		"should not allow invalid buyer address",
	)

	require.Error(
		t,
		dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue", "@"),
		"should not allow invalid offer ID",
	)

	_, err := dk.GetOfferToBuyByBuyer(ctx, "0x")
	require.Error(
		t,
		err,
		"should not allow invalid buyer address",
	)

	buyer1 := "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	buyer2 := "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d"

	offer1 := dymnstypes.OfferToBuy{
		Id:                     "1",
		Name:                   "a",
		Buyer:                  buyer1,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer1))
	err = dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer1, offer1.Id)
	require.NoError(t, err)

	offer2 := dymnstypes.OfferToBuy{
		Id:                     "2",
		Name:                   "b",
		Buyer:                  buyer2,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer2))
	err = dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer2, offer2.Id)
	require.NoError(t, err)

	offer3 := dymnstypes.OfferToBuy{
		Id:                     "3",
		Name:                   "c",
		Buyer:                  buyer2,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer3))
	err = dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer2, offer3.Id)
	require.NoError(t, err)

	require.NoError(
		t,
		dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer2, "3721461"),
		"no check non-existing offer record",
	)

	t.Run("no error if duplicated ID", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer2, offer2.Id),
			)
		}
	})

	placedBy1, err1 := dk.GetOfferToBuyByBuyer(ctx, buyer1)
	require.NoError(t, err1)
	require.Len(t, placedBy1, 1)

	placedBy2, err2 := dk.GetOfferToBuyByBuyer(ctx, buyer2)
	require.NoError(t, err2)
	require.NotEqual(t, 3, len(placedBy2), "should not include non-existing offers")
	require.Len(t, placedBy2, 2)

	placedByNonExists, err3 := dk.GetDymNamesOwnedBy(ctx, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96")
	require.NoError(t, err3)
	require.Len(t, placedByNonExists, 0)

	require.NoError(
		t,
		dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer2, offer1.Id),
		"no error if offer placed by another buyer",
	)
	placedBy2, err2 = dk.GetOfferToBuyByBuyer(ctx, buyer2)
	require.NoError(t, err2)
	require.Len(t, placedBy2, 2, "should not include offers placed by another buyer")
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_RemoveReverseMappingBuyerToPlacedOfferToBuy(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.RemoveReverseMappingBuyerToOfferToBuy(ctx, "0x", "1"),
		"should not allow invalid buyer address",
	)

	require.Error(
		t,
		dk.RemoveReverseMappingBuyerToOfferToBuy(ctx, "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue", "@"),
		"should not allow invalid offer ID",
	)

	const buyer = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	offer1 := dymnstypes.OfferToBuy{
		Id:                     "1",
		Name:                   "a",
		Buyer:                  buyer,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer1))
	require.NoError(t, dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer, offer1.Id))

	offer2 := dymnstypes.OfferToBuy{
		Id:                     "2",
		Name:                   "b",
		Buyer:                  buyer,
		OfferPrice:             dymnsutils.TestCoin(1),
		CounterpartyOfferPrice: nil,
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer2))
	require.NoError(t, dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, buyer, offer2.Id))

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToOfferToBuy(ctx, "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4", offer1.Id),
		"no error if buyer non-exists",
	)

	placedByBuyer, err := dk.GetOfferToBuyByBuyer(ctx, buyer)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 2, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToOfferToBuy(ctx, buyer, "138132187"),
		"no error if not placed order",
	)

	placedByBuyer, err = dk.GetOfferToBuyByBuyer(ctx, buyer)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 2, "existing data must be kept")

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToOfferToBuy(ctx, buyer, offer1.Id),
	)
	placedByBuyer, err = dk.GetOfferToBuyByBuyer(ctx, buyer)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 1)
	require.Equal(t, offer2.Id, placedByBuyer[0].Id)

	require.NoError(
		t,
		dk.RemoveReverseMappingBuyerToOfferToBuy(ctx, buyer, offer2.Id),
	)
	placedByBuyer, err = dk.GetOfferToBuyByBuyer(ctx, buyer)
	require.NoError(t, err)
	require.Len(t, placedByBuyer, 0)
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetAddReverseMappingDymNameToOfferToBuy(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, "@", "1"),
		"fail - should reject invalid Dym-Name",
	)
	require.Error(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, "a", "@"),
		"fail - should reject invalid offer-id",
	)

	_, err := dk.GetOffersToBuyOfDymName(ctx, "@")
	require.Error(
		t,
		err,
		"fail - should reject invalid Dym-Name",
	)

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const buyer = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))

	offer11 := dymnstypes.OfferToBuy{
		Id:         "11",
		Name:       dymName1.Name,
		Buyer:      buyer,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer11))

	offer12 := dymnstypes.OfferToBuy{
		Id:         "12",
		Name:       dymName1.Name,
		Buyer:      buyer,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer12.Id),
	)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	offer2 := dymnstypes.OfferToBuy{
		Id:         "2",
		Name:       dymName2.Name,
		Buyer:      buyer,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName2.Name, offer2.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, "12356215631"),
		"no check non-existing offer id",
	)

	t.Run("no error if duplicated name", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			require.NoError(t,
				dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName2.Name, offer2.Id),
			)
		}
	})

	linked1, err1 := dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
	require.NoError(t, err1)
	require.Len(t, linked1, 2)
	require.Equal(t, offer11.Id, linked1[0].Id)
	require.Equal(t, offer12.Id, linked1[1].Id)

	linked2, err2 := dk.GetOffersToBuyOfDymName(ctx, dymName2.Name)
	require.NoError(t, err2)
	require.NotEqual(t, 2, len(linked2), "should not include non-existing offers")
	require.Len(t, linked2, 1)
	require.Equal(t, offer2.Id, linked2[0].Id)

	linkedByNotExists, err3 := dk.GetOffersToBuyOfDymName(ctx, "non-exists")
	require.NoError(t, err3)
	require.Len(t, linkedByNotExists, 0)

	t.Run("should be linked regardless of expired Dym-Name", func(t *testing.T) {
		dymName1.ExpireAt = time.Now().UTC().Add(-time.Hour).Unix()
		require.NoError(t, dk.SetDymName(ctx, dymName1))

		linked1, err1 = dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
		require.NoError(t, err1)
		require.Len(t, linked1, 2)
		require.Equal(t, offer11.Id, linked1[0].Id)
		require.Equal(t, offer12.Id, linked1[1].Id)
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_RemoveReverseMappingDymNameToOfferToBuy(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Error(
		t,
		dk.RemoveReverseMappingDymNameToOfferToBuy(ctx, "@", "1"),
		"fail - should reject invalid Dym-Name",
	)
	require.Error(
		t,
		dk.RemoveReverseMappingDymNameToOfferToBuy(ctx, "a", "@"),
		"fail - should reject invalid offer-id",
	)

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const buyer = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))

	offer11 := dymnstypes.OfferToBuy{
		Id:         "11",
		Name:       dymName1.Name,
		Buyer:      buyer,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer11))

	offer12 := dymnstypes.OfferToBuy{
		Id:         "12",
		Name:       dymName1.Name,
		Buyer:      buyer,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer12))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer11.Id),
	)

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer12.Id),
	)

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	offer2 := dymnstypes.OfferToBuy{
		Id:         "2",
		Name:       dymName2.Name,
		Buyer:      buyer,
		OfferPrice: dymnsutils.TestCoin(1),
	}
	require.NoError(t, dk.SetOfferToBuy(ctx, offer2))

	require.NoError(
		t,
		dk.AddReverseMappingDymNameToOfferToBuy(ctx, dymName2.Name, offer2.Id),
	)

	t.Run("no error if remove a record that not linked", func(t *testing.T) {
		linked, _ := dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
		require.Len(t, linked, 2)

		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer2.Id),
		)

		linked, err := dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("no error if element is not in the list", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, "218362184621"),
		)

		linked, err := dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 2, "existing data must be kept")
	})

	t.Run("remove correctly", func(t *testing.T) {
		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer11.Id),
		)

		linked, err := dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Len(t, linked, 1)
		require.Equal(t, offer12.Id, linked[0].Id)

		require.NoError(
			t,
			dk.RemoveReverseMappingDymNameToOfferToBuy(ctx, dymName1.Name, offer12.Id),
		)

		linked, err = dk.GetOffersToBuyOfDymName(ctx, dymName1.Name)
		require.NoError(t, err)
		require.Empty(t, linked)
	})

	linked, err := dk.GetOffersToBuyOfDymName(ctx, dymName2.Name)
	require.NoError(t, err)
	require.Len(t, linked, 1)
}
