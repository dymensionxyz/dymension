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

func TestKeeper_IncreaseOfferToBuyCountAndGet(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	require.Zero(t, dk.GetCountOfferToBuy(ctx))

	count := dk.IncreaseOfferToBuyCountAndGet(ctx)
	require.Equal(t, uint64(1), count)
	require.Equal(t, uint64(1), dk.GetCountOfferToBuy(ctx))

	count = dk.IncreaseOfferToBuyCountAndGet(ctx)
	require.Equal(t, uint64(2), count)
	require.Equal(t, uint64(2), dk.GetCountOfferToBuy(ctx))

	count = dk.IncreaseOfferToBuyCountAndGet(ctx)
	require.Equal(t, uint64(3), count)
	require.Equal(t, uint64(3), dk.GetCountOfferToBuy(ctx))

	dk.SetCountOfferToBuy(ctx, math.MaxUint64-1)

	count = dk.IncreaseOfferToBuyCountAndGet(ctx)
	require.Equal(t, uint64(math.MaxUint64), count)
	require.Equal(t, uint64(math.MaxUint64), dk.GetCountOfferToBuy(ctx))

	require.Panics(t, func() {
		dk.IncreaseOfferToBuyCountAndGet(ctx)
	}, "expect panic on overflow when increasing count of Offer-To-Buy greater than uint64")
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetSetInsertOfferToBuy(t *testing.T) {
	t.Run("get non-exists offer should returns nil", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer := dk.GetOfferToBuy(ctx, "non-exists")
		require.Nil(t, offer)

		offer = dk.GetOfferToBuy(ctx, "183418")
		require.Nil(t, offer)
	})

	t.Run("should returns error when set empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		err := dk.SetOfferToBuy(ctx, dymnstypes.OfferToBuy{
			Id:         "",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		})
		require.Error(t, err)
	})

	t.Run("can set and can get", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer := dymnstypes.OfferToBuy{
			Id:         "1",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetOfferToBuy(ctx, offer)
		require.NoError(t, err)

		offerGot := dk.GetOfferToBuy(ctx, "1")
		require.NotNil(t, offerGot)

		require.Equal(t, offer, *offerGot)
	})

	t.Run("should panic when insert non-empty ID offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.Panics(t, func() {
			_, _ = dk.InsertOfferToBuy(ctx, dymnstypes.OfferToBuy{
				Id:         "1",
				Name:       "a",
				Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				OfferPrice: dymnsutils.TestCoin(1),
			})
		})
	})

	t.Run("can insert", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1 := dymnstypes.OfferToBuy{
			Id:         "",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer, err := dk.InsertOfferToBuy(ctx, offer1)
		require.NoError(t, err)
		require.Equal(t, "1", offer.Id)

		offerGot := dk.GetOfferToBuy(ctx, "1")
		require.NotNil(t, offerGot)

		offer1.Id = "1"
		require.Equal(t, offer1, *offerGot)
	})

	t.Run("can not insert duplicated", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetCountOfferToBuy(ctx, 1)
		nextId := "2"

		existing := dymnstypes.OfferToBuy{
			Id:         nextId,
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetOfferToBuy(ctx, existing)
		require.NoError(t, err)

		offer := dymnstypes.OfferToBuy{
			Id:         "",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}

		_, err = dk.InsertOfferToBuy(ctx, offer)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Offer-To-Buy with ID 2 already exists")
	})

	t.Run("should automatically fill ID when insert", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer1 := dymnstypes.OfferToBuy{
			Id:         "",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}

		offer, err := dk.InsertOfferToBuy(ctx, offer1)
		require.NoError(t, err)
		require.Equal(t, "1", offer.Id)

		offerGot := dk.GetOfferToBuy(ctx, "1")
		require.NotNil(t, offerGot)

		offer1.Id = "1"
		require.Equal(t, offer1, *offerGot)

		dk.SetCountOfferToBuy(ctx, 99)

		offer2 := dymnstypes.OfferToBuy{
			Id:         "",
			Name:       "b",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}
		offer, err = dk.InsertOfferToBuy(ctx, offer2)
		require.NoError(t, err)
		require.Equal(t, "100", offer.Id)

		offerGot = dk.GetOfferToBuy(ctx, "100")
		require.NotNil(t, offerGot)

		offer2.Id = "100"
		require.Equal(t, offer2, *offerGot)
	})

	t.Run("can delete", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		var err error

		offer1 := dymnstypes.OfferToBuy{
			Id:         "1",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}
		err = dk.SetOfferToBuy(ctx, offer1)
		require.NoError(t, err)

		offer2 := dymnstypes.OfferToBuy{
			Id:         "2",
			Name:       "b",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(2),
		}
		err = dk.SetOfferToBuy(ctx, offer2)
		require.NoError(t, err)

		offer3 := dymnstypes.OfferToBuy{
			Id:         "3",
			Name:       "c",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(3),
		}
		err = dk.SetOfferToBuy(ctx, offer3)
		require.NoError(t, err)

		offer4 := dymnstypes.OfferToBuy{
			Id:         "4",
			Name:       "d",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(4),
		}
		err = dk.SetOfferToBuy(ctx, offer4)
		require.NoError(t, err)

		require.NotNil(t, dk.GetOfferToBuy(ctx, "1"))
		require.NotNil(t, dk.GetOfferToBuy(ctx, "2"))
		require.NotNil(t, dk.GetOfferToBuy(ctx, "3"))
		require.NotNil(t, dk.GetOfferToBuy(ctx, "4"))

		dk.DeleteOfferToBuy(ctx, "2")
		require.NotNil(t, dk.GetOfferToBuy(ctx, "1"))
		require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
		require.NotNil(t, dk.GetOfferToBuy(ctx, "3"))
		require.NotNil(t, dk.GetOfferToBuy(ctx, "4"))

		dk.DeleteOfferToBuy(ctx, "4")
		require.NotNil(t, dk.GetOfferToBuy(ctx, "1"))
		require.Nil(t, dk.GetOfferToBuy(ctx, "2"))
		require.NotNil(t, dk.GetOfferToBuy(ctx, "3"))
		require.Nil(t, dk.GetOfferToBuy(ctx, "4"))
	})

	t.Run("delete non-existing will not error", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.DeleteOfferToBuy(ctx, "99999")
	})

	t.Run("event should be fired on set/insert offer", func(t *testing.T) {
		tests := []struct {
			name    string
			offer   dymnstypes.OfferToBuy
			setFunc func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.OfferToBuy)
		}{
			{
				name: "set offer",
				offer: dymnstypes.OfferToBuy{
					Id:         "1",
					Name:       "a",
					Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
					OfferPrice: dymnsutils.TestCoin(1),
				},
				setFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper, offer dymnstypes.OfferToBuy) {
					err := dk.SetOfferToBuy(ctx, offer)
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
					if event.Type != dymnstypes.EventTypeOfferToBuy {
						continue
					}

					var actionName string
					for _, attr := range event.Attributes {
						if attr.Key == dymnstypes.AttributeKeyOtbActionName {
							actionName = attr.Value
						}
					}
					require.NotEmpty(t, actionName, "event attr action name could not be found")
					require.Equalf(t,
						actionName, dymnstypes.AttributeKeyOtbActionNameSet,
						"event attr action name should be `%s`", dymnstypes.AttributeKeyOtbActionNameSet,
					)
					return
				}

				t.Errorf("event %s not found", dymnstypes.EventTypeOfferToBuy)
			})
		}
	})

	t.Run("event should be fired on delete offer", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		offer := dymnstypes.OfferToBuy{
			Id:         "1",
			Name:       "a",
			Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			OfferPrice: dymnsutils.TestCoin(1),
		}

		err := dk.SetOfferToBuy(ctx, offer)
		require.NoError(t, err)

		ctx = ctx.WithEventManager(sdk.NewEventManager())

		dk.DeleteOfferToBuy(ctx, offer.Id)

		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

		for _, event := range events {
			if event.Type != dymnstypes.EventTypeOfferToBuy {
				continue
			}

			var actionName string
			for _, attr := range event.Attributes {
				if attr.Key == dymnstypes.AttributeKeyOtbActionName {
					actionName = attr.Value
				}
			}
			require.NotEmpty(t, actionName, "event attr action name could not be found")
			require.Equalf(t,
				actionName, dymnstypes.AttributeKeyOtbActionNameDelete,
				"event attr action name should be `%s`", dymnstypes.AttributeKeyOtbActionNameDelete,
			)
			return
		}

		t.Errorf("event %s not found", dymnstypes.EventTypeOfferToBuy)
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetAllOffersToBuy(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	offer1 := dymnstypes.OfferToBuy{
		Id:         "1",
		Name:       "a",
		Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err := dk.SetOfferToBuy(ctx, offer1)
	require.NoError(t, err)

	offers := dk.GetAllOffersToBuy(ctx)
	require.Len(t, offers, 1)
	require.Equal(t, offer1, offers[0])

	offer2 := dymnstypes.OfferToBuy{
		Id:         "2",
		Name:       "a",
		Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		OfferPrice: dymnsutils.TestCoin(1),
	}
	err = dk.SetOfferToBuy(ctx, offer2)
	require.NoError(t, err)

	offers = dk.GetAllOffersToBuy(ctx)
	require.Len(t, offers, 2)
	require.Equal(t, []dymnstypes.OfferToBuy{offer1, offer2}, offers)

	offer3 := dymnstypes.OfferToBuy{
		Id:         "3",
		Name:       "b",
		Buyer:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		OfferPrice: dymnsutils.TestCoin(3),
	}
	err = dk.SetOfferToBuy(ctx, offer3)
	require.NoError(t, err)

	offers = dk.GetAllOffersToBuy(ctx)
	require.Len(t, offers, 3)
	require.Equal(t, []dymnstypes.OfferToBuy{offer1, offer2, offer3}, offers)
}
