package keeper_test

import (
	"fmt"
	"time"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

func (s *KeeperTestSuite) Test_msgServer_PlaceSellOrder_DymName() {
	const daysProhibitSell = 30
	const daysSellOrderDuration = 7

	setupParams := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Misc.ProhibitSellDuration = daysProhibitSell * 24 * time.Hour
			moduleParams.Misc.SellOrderDuration = daysSellOrderDuration * 24 * time.Hour
			return moduleParams
		})
	}

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceSellOrder(s.ctx, &dymnstypes.MsgPlaceSellOrder{})

		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	const name = "my-name"

	ownerA := testAddr(1).bech32()
	notOwnerA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	coin100 := dymnsutils.TestCoin(100)
	coin200 := dymnsutils.TestCoin(200)
	coin300 := dymnsutils.TestCoin(300)

	tests := []struct {
		name                    string
		withoutDymName          bool
		existingSo              *dymnstypes.SellOrder
		dymNameExpiryOffsetDays int64
		customOwner             string
		customDymNameOwner      string
		minPrice                sdk.Coin
		sellPrice               *sdk.Coin
		preRunSetup             func(*KeeperTestSuite)
		wantErr                 bool
		wantErrContains         string
	}{
		{
			name:            "fail - Dym-Name does not exists",
			withoutDymName:  true,
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: fmt.Sprintf("Dym-Name: %s: not found", name),
		},
		{
			name:               "fail - wrong owner",
			customOwner:        ownerA,
			customDymNameOwner: notOwnerA,
			minPrice:           coin100,
			wantErr:            true,
			wantErrContains:    "not the owner of the Dym-Name",
		},
		{
			name:                    "fail - expired Dym-Name",
			withoutDymName:          false,
			existingSo:              nil,
			dymNameExpiryOffsetDays: -1,
			minPrice:                coin100,
			wantErr:                 true,
			wantErrContains:         "Dym-Name is already expired",
		},
		{
			name: "fail - existing active SO, not finished",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active Sell-Order already exists",
		},
		{
			name: "fail - existing active SO, expired",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "fail - existing active SO, not expired, completed",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "fail - existing active SO, expired, completed",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeName,
				ExpireAt:  s.now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists",
		},
		{
			name:            "fail - not allowed denom",
			minPrice:        sdk.NewInt64Coin("u"+s.priceDenom(), 100),
			wantErr:         true,
			wantErrContains: fmt.Sprintf("the only denom allowed as price: %s", s.priceDenom()),
		},
		{
			name:                    "fail - can not sell Dym-Name that almost expired",
			dymNameExpiryOffsetDays: daysProhibitSell - 1,
			minPrice:                coin100,
			wantErr:                 true,
			wantErrContains:         "duration before Dym-Name expiry, prohibited to sell",
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, without sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               nil,
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, without sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               dymnsutils.TestCoinP(0),
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, with sell price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               &coin300,
		},
		{
			name:                    "pass - successfully place Dym-Name Sell-Order, with sell price equals to min-price",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               &coin100,
		},
		{
			name:                    "fail - can NOT place Dym-Name Sell-Order, when Dym-Name trading is disabled",
			dymNameExpiryOffsetDays: 9999,
			minPrice:                coin100,
			sellPrice:               nil,
			preRunSetup: func(*KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Misc.EnableTradingName = false
					return moduleParams
				})
			},
			wantErr:         true,
			wantErrContains: "trading of Dym-Name is disabled",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			setupParams(s)

			useDymNameOwner := ownerA
			if tt.customDymNameOwner != "" {
				useDymNameOwner = tt.customDymNameOwner
			}
			useDymNameExpiry := s.ctx.BlockTime().Add(
				time.Hour * 24 * time.Duration(tt.dymNameExpiryOffsetDays),
			).Unix()

			if !tt.withoutDymName {
				dymName := dymnstypes.DymName{
					Name:       name,
					Owner:      useDymNameOwner,
					Controller: useDymNameOwner,
					ExpireAt:   useDymNameExpiry,
				}
				err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
				s.Require().NoError(err)
			}

			if tt.existingSo != nil {
				tt.existingSo.AssetId = name
				err := s.dymNsKeeper.SetSellOrder(s.ctx, *tt.existingSo)
				s.Require().NoError(err)
			}

			useOwner := ownerA
			if tt.customOwner != "" {
				useOwner = tt.customOwner
			}
			msg := &dymnstypes.MsgPlaceSellOrder{
				AssetId:   name,
				AssetType: dymnstypes.TypeName,
				MinPrice:  tt.minPrice,
				SellPrice: tt.sellPrice,
				Owner:     useOwner,
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceSellOrder(s.ctx, msg)
			moduleParams := s.dymNsKeeper.GetParams(s.ctx)

			defer func() {
				laterDymName := s.dymNsKeeper.GetDymName(s.ctx, name)
				if tt.withoutDymName {
					s.Require().Nil(laterDymName)
					return
				}

				s.Require().NotNil(laterDymName)
				s.Require().Equal(dymnstypes.DymName{
					Name:       name,
					Owner:      useDymNameOwner,
					Controller: useDymNameOwner,
					ExpireAt:   useDymNameExpiry,
				}, *laterDymName, "Dym-Name record should not be changed in any case")
			}()

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)

				s.Require().Nil(resp)

				so := s.dymNsKeeper.GetSellOrder(s.ctx, name, dymnstypes.TypeName)
				if tt.existingSo != nil {
					s.Require().NotNil(so)
					s.Require().Equal(*tt.existingSo, *so)
				} else {
					s.Require().Nil(so)
				}

				s.Require().Less(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceSellOrder,
					"should not consume params gas on failed operation",
				)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)

			so := s.dymNsKeeper.GetSellOrder(s.ctx, name, dymnstypes.TypeName)
			s.Require().NotNil(so)

			expectedSo := dymnstypes.SellOrder{
				AssetId:    name,
				AssetType:  dymnstypes.TypeName,
				ExpireAt:   s.ctx.BlockTime().Add(moduleParams.Misc.SellOrderDuration).Unix(),
				MinPrice:   msg.MinPrice,
				SellPrice:  msg.SellPrice,
				HighestBid: nil,
			}
			if !expectedSo.HasSetSellPrice() {
				expectedSo.SellPrice = nil
			}

			s.Require().Nil(so.HighestBid, "highest bid should not be set")

			s.Require().Equal(expectedSo, *so)

			s.Require().GreaterOrEqual(
				s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceSellOrder,
				"should consume params gas",
			)

			aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeName)

			var found bool
			for _, record := range aSoe.Records {
				if record.AssetId == name {
					found = true
					s.Require().Equal(expectedSo.ExpireAt, record.ExpireAt)
					break
				}
			}

			s.Require().True(found)
		})
	}
}

func (s *KeeperTestSuite) Test_msgServer_PlaceSellOrder_Alias() {
	const daysSellOrderDuration = 7
	denom := dymnsutils.TestCoin(0).Denom

	setupParams := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Misc.SellOrderDuration = daysSellOrderDuration * 24 * time.Hour
			return moduleParams
		})
	}

	const srcRollAppId = "rollapp_1-1"
	const alias = "alias"
	const dstRollAppId = "rollapp_2-2"

	ownerA := testAddr(1).bech32()
	notOwnerA := testAddr(2).bech32()
	bidderA := testAddr(3).bech32()

	coin100 := dymnsutils.TestCoin(100)
	coin200 := dymnsutils.TestCoin(200)
	coin300 := dymnsutils.TestCoin(300)

	tests := []struct {
		name               string
		withoutAlias       bool
		existingSo         *dymnstypes.SellOrder
		customOwner        string
		customRollAppOwner string
		minPrice           sdk.Coin
		sellPrice          *sdk.Coin
		preRunSetup        func(*KeeperTestSuite)
		wantErr            bool
		wantErrContains    string
	}{
		{
			name:            "fail - alias does not exists",
			withoutAlias:    true,
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "alias: alias: not found",
		},
		{
			name:               "fail - wrong owner",
			customOwner:        ownerA,
			customRollAppOwner: notOwnerA,
			minPrice:           coin100,
			wantErr:            true,
			wantErrContains:    "not the owner of the RollApp using the alias",
		},
		{
			name: "fail - existing active SO, not finished",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeAlias,
				ExpireAt:  s.now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active Sell-Order already exists",
		},
		{
			name: "fail - existing active SO, expired",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeAlias,
				ExpireAt:  s.now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "fail - existing active SO, not expired, completed",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeAlias,
				ExpireAt:  s.now.Add(time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
					Params: []string{dstRollAppId},
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists ",
		},
		{
			name: "fail - existing active SO, expired, completed",
			existingSo: &dymnstypes.SellOrder{
				AssetType: dymnstypes.TypeAlias,
				ExpireAt:  s.now.Add(-1 * time.Hour).Unix(),
				MinPrice:  coin100,
				SellPrice: &coin200,
				HighestBid: &dymnstypes.SellOrderBid{
					Bidder: bidderA,
					Price:  coin200,
					Params: []string{dstRollAppId},
				},
			},
			minPrice:        coin100,
			wantErr:         true,
			wantErrContains: "an active expired/completed Sell-Order already exists",
		},
		{
			name:            "fail - not allowed denom",
			minPrice:        sdk.NewInt64Coin("u"+denom, 100),
			wantErr:         true,
			wantErrContains: fmt.Sprintf("the only denom allowed as price: %s", denom),
		},
		{
			name:      "pass - successfully place Alias Sell-Order, without sell price",
			minPrice:  coin100,
			sellPrice: nil,
		},
		{
			name:      "fail - can NOT place sell order if alias which present in params",
			minPrice:  coin100,
			sellPrice: nil,
			preRunSetup: func(*KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "some-chain",
						Aliases: []string{alias},
					},
				}
				err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "prohibited to trade aliases which is reserved for chain-id or alias in module params",
		},
		{
			name:      "pass - successfully place Alias Sell-Order, without sell price",
			minPrice:  coin100,
			sellPrice: dymnsutils.TestCoinP(0),
		},
		{
			name:      "pass - successfully place Alias Sell-Order, with sell price",
			minPrice:  coin100,
			sellPrice: &coin300,
		},
		{
			name:      "pass - successfully place Alias Sell-Order, with sell price equals to min-price",
			minPrice:  coin100,
			sellPrice: &coin100,
		},
		{
			name:      "fail - can NOT place Alias Sell-Order, when Alias trading is disabled",
			minPrice:  coin100,
			sellPrice: nil,
			preRunSetup: func(*KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Misc.EnableTradingAlias = false
				err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "trading of Alias is disabled",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			setupParams(s)

			useRollAppOwner := ownerA
			if tt.customRollAppOwner != "" {
				useRollAppOwner = tt.customRollAppOwner
			}

			s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
				RollappId: srcRollAppId,
				Owner:     useRollAppOwner,
			})
			if !tt.withoutAlias {
				err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, srcRollAppId, alias)
				s.Require().NoError(err)
			}

			if tt.existingSo != nil {
				tt.existingSo.AssetId = alias
				err := s.dymNsKeeper.SetSellOrder(s.ctx, *tt.existingSo)
				s.Require().NoError(err)
			}

			useOwner := ownerA
			if tt.customOwner != "" {
				useOwner = tt.customOwner
			}
			msg := &dymnstypes.MsgPlaceSellOrder{
				AssetId:   alias,
				AssetType: dymnstypes.TypeAlias,
				MinPrice:  tt.minPrice,
				SellPrice: tt.sellPrice,
				Owner:     useOwner,
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceSellOrder(s.ctx, msg)
			moduleParams := s.dymNsKeeper.GetParams(s.ctx)

			defer func() {
				if tt.withoutAlias {
					s.requireAlias(alias).NotInUse()
					s.requireRollApp(srcRollAppId).HasNoAlias()
				} else {
					s.requireAlias(alias).LinkedToRollApp(srcRollAppId)
				}
			}()

			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)

				s.Require().Nil(resp)

				so := s.dymNsKeeper.GetSellOrder(s.ctx, alias, dymnstypes.TypeAlias)
				if tt.existingSo != nil {
					s.Require().NotNil(so)
					s.Require().Equal(*tt.existingSo, *so)
				} else {
					s.Require().Nil(so)
				}

				s.Require().Less(
					s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceSellOrder,
					"should not consume params gas on failed operation",
				)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)

			so := s.dymNsKeeper.GetSellOrder(s.ctx, alias, dymnstypes.TypeAlias)
			s.Require().NotNil(so)

			expectedSo := dymnstypes.SellOrder{
				AssetId:    alias,
				AssetType:  dymnstypes.TypeAlias,
				ExpireAt:   s.ctx.BlockTime().Add(moduleParams.Misc.SellOrderDuration).Unix(),
				MinPrice:   msg.MinPrice,
				SellPrice:  msg.SellPrice,
				HighestBid: nil,
			}
			if !expectedSo.HasSetSellPrice() {
				expectedSo.SellPrice = nil
			}

			s.Require().Nil(so.HighestBid, "highest bid should not be set")

			s.Require().Equal(expectedSo, *so)

			s.Require().GreaterOrEqual(
				s.ctx.GasMeter().GasConsumed(), dymnstypes.OpGasPlaceSellOrder,
				"should consume params gas",
			)

			aSoe := s.dymNsKeeper.GetActiveSellOrdersExpiration(s.ctx, dymnstypes.TypeAlias)

			var found bool
			for _, record := range aSoe.Records {
				if record.AssetId == alias {
					found = true
					s.Require().Equal(expectedSo.ExpireAt, record.ExpireAt)
					break
				}
			}

			s.Require().True(found)
		})
	}
}
