package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) TestKeeper_RefundBid() {
	bidderA := testAddr(1).bech32()

	tests := []struct {
		name                     string
		refundToAccount          string
		refundAmount             sdk.Coin
		fundModuleAccountBalance sdk.Coin
		genesis                  bool
		wantErr                  bool
		wantErrContains          string
	}{
		{
			name:                     "pass - refund bid",
			refundToAccount:          bidderA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(150),
			genesis:                  false,
		},
		{
			name:                     "pass - refund bid genesis",
			refundToAccount:          bidderA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(0), // no need balance, will mint
			genesis:                  true,
			wantErr:                  false,
		},
		{
			name:                     "fail - refund bid normally but module account has no balance",
			refundToAccount:          bidderA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(0),
			genesis:                  false,
			wantErr:                  true,
			wantErrContains:          "insufficient funds",
		},
		{
			name:                     "fail - refund bid normally but module account does not have enough balance",
			refundToAccount:          bidderA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(50),
			genesis:                  false,
			wantErr:                  true,
			wantErrContains:          "insufficient funds",
		},
		{
			name:                     "fail - bad bidder",
			refundToAccount:          "0x1",
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(100),
			wantErr:                  true,
			wantErrContains:          "SO bidder is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if !tt.fundModuleAccountBalance.IsNil() {
				if !tt.fundModuleAccountBalance.IsZero() {
					err := s.bankKeeper.MintCoins(s.ctx, dymnstypes.ModuleName, sdk.Coins{tt.fundModuleAccountBalance})
					s.Require().NoError(err)
				}
			}

			soBid := dymnstypes.SellOrderBid{
				Bidder: tt.refundToAccount,
				Price:  tt.refundAmount,
			}

			var err error
			if tt.genesis {
				err = s.dymNsKeeper.GenesisRefundBid(s.ctx, soBid)
			} else {
				err = s.dymNsKeeper.RefundBid(s.ctx, soBid, dymnstypes.TypeName)
			}

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)

			laterBidderBalance := s.balance2(tt.refundToAccount)
			s.Require().Equal(tt.refundAmount.Amount.String(), laterBidderBalance.String())

			laterDymNsModuleBalance := s.moduleBalance2()
			if tt.genesis {
				s.Require().True(laterDymNsModuleBalance.IsZero())
			} else {
				s.Require().Equal(
					tt.fundModuleAccountBalance.Sub(tt.refundAmount).Amount.String(),
					laterDymNsModuleBalance.String(),
				)
			}

			// event should be fired
			events := s.ctx.EventManager().Events()
			s.Require().NotEmpty(events)

			var found bool
			for _, event := range events {
				if event.Type == dymnstypes.EventTypeSoRefundBid {
					found = true
					break
				}
			}

			if !found {
				s.T().Errorf("event %s not found", dymnstypes.EventTypeSoRefundBid)
			}
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_RefundBuyOrder() {
	buyerA := testAddr(1).bech32()

	supportedAssetTypes := []dymnstypes.AssetType{
		dymnstypes.TypeName, dymnstypes.TypeAlias,
	}

	tests := []struct {
		name                     string
		refundToAccount          string
		refundAmount             sdk.Coin
		fundModuleAccountBalance sdk.Coin
		genesis                  bool
		wantErr                  bool
		wantErrContains          string
	}{
		{
			name:                     "pass - refund offer",
			refundToAccount:          buyerA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(150),
			genesis:                  false,
		},
		{
			name:                     "pass - refund offer genesis",
			refundToAccount:          buyerA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(0), // no need balance, will mint
			genesis:                  true,
			wantErr:                  false,
		},
		{
			name:                     "fail - refund offer normally but module account has no balance",
			refundToAccount:          buyerA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(0),
			genesis:                  false,
			wantErr:                  true,
			wantErrContains:          "insufficient funds",
		},
		{
			name:                     "fail - refund offer normally but module account does not have enough balance",
			refundToAccount:          buyerA,
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(50),
			genesis:                  false,
			wantErr:                  true,
			wantErrContains:          "insufficient funds",
		},
		{
			name:                     "fail - bad offer buyer address",
			refundToAccount:          "0x1",
			refundAmount:             s.coin(100),
			fundModuleAccountBalance: s.coin(100),
			wantErr:                  true,
			wantErrContains:          "buyer is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			for _, assetType := range supportedAssetTypes {
				s.Run(assetType.PrettyName(), func() {
					s.RefreshContext()

					if !tt.fundModuleAccountBalance.IsNil() {
						if !tt.fundModuleAccountBalance.IsZero() {
							err := s.bankKeeper.MintCoins(s.ctx, dymnstypes.ModuleName, sdk.Coins{tt.fundModuleAccountBalance})
							s.Require().NoError(err)
						}
					}

					var orderParams []string
					if assetType == dymnstypes.TypeAlias {
						orderParams = []string{"rollapp_1-1"}
					}

					offer := dymnstypes.BuyOrder{
						Id:         dymnstypes.CreateBuyOrderId(assetType, 1),
						AssetId:    "asset",
						AssetType:  assetType,
						Params:     orderParams,
						Buyer:      tt.refundToAccount,
						OfferPrice: tt.refundAmount,
					}

					var err error
					if tt.genesis {
						err = s.dymNsKeeper.GenesisRefundBuyOrder(s.ctx, offer)
					} else {
						err = s.dymNsKeeper.RefundBuyOrder(s.ctx, offer)
					}

					if tt.wantErr {
						s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
						s.Require().Error(err)
						s.Require().Contains(err.Error(), tt.wantErrContains)
						return
					}

					s.Require().NoError(err)

					laterBidderBalance := s.balance2(tt.refundToAccount)
					s.Require().Equal(tt.refundAmount.Amount.String(), laterBidderBalance.String())

					laterDymNsModuleBalance := s.moduleBalance2()
					if tt.genesis {
						s.Require().True(laterDymNsModuleBalance.IsZero())
					} else {
						s.Require().Equal(
							tt.fundModuleAccountBalance.Sub(tt.refundAmount).Amount.String(),
							laterDymNsModuleBalance.String(),
						)
					}

					// event should be fired
					events := s.ctx.EventManager().Events()
					s.Require().NotEmpty(events)

					var found bool
					for _, event := range events {
						if event.Type == dymnstypes.EventTypeBoRefundOffer {
							found = true
							break
						}
					}

					if !found {
						s.T().Errorf("event %s not found", dymnstypes.EventTypeBoRefundOffer)
					}
				})
			}
		})
	}
}
