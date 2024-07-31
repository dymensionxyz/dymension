package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetSetDeleteSellOrder(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	t.Run("reject invalid SO", func(t *testing.T) {
		err := dk.SetSellOrder(ctx, dymnstypes.SellOrder{})
		require.Error(t, err)
	})

	dymName1 := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   1,
	}
	err := dk.SetDymName(ctx, dymName1)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "name2",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   1,
	}
	err = dk.SetDymName(ctx, dymName2)
	require.NoError(t, err)

	so1 := dymnstypes.SellOrder{
		Name:      dymName1.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
		HighestBid: &dymnstypes.SellOrderBid{
			Bidder: "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
			Price:  dymnsutils.TestCoin(200),
		},
	}
	err = dk.SetSellOrder(ctx, so1)
	require.NoError(t, err)
	t.Run("so1 should be equals to original", func(t *testing.T) {
		require.Equal(t, so1, *dk.GetSellOrder(ctx, so1.Name))
	})
	t.Run("SO list should have length 1", func(t *testing.T) {
		require.Len(t, dk.GetAllSellOrders(ctx), 1)
	})
	t.Run("event should be fired on set sell order", func(t *testing.T) {
		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

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
				actionName, dymnstypes.AttributeKeyDymNameSoActionNameSet,
				"event attr action name should be `%s`", dymnstypes.AttributeKeyDymNameSoActionNameSet,
			)
			return
		}

		t.Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
	})

	so2 := dymnstypes.SellOrder{
		Name:     dymName2.Name,
		ExpireAt: 1,
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so2)
	require.NoError(t, err)
	t.Run("so2 should be equals to original", func(t *testing.T) {
		require.Equal(t, so2, *dk.GetSellOrder(ctx, so2.Name))
	})
	t.Run("SO list should have length 2", func(t *testing.T) {
		require.Len(t, dk.GetAllSellOrders(ctx), 2)
	})

	dk.DeleteSellOrder(ctx, so1.Name)
	t.Run("event should be fired on delete sell order", func(t *testing.T) {
		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

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
				actionName, dymnstypes.AttributeKeyDymNameSoActionNameSet,
				"event attr action name should be `%s`", dymnstypes.AttributeKeyDymNameSoActionNameDelete,
			)
			return
		}

		t.Errorf("event %s not found", dymnstypes.EventTypeSellOrder)
	})

	t.Run("so1 should be nil", func(t *testing.T) {
		require.Nil(t, dk.GetSellOrder(ctx, so1.Name))
	})
	t.Run("SO list should have length 1", func(t *testing.T) {
		list := dk.GetAllSellOrders(ctx)
		require.Len(t, list, 1)
		require.Equal(t, so2.Name, list[0].Name)
	})

	t.Run("non-exists returns nil", func(t *testing.T) {
		require.Nil(t, dk.GetSellOrder(ctx, "non-exists"))
	})

	t.Run("omit Sell Price if not nil but zero", func(t *testing.T) {
		so3 := dymnstypes.SellOrder{
			Name:      "hello",
			ExpireAt:  1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(0),
		}
		err = dk.SetSellOrder(ctx, so3)
		require.NoError(t, err)

		require.Nil(t, dk.GetSellOrder(ctx, so3.Name).SellPrice)
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_MoveSellOrderToHistorical(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	// setting block time
	ctx = ctx.WithBlockHeader(tmproto.Header{
		Time: time.Now().UTC(),
	})

	futureEpoch := ctx.BlockTime().Add(time.Hour).Unix()

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName1 := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   futureEpoch,
	}
	err := dk.SetDymName(ctx, dymName1)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "owned-by-1",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   futureEpoch,
	}
	err = dk.SetDymName(ctx, dymName2)
	require.NoError(t, err)

	dymNames := dk.GetAllNonExpiredDymNames(ctx, time.Now().Unix())
	require.Len(t, dymNames, 2)

	so11 := dymnstypes.SellOrder{
		Name:      dymName1.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so11)
	require.NoError(t, err)

	t.Run("should able to move", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so11.Name)
		require.NoError(t, err)
	})

	t.Run("moved SO should be removed from active", func(t *testing.T) {
		require.Nil(t, dk.GetSellOrder(ctx, so11.Name))
	})

	t.Run("has min expiry mapping", func(t *testing.T) {
		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, so11.Name)
		require.True(t, found)
		require.Equal(t, so11.ExpireAt, minExpiry)
	})

	t.Run("should not move non-exists", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, "non-exists")
		require.Error(t, err)
		require.Contains(t, err.Error(), dymnstypes.ErrSellOrderNotFound.Error())
	})

	t.Run("should able to move a duplicated without error", func(t *testing.T) {
		err = dk.SetSellOrder(ctx, so11)
		require.NoError(t, err)

		defer func() {
			dk.DeleteSellOrder(ctx, so11.Name)
		}()

		err = dk.MoveSellOrderToHistorical(ctx, so11.Name)
		require.NoError(t, err)

		list := dk.GetHistoricalSellOrders(ctx, so11.Name)
		require.Len(t, list, 1, "do not persist duplicated historical SO")
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Empty(t, dk.GetSellOrder(ctx, dymName2.Name))
	})

	so2 := dymnstypes.SellOrder{
		Name:     dymName2.Name,
		ExpireAt: 1,
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so2)
	require.NoError(t, err)

	t.Run("should able to move", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so2.Name)
		require.NoError(t, err)
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name), 1)
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name), 1)
	})

	so12 := dymnstypes.SellOrder{
		Name:      dymName1.Name,
		ExpireAt:  futureEpoch,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so12)
	require.NoError(t, err)
	t.Run("should not move yet finished SO", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so12.Name)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has not expired yet")
	})

	so12.HighestBid = &dymnstypes.SellOrderBid{
		Bidder: "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
		Price:  dymnsutils.TestCoin(300),
	}
	err = dk.SetSellOrder(ctx, so12)
	require.NoError(t, err)

	t.Run("should able to move finished SO", func(t *testing.T) {
		err := dk.MoveSellOrderToHistorical(ctx, so12.Name)
		require.NoError(t, err)

		list := dk.GetHistoricalSellOrders(ctx, so12.Name)
		require.Len(t, list, 2, "should appended to historical")

		minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, so12.Name)
		require.True(t, found)
		require.Equal(t, so11.ExpireAt, minExpiry, "should keep the minimum")
		require.NotEqual(t, so12.ExpireAt, minExpiry, "should keep the minimum")
	})

	t.Run("other records remaining as-is", func(t *testing.T) {
		require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name), 1)
	})
}

func TestKeeper_GetAndDeleteHistoricalSellOrders(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	now := time.Now().UTC()
	futureEpoch := now.Unix() + 1

	// setting block time
	ctx = ctx.WithBlockHeader(tmproto.Header{
		Time: now,
	})

	//goland:noinspection SpellCheckingInspection
	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName1 := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   futureEpoch,
	}
	err := dk.SetDymName(ctx, dymName1)
	require.NoError(t, err)

	dymName2 := dymnstypes.DymName{
		Name:       "owned-by-1",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   futureEpoch,
	}
	err = dk.SetDymName(ctx, dymName2)
	require.NoError(t, err)

	t.Run("getting non-exists should returns empty", func(t *testing.T) {
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name))
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name))
	})

	so11 := dymnstypes.SellOrder{
		Name:      dymName1.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, so11)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so11.Name)
	require.NoError(t, err)

	so2 := dymnstypes.SellOrder{
		Name:     dymName2.Name,
		ExpireAt: 1,
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so2)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so2.Name)
	require.NoError(t, err)

	so2.ExpireAt++
	err = dk.SetSellOrder(ctx, so2)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, so2.Name)
	require.NoError(t, err)

	t.Run("fetch correctly", func(t *testing.T) {
		list1 := dk.GetHistoricalSellOrders(ctx, dymName1.Name)
		require.Len(t, list1, 1)
		list2 := dk.GetHistoricalSellOrders(ctx, dymName2.Name)
		require.Len(t, list2, 2)
		require.Equal(t, so2.Name, list2[0].Name)
		require.Equal(t, so2.Name, list2[1].Name)
		require.Equal(t, int64(1), list2[0].ExpireAt)
		require.Equal(t, int64(2), list2[1].ExpireAt)
	})

	t.Run("delete", func(t *testing.T) {
		dk.DeleteHistoricalSellOrders(ctx, dymName1.Name)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name))

		list2 := dk.GetHistoricalSellOrders(ctx, dymName2.Name)
		require.Len(t, list2, 2)

		dk.DeleteHistoricalSellOrders(ctx, dymName2.Name)
		require.Empty(t, dk.GetHistoricalSellOrders(ctx, dymName2.Name))
	})
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_CompleteSellOrder(t *testing.T) {
	now := time.Now().UTC()
	futureEpoch := now.Unix() + 1

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockHeader(tmproto.Header{
			Time: now,
		})

		return dk, bk, ctx
	}

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const buyer = "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	const contactEmail = "contact@example.com"

	originalDymNameExpiry := futureEpoch
	dymName := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   originalDymNameExpiry,
		Contact:    contactEmail,
	}

	t.Run("Dym-Name not found", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorContains(t, dk.CompleteSellOrder(ctx, "non-exists"), dymnstypes.ErrDymNameNotFound.Error())
	})

	t.Run("SO not found", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteSellOrder(ctx, dymName.Name), dymnstypes.ErrSellOrderNotFound.Error())
	})

	t.Run("SO not yet completed, no bidder", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			Name:     dymName.Name,
			ExpireAt: futureEpoch,
			MinPrice: dymnsutils.TestCoin(100),
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteSellOrder(ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	t.Run("SO has bidder but not yet completed", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			Name:      dymName.Name,
			ExpireAt:  futureEpoch,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: buyer,
				Price:  dymnsutils.TestCoin(200), // lower than sell price
			},
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteSellOrder(ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	t.Run("SO expired without bidder", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			Name:      dymName.Name,
			ExpireAt:  now.Unix() - 1,
			MinPrice:  dymnsutils.TestCoin(100),
			SellPrice: dymnsutils.TestCoinP(300),
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteSellOrder(ctx, dymName.Name), "no bid placed")
	})

	t.Run("SO without sell price, with bid, finished by expiry", func(t *testing.T) {
		dk, _, ctx := setupTest()

		err := dk.SetDymName(ctx, dymName)
		require.NoError(t, err)

		so := dymnstypes.SellOrder{
			Name:     dymName.Name,
			ExpireAt: futureEpoch,
			MinPrice: dymnsutils.TestCoin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: buyer,
				Price:  dymnsutils.TestCoin(200),
			},
		}
		err = dk.SetSellOrder(ctx, so)
		require.NoError(t, err)

		requireErrorContains(t, dk.CompleteSellOrder(ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	const ownerOriginalBalance int64 = 1000
	const buyerOriginalBalance int64 = 500
	tests := []struct {
		name                  string
		expiredSO             bool
		sellPrice             int64
		bid                   int64
		wantErr               bool
		wantErrContains       string
		wantOwnerBalanceLater int64
	}{
		{
			name:                  "completed, expired, no sell price",
			expiredSO:             true,
			sellPrice:             0,
			bid:                   200,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 200,
		},
		{
			name:                  "completed, expired, under sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   200,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 200,
		},
		{
			name:                  "completed, expired, equals sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   300,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 300,
		},
		{
			name:                  "completed by sell-price met, not expired",
			expiredSO:             false,
			sellPrice:             300,
			bid:                   300,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 300,
		},
		{
			name:                  "expired without bid, no sell price",
			expiredSO:             true,
			sellPrice:             0,
			bid:                   0,
			wantErr:               true,
			wantErrContains:       "no bid placed",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "expired without bid, with sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   0,
			wantErr:               true,
			wantErrContains:       "no bid placed",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "not expired but bid under sell price",
			expiredSO:             false,
			sellPrice:             300,
			bid:                   200,
			wantErr:               true,
			wantErrContains:       "Sell-Order has not finished yet",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "not expired has bid, no sell price",
			expiredSO:             false,
			sellPrice:             0,
			bid:                   200,
			wantErr:               true,
			wantErrContains:       "Sell-Order has not finished yet",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup execution context
			dk, bk, ctx := setupTest()

			err := bk.MintCoins(ctx,
				dymnstypes.ModuleName,
				dymnsutils.TestCoins(ownerOriginalBalance+buyerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(owner),
				dymnsutils.TestCoins(ownerOriginalBalance),
			)
			require.NoError(t, err)
			err = bk.SendCoinsFromModuleToAccount(ctx,
				dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(buyer),
				dymnsutils.TestCoins(buyerOriginalBalance),
			)
			require.NoError(t, err)

			dymName.Configs = []dymnstypes.DymNameConfig{
				{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: owner,
				},
			}
			setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

			so := dymnstypes.SellOrder{
				Name:     dymName.Name,
				MinPrice: dymnsutils.TestCoin(100),
			}

			if tt.expiredSO {
				so.ExpireAt = now.Unix() - 1
			} else {
				so.ExpireAt = futureEpoch
			}

			require.GreaterOrEqual(t, tt.sellPrice, int64(0), "bad setup")
			so.SellPrice = dymnsutils.TestCoinP(tt.sellPrice)

			require.GreaterOrEqual(t, tt.bid, int64(0), "bad setup")
			if tt.bid > 0 {
				so.HighestBid = &dymnstypes.SellOrderBid{
					Bidder: buyer,
					Price:  dymnsutils.TestCoin(tt.bid),
				}

				// mint coin to module account because we charged buyer before update SO
				err = bk.MintCoins(ctx, dymnstypes.ModuleName, sdk.NewCoins(so.HighestBid.Price))
				require.NoError(t, err)
			}
			err = dk.SetSellOrder(ctx, so)
			require.NoError(t, err)

			// test

			errCompleteSellOrder := dk.CompleteSellOrder(ctx, dymName.Name)
			laterDymName := dk.GetDymName(ctx, dymName.Name)
			require.NotNil(t, laterDymName)
			laterSo := dk.GetSellOrder(ctx, dymName.Name)
			historicalSo := dk.GetHistoricalSellOrders(ctx, dymName.Name)
			laterOwnerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(owner), params.BaseDenom)
			laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(buyer), params.BaseDenom)
			laterDymNamesOwnedByOwner, err := dk.GetDymNamesOwnedBy(ctx, owner, now.Unix())
			require.NoError(t, err)
			laterDymNamesOwnedByBuyer, err := dk.GetDymNamesOwnedBy(ctx, buyer, now.Unix())
			require.NoError(t, err)
			laterConfiguredAddressOwnerDymNames, err := dk.GetDymNamesContainsConfiguredAddress(ctx, owner, now.Unix())
			require.NoError(t, err)
			laterConfiguredAddressBuyerDymNames, err := dk.GetDymNamesContainsConfiguredAddress(ctx, buyer, now.Unix())
			require.NoError(t, err)
			laterHexAddressOwnerDymNames, err := dk.GetDymNamesContainsHexAddress(ctx, sdk.MustAccAddressFromBech32(owner), now.Unix())
			require.NoError(t, err)
			laterHexAddressBuyerDymNames, err := dk.GetDymNamesContainsHexAddress(ctx, sdk.MustAccAddressFromBech32(buyer), now.Unix())
			require.NoError(t, err)

			require.Equal(t, dymName.Name, laterDymName.Name, "name should not be changed")
			require.Equal(t, originalDymNameExpiry, laterDymName.ExpireAt, "expiry should not be changed")

			if tt.wantErr {
				require.Error(t, errCompleteSellOrder, "action should be failed")
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Contains(t, errCompleteSellOrder.Error(), tt.wantErrContains)

				require.NotNil(t, laterSo, "SO should not be deleted")
				require.Empty(t, historicalSo, "SO should not be moved to historical")

				require.Equal(t, owner, laterDymName.Owner, "ownership should not be changed")
				require.Equal(t, owner, laterDymName.Controller, "controller should not be changed")
				require.NotEmpty(t, laterDymName.Configs, "configs should be kept")
				require.Equal(t, dymName.Configs, laterDymName.Configs, "configs not be changed")
				require.Equal(t, contactEmail, dymName.Contact, "contact should not be changed")
				require.Len(t, laterDymNamesOwnedByOwner, 1, "reverse record should be kept")
				require.Empty(t, laterDymNamesOwnedByBuyer, "reverse record should not be added")
				require.Len(t, laterConfiguredAddressOwnerDymNames, 1, "reverse record should be kept")
				require.Empty(t, laterConfiguredAddressBuyerDymNames, "reverse record should not be added")
				require.Len(t, laterHexAddressOwnerDymNames, 1, "reverse record should be kept")
				require.Empty(t, laterHexAddressBuyerDymNames, "reverse record should not be added")

				require.Equal(t, ownerOriginalBalance, laterOwnerBalance.Amount.Int64(), "owner balance should not be changed")
				require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
				require.Equal(t, buyerOriginalBalance, laterBuyerBalance.Amount.Int64(), "buyer balance should not be changed")
				return
			}

			require.NoError(t, errCompleteSellOrder, "action should be successful")

			require.Nil(t, laterSo, "SO should be deleted")
			require.Len(t, historicalSo, 1, "SO should be moved to historical")

			require.Equal(t, buyer, laterDymName.Owner, "ownership should be changed")
			require.Equal(t, buyer, laterDymName.Controller, "controller should be changed")
			require.Empty(t, laterDymName.Configs, "configs should be cleared")
			require.Empty(t, laterDymName.Contact, "contact should be cleared")
			require.Empty(t, laterDymNamesOwnedByOwner, "reverse record should be removed")
			require.Len(t, laterDymNamesOwnedByBuyer, 1, "reverse record should be added")
			require.Empty(t, laterConfiguredAddressOwnerDymNames, "reverse record should be removed")
			require.Len(t, laterConfiguredAddressBuyerDymNames, 1, "reverse record should be added")
			require.Empty(t, laterHexAddressOwnerDymNames, "reverse record should be removed")
			require.Len(t, laterHexAddressBuyerDymNames, 1, "reverse record should be added")

			require.Equal(t, tt.wantOwnerBalanceLater, laterOwnerBalance.Amount.Int64(), "owner balance mis-match")
			require.Equal(t, buyerOriginalBalance, laterBuyerBalance.Amount.Int64(), "buyer balance should not be changed")
		})
	}
}

func TestKeeper_GetSetActiveSellOrdersExpiration(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	t.Run("get", func(t *testing.T) {
		aope := dk.GetActiveSellOrdersExpiration(ctx)
		require.Empty(t, aope.Records, "default list must be empty")
		require.NotNil(t, aope.Records, "list must be initialized")
	})

	t.Run("set", func(t *testing.T) {
		aope := &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     "hello",
					ExpireAt: 123,
				},
				{
					Name:     "world",
					ExpireAt: 456,
				},
			},
		}
		err := dk.SetActiveSellOrdersExpiration(ctx, aope)
		require.NoError(t, err)

		aope = dk.GetActiveSellOrdersExpiration(ctx)
		require.Len(t, aope.Records, 2)
		require.Equal(t, "hello", aope.Records[0].Name)
		require.Equal(t, int64(123), aope.Records[0].ExpireAt)
		require.Equal(t, "world", aope.Records[1].Name)
		require.Equal(t, int64(456), aope.Records[1].ExpireAt)
	})

	t.Run("must automatically sort when set", func(t *testing.T) {
		err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     "b",
					ExpireAt: 456,
				},
				{
					Name:     "a",
					ExpireAt: 123,
				},
			},
		})
		require.NoError(t, err)

		aope := dk.GetActiveSellOrdersExpiration(ctx)
		require.Len(t, aope.Records, 2)

		require.Equal(t, "a", aope.Records[0].Name)
		require.Equal(t, int64(123), aope.Records[0].ExpireAt)
		require.Equal(t, "b", aope.Records[1].Name)
		require.Equal(t, int64(456), aope.Records[1].ExpireAt)
	})

	t.Run("can not set if set is not valid", func(t *testing.T) {
		// not unique
		err := dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     "a",
					ExpireAt: 456,
				},
				{
					Name:     "a",
					ExpireAt: 123,
				},
			},
		})
		require.Error(t, err)

		// zero expiry
		err = dk.SetActiveSellOrdersExpiration(ctx, &dymnstypes.ActiveSellOrdersExpiration{
			Records: []dymnstypes.ActiveSellOrdersExpirationRecord{
				{
					Name:     "a",
					ExpireAt: -1,
				},
				{
					Name:     "b",
					ExpireAt: 0,
				},
			},
		})
		require.Error(t, err)
	})
}

func TestKeeper_GetSetMinExpiryHistoricalSellOrder(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	dk.SetMinExpiryHistoricalSellOrder(ctx, "hello", 123)
	dk.SetMinExpiryHistoricalSellOrder(ctx, "world", 456)

	min, found := dk.GetMinExpiryHistoricalSellOrder(ctx, "hello")
	require.True(t, found)
	require.Equal(t, int64(123), min)

	min, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "world")
	require.True(t, found)
	require.Equal(t, int64(456), min)

	min, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "non-exists")
	require.False(t, found)
	require.Zero(t, min)

	t.Run("set zero means delete", func(t *testing.T) {
		dk.SetMinExpiryHistoricalSellOrder(ctx, "hello", 0)

		min, found = dk.GetMinExpiryHistoricalSellOrder(ctx, "hello")
		require.False(t, found)
		require.Zero(t, min)
	})
}

func TestKeeper_GetMinExpiryOfAllHistoricalSellOrders(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	dk.SetMinExpiryHistoricalSellOrder(ctx, "one", 1)
	dk.SetMinExpiryHistoricalSellOrder(ctx, "two", 22)
	dk.SetMinExpiryHistoricalSellOrder(ctx, "three", 333)

	records := dk.GetMinExpiryOfAllHistoricalSellOrders(ctx)
	require.Len(t, records, 3)
	require.Equal(t, []dymnstypes.HistoricalSellOrderMinExpiry{
		{
			DymName:   "one",
			MinExpiry: 1,
		},
		{
			DymName:   "three",
			MinExpiry: 333,
		},
		{
			DymName:   "two",
			MinExpiry: 22,
		},
	}, records)

	dk.SetMinExpiryHistoricalSellOrder(ctx, "three", 0)
	records = dk.GetMinExpiryOfAllHistoricalSellOrders(ctx)
	require.Len(t, records, 2)
	require.Equal(t, []dymnstypes.HistoricalSellOrderMinExpiry{
		{
			DymName:   "one",
			MinExpiry: 1,
		},
		{
			DymName:   "two",
			MinExpiry: 22,
		},
	}, records)

	t.Run("result must be sorted by Dym-Name", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		dk.SetMinExpiryHistoricalSellOrder(ctx, "a", 1)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "c", 2)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "b", 3)
		dk.SetMinExpiryHistoricalSellOrder(ctx, "d", 4)

		records := dk.GetMinExpiryOfAllHistoricalSellOrders(ctx)
		require.Equal(t, []dymnstypes.HistoricalSellOrderMinExpiry{
			{
				DymName:   "a",
				MinExpiry: 1,
			},
			{
				DymName:   "b",
				MinExpiry: 3,
			},
			{
				DymName:   "c",
				MinExpiry: 2,
			},
			{
				DymName:   "d",
				MinExpiry: 4,
			},
		}, records)
	})
}
