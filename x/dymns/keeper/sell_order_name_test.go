package keeper_test

import (
	"fmt"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_CompleteDymNameSellOrder() {
	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()
	const contactEmail = "contact@example.com"

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
		Contact:    contactEmail,
	}

	s.Run("Dym-Name not found", func() {
		s.Require().ErrorContains(
			s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, "non-exists"),
			"Dym-Name: non-exists: not found",
		)
	})

	s.Run("SO not found", func() {
		err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
		s.Require().NoError(err)

		s.Require().ErrorContains(
			s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name),
			fmt.Sprintf("Sell-Order: %s: not found", dymName.Name),
		)
	})

	s.Run("SO not yet completed, no bidder", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
		s.Require().NoError(err)

		so := dymnstypes.SellOrder{
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	s.Run("SO has bidder but not yet completed", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
		s.Require().NoError(err)

		so := dymnstypes.SellOrder{
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(300)),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: buyerA,
				Price:  s.coin(200), // lower than sell price
			},
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name), "Sell-Order has not finished yet")
	})

	s.Run("SO expired without bidder", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
		s.Require().NoError(err)

		so := dymnstypes.SellOrder{
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() - 1,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(300)),
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name), "no bid placed")
	})

	s.Run("SO without sell price, with bid, finished by expiry", func() {
		s.RefreshContext()

		err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
		s.Require().NoError(err)

		so := dymnstypes.SellOrder{
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: buyerA,
				Price:  s.coin(200),
			},
		}
		err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		s.Require().ErrorContains(s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name), "Sell-Order has not finished yet")
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
			name:                  "pass - completed, expired, no sell price",
			expiredSO:             true,
			sellPrice:             0,
			bid:                   200,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 200,
		},
		{
			name:                  "pass - completed, expired, under sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   200,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 200,
		},
		{
			name:                  "pass - completed, expired, equals sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   300,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 300,
		},
		{
			name:                  "pass - completed by sell-price met, not expired",
			expiredSO:             false,
			sellPrice:             300,
			bid:                   300,
			wantErr:               false,
			wantOwnerBalanceLater: ownerOriginalBalance + 300,
		},
		{
			name:                  "fail - expired without bid, no sell price",
			expiredSO:             true,
			sellPrice:             0,
			bid:                   0,
			wantErr:               true,
			wantErrContains:       "no bid placed",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "fail - expired without bid, with sell price",
			expiredSO:             true,
			sellPrice:             300,
			bid:                   0,
			wantErr:               true,
			wantErrContains:       "no bid placed",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "fail - not expired but bid under sell price",
			expiredSO:             false,
			sellPrice:             300,
			bid:                   200,
			wantErr:               true,
			wantErrContains:       "Sell-Order has not finished yet",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
		{
			name:                  "fail - not expired has bid, no sell price",
			expiredSO:             false,
			sellPrice:             0,
			bid:                   200,
			wantErr:               true,
			wantErrContains:       "Sell-Order has not finished yet",
			wantOwnerBalanceLater: ownerOriginalBalance,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			// setup execution context
			s.RefreshContext()

			s.mintToAccount(ownerA, ownerOriginalBalance)
			s.mintToAccount(buyerA, buyerOriginalBalance)

			dymName.Configs = []dymnstypes.DymNameConfig{
				{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: ownerA,
				},
			}
			s.setDymNameWithFunctionsAfter(dymName)

			so := dymnstypes.SellOrder{
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				MinPrice:  s.coin(100),
			}

			if tt.expiredSO {
				so.ExpireAt = s.now.Unix() - 1
			} else {
				so.ExpireAt = s.now.Unix() + 1
			}

			s.Require().GreaterOrEqual(tt.sellPrice, int64(0), "bad setup")
			so.SellPrice = uptr.To(s.coin(tt.sellPrice))

			s.Require().GreaterOrEqual(tt.bid, int64(0), "bad setup")
			if tt.bid > 0 {
				so.HighestBid = &dymnstypes.SellOrderBid{
					Bidder: buyerA,
					Price:  s.coin(tt.bid),
				}

				// mint coin to module account because we charged buyer before update SO
				err := s.bankKeeper.MintCoins(s.ctx, dymnstypes.ModuleName, sdk.NewCoins(so.HighestBid.Price))
				s.Require().NoError(err)
			}
			err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
			s.Require().NoError(err)

			// test

			errCompleteSellOrder := s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name)
			laterDymName := s.dymNsKeeper.GetDymName(s.ctx, dymName.Name)
			s.Require().NotNil(laterDymName)
			laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, dymName.Name, dymnstypes.TypeName)
			laterOwnerBalance := s.balance(ownerA)
			laterBuyerBalance := s.balance(buyerA)
			laterDymNamesOwnedByOwner, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
			s.Require().NoError(err)
			laterDymNamesOwnedByBuyer, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, buyerA)
			s.Require().NoError(err)
			laterConfiguredAddressOwnerDymNames, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, ownerA)
			s.Require().NoError(err)
			laterConfiguredAddressBuyerDymNames, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, buyerA)
			s.Require().NoError(err)
			laterFallbackAddressOwnerDymNames, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, sdk.MustAccAddressFromBech32(ownerA).Bytes())
			s.Require().NoError(err)
			laterFallbackAddressBuyerDymNames, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, sdk.MustAccAddressFromBech32(buyerA).Bytes())
			s.Require().NoError(err)

			s.Require().Equal(dymName.Name, laterDymName.Name, "name should not be changed")
			s.Require().Equal(dymName.ExpireAt, laterDymName.ExpireAt, "expiry should not be changed")

			if tt.wantErr {
				s.Require().Error(errCompleteSellOrder, "action should be failed")
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Contains(errCompleteSellOrder.Error(), tt.wantErrContains)

				s.Require().NotNil(laterSo, "SO should not be deleted")

				s.Require().Equal(ownerA, laterDymName.Owner, "ownership should not be changed")
				s.Require().Equal(ownerA, laterDymName.Controller, "controller should not be changed")
				s.Require().NotEmpty(laterDymName.Configs, "configs should be kept")
				s.Require().Equal(dymName.Configs, laterDymName.Configs, "configs not be changed")
				s.Require().Equal(contactEmail, dymName.Contact, "contact should not be changed")
				s.Require().Len(laterDymNamesOwnedByOwner, 1, "reverse record should be kept")
				s.Require().Empty(laterDymNamesOwnedByBuyer, "reverse record should not be added")
				s.Require().Len(laterConfiguredAddressOwnerDymNames, 1, "reverse record should be kept")
				s.Require().Empty(laterConfiguredAddressBuyerDymNames, "reverse record should not be added")
				s.Require().Len(laterFallbackAddressOwnerDymNames, 1, "reverse record should be kept")
				s.Require().Empty(laterFallbackAddressBuyerDymNames, "reverse record should not be added")

				s.Require().Equal(ownerOriginalBalance, laterOwnerBalance, "owner balance should not be changed")
				s.Require().Equal(tt.wantOwnerBalanceLater, laterOwnerBalance, "owner balance mis-match")
				s.Require().Equal(buyerOriginalBalance, laterBuyerBalance, "buyer balance should not be changed")
				return
			}

			s.Require().NoError(errCompleteSellOrder, "action should be successful")

			s.Require().Nil(laterSo, "SO should be deleted")

			s.Require().Equal(buyerA, laterDymName.Owner, "ownership should be changed")
			s.Require().Equal(buyerA, laterDymName.Controller, "controller should be changed")
			s.Require().Empty(laterDymName.Configs, "configs should be cleared")
			s.Require().Empty(laterDymName.Contact, "contact should be cleared")
			s.Require().Empty(laterDymNamesOwnedByOwner, "reverse record should be removed")
			s.Require().Len(laterDymNamesOwnedByBuyer, 1, "reverse record should be added")
			s.Require().Empty(laterConfiguredAddressOwnerDymNames, "reverse record should be removed")
			s.Require().Len(laterConfiguredAddressBuyerDymNames, 1, "reverse record should be added")
			s.Require().Empty(laterFallbackAddressOwnerDymNames, "reverse record should be removed")
			s.Require().Len(laterFallbackAddressBuyerDymNames, 1, "reverse record should be added")

			s.Require().Equal(tt.wantOwnerBalanceLater, laterOwnerBalance, "owner balance mis-match")
			s.Require().Equal(buyerOriginalBalance, laterBuyerBalance, "buyer balance should not be changed")
		})
	}

	s.Run("if buyer is owner, can still process", func() {
		s.RefreshContext()

		const ownerOriginalBalance = 100
		const moduleAccountOriginalBalance = 1000
		const offerValue = 300

		s.mintToModuleAccount(moduleAccountOriginalBalance)
		s.mintToAccount(ownerA, ownerOriginalBalance)

		dymName := dymnstypes.DymName{
			Name:       "a",
			Owner:      ownerA,
			Controller: testAddr(3).bech32(),
			ExpireAt:   s.now.Unix() + 100,
			Contact:    contactEmail,
		}
		s.Require().NotEqual(ownerA, dymName.Controller, "bad setup")
		s.setDymNameWithFunctionsAfter(dymName)

		so := dymnstypes.SellOrder{
			AssetId:   dymName.Name,
			AssetType: dymnstypes.TypeName,
			ExpireAt:  s.now.Unix() + 100,
			MinPrice:  s.coin(100),
			SellPrice: uptr.To(s.coin(offerValue)),
			HighestBid: &dymnstypes.SellOrderBid{
				Bidder: ownerA,
				Price:  s.coin(offerValue),
			},
		}

		err := s.dymNsKeeper.SetSellOrder(s.ctx, so)
		s.Require().NoError(err)

		err = s.dymNsKeeper.CompleteDymNameSellOrder(s.ctx, dymName.Name)
		s.Require().NoError(err)

		// Dym-Name should be updated as normal
		laterDymName := s.dymNsKeeper.GetDymName(s.ctx, dymName.Name)
		s.Require().NotNil(laterDymName)
		s.Require().Equal(ownerA, laterDymName.Owner, "ownership should be kept because buyer and owner is the same")
		s.Require().Equal(ownerA, laterDymName.Controller, "controller should be changed to owner as standard")
		s.Require().Empty(laterDymName.Configs, "configs should be cleared")
		s.Require().Empty(laterDymName.Contact, "contact should be cleared")

		// owner receives the offer amount because owner also the buyer
		laterOwnerBalance := s.balance(ownerA)
		s.Require().Equal(int64(offerValue+ownerOriginalBalance), laterOwnerBalance)

		// reverse records should be kept
		laterDymNamesOwnedByOwner, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
		s.Require().NoError(err)
		s.Require().Len(laterDymNamesOwnedByOwner, 1)
		s.Require().Equal(dymName.Name, laterDymNamesOwnedByOwner[0].Name)

		laterConfiguredAddressOwnerDymNames, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, ownerA)
		s.Require().NoError(err)
		s.Require().Len(laterConfiguredAddressOwnerDymNames, 1)
		s.Require().Equal(dymName.Name, laterConfiguredAddressOwnerDymNames[0].Name)

		laterFallbackAddressOwnerDymNames, err := s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, sdk.MustAccAddressFromBech32(ownerA).Bytes())
		s.Require().NoError(err)
		s.Require().Len(laterFallbackAddressOwnerDymNames, 1)
		s.Require().Equal(dymName.Name, laterFallbackAddressOwnerDymNames[0].Name)

		// SO records should be processed as normal
		laterSo := s.dymNsKeeper.GetSellOrder(s.ctx, dymName.Name, dymnstypes.TypeName)
		s.Require().Nil(laterSo, "SO should be deleted")
	})
}
