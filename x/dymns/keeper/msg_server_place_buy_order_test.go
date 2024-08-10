package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (s *KeeperTestSuite) Test_msgServer_PlaceBuyOrder_DymName() {
	const minOfferPrice = 5

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	minOfferPriceCoin := sdk.NewCoin(s.priceDenom(), sdk.NewInt(minOfferPrice).Mul(priceMultiplier))

	ownerA := testAddr(1).bech32()
	buyerA := testAddr(2).bech32()
	anotherBuyerA := testAddr(3).bech32()

	setupParams := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Price.MinOfferPrice = minOfferPriceCoin.Amount
			// force enable trading
			moduleParams.Misc.EnableTradingName = true
			moduleParams.Misc.EnableTradingAlias = true
			return moduleParams
		})
	}

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Add(9 * 365 * 24 * time.Hour).Unix(),
	}

	tests := []struct {
		name                        string
		existingDymName             *dymnstypes.DymName
		existingOffer               *dymnstypes.BuyOrder
		dymName                     string
		buyer                       string
		offer                       sdk.Coin
		existingBuyOrderId          string
		originalModuleBalance       sdkmath.Int
		originalBuyerBalance        sdkmath.Int
		originalAnotherBuyerBalance sdkmath.Int
		preRunSetupFunc             func(s *KeeperTestSuite)
		wantErr                     bool
		wantErrContains             string
		wantBuyOrderId              string
		wantLaterOffer              *dymnstypes.BuyOrder
		wantLaterModuleBalance      sdkmath.Int
		wantLaterBuyerBalance       sdkmath.Int
		wantMinConsumeGas           sdk.Gas
		afterTestFunc               func(s *KeeperTestSuite)
	}{
		{
			name:                  "pass - can place offer",
			existingDymName:       dymName,
			existingOffer:         nil,
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin,
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantErr:               false,
			wantBuyOrderId:        "101",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.AddRaw(5),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:            "pass - can extends offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "102",
			originalModuleBalance: sdk.NewInt(0),
			originalBuyerBalance:  sdk.NewInt(1),
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "fail - can NOT extends offer of type mis-match",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{"rollapp_1-1"},
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(0),
			originalBuyerBalance:  sdk.NewInt(1),
			wantErr:               true,
			wantErrContains:       "asset type mismatch with existing offer",
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{"rollapp_1-1"},
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      1,
		},
		{
			name:            "pass - can extends offer with counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "102",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  sdk.NewInt(2),
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			wantLaterModuleBalance: sdk.NewInt(2),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "pass - can extends offer with offer equals to counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(3)),
			existingBuyOrderId:    "102",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  sdk.NewInt(5),
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(3)),          // updated
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))), // unchanged
			},
			wantLaterModuleBalance: sdk.NewInt(4),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "pass - can extends offer with offer greater than counterparty offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(4)),
			existingBuyOrderId:    "102",
			originalModuleBalance: sdk.NewInt(2),
			originalBuyerBalance:  sdk.NewInt(5),
			wantErr:               false,
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(4)),
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			wantLaterModuleBalance: sdk.NewInt(6),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:            "pass - extends an existing offer only take the extra amount instead of all",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
			existingBuyOrderId:    "101",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  sdk.NewInt(3),
			wantErr:               false,
			wantBuyOrderId:        "101",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
			},
			wantLaterModuleBalance: sdk.NewInt(7),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "fail - can NOT place offer if trading Dym-Name is disabled",
			existingDymName:       dymName,
			existingOffer:         nil,
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin,
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Misc.EnableTradingName = false
				err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "trading of Dym-Name is disabled",
			wantBuyOrderId:         "",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(5),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - reject offer for non-existing Dym-Name",
			existingDymName:        nil,
			dymName:                "non-exists",
			buyer:                  buyerA,
			offer:                  minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "Dym-Name: non-exists: not found",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name: "fail - reject offer for expired Dym-Name",
			existingDymName: &dymnstypes.DymName{
				Name:       "expired",
				Owner:      ownerA,
				Controller: ownerA,
				ExpireAt:   s.now.Unix() - 1,
			},
			dymName:                "expired",
			buyer:                  buyerA,
			offer:                  minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "Dym-Name: expired: not found",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:                   "fail - can not offer own Dym-Name",
			existingDymName:        dymName,
			dymName:                dymName.Name,
			buyer:                  dymName.Owner,
			offer:                  minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "cannot buy own Dym-Name",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name: "fail - reject Dym-Name that almost expired",
			existingDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      dymName.Owner,
				Controller: dymName.Owner,
				ExpireAt:   s.now.Add(1 * time.Second).Unix(),
			},
			dymName:                dymName.Name,
			buyer:                  buyerA,
			offer:                  minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "duration before Dym-Name expiry, prohibited to trade",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:            "fail - offer denom must match params",
			existingDymName: dymName,
			dymName:         dymName.Name,
			buyer:           buyerA,
			offer: sdk.Coin{
				Denom:  "u" + s.priceDenom(),
				Amount: sdk.NewInt(minOfferPrice),
			},
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "invalid offer denomination, only accept",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:                   "fail - offer price can not lower than min defined in params",
			existingDymName:        dymName,
			dymName:                dymName.Name,
			buyer:                  buyerA,
			offer:                  minOfferPriceCoin.SubAmount(sdk.NewInt(1)),
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "offer price must be greater than or equal to",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:            "pass - if NOT continue offer, create another and charges full offer price",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "",
			originalModuleBalance: minOfferPriceCoin.Amount,
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:        false,
			wantBuyOrderId: "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.Add(minOfferPriceCoin.Amount.AddRaw(1)),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:            "fail - continue a non-existing offer",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "102",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Buy-Order ID: 102: not found",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but not yours",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  anotherBuyerA, // not the buyer
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA, // not the existing offer's buyer
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "101",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "not the owner of the offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      anotherBuyerA,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but the Dym-Name mismatch",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                "another-name",
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "101",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Dym-Name mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    "another-name",
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but mis-match offer denom",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:        "101",
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				Buyer:     buyerA,
				OfferPrice: sdk.Coin{
					Denom:  "u" + s.priceDenom(),
					Amount: sdk.NewInt(minOfferPrice),
				},
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "101",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer denomination mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:        "101",
				AssetId:   dymName.Name,
				AssetType: dymnstypes.TypeName,
				Buyer:     buyerA,
				OfferPrice: sdk.Coin{
					Denom:  "u" + s.priceDenom(),
					Amount: sdk.NewInt(minOfferPrice),
				},
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but new offer less than previous",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)), // less
			existingBuyOrderId:    "101",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)), // keep
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "102"))
			},
		},
		{
			name:            "fail - continue an existing offer but new offer equals to previous",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "101",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				CounterpartyOfferPrice: nil,
			},
			dymName:               dymName.Name,
			buyer:                 buyerA,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(2)), // same
			existingBuyOrderId:    "101",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "102"))
			},
		},
		{
			name:                        "pass - reverse record added after successful offer",
			existingDymName:             dymName,
			existingOffer:               nil,
			dymName:                     dymName.Name,
			buyer:                       buyerA,
			offer:                       minOfferPriceCoin,
			existingBuyOrderId:          "",
			originalModuleBalance:       sdk.NewInt(5),
			originalBuyerBalance:        minOfferPriceCoin.Amount.AddRaw(2),
			originalAnotherBuyerBalance: minOfferPriceCoin.Amount,
			wantErr:                     false,
			wantBuyOrderId:              "101",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "101",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.AddRaw(5),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"101"}, orderIds.OrderIds)

				offers, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName.Name)
				s.Require().NoError(err)
				s.Equal("101", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"101"}, orderIds.OrderIds)

				offers, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyerA)
				s.Require().NoError(err)
				s.Equal("101", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   dymName.Name,
					AssetType: dymnstypes.TypeName,
					Buyer:     anotherBuyerA,
					Offer:     minOfferPriceCoin,
				})
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Equal("102", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(anotherBuyerA))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"102"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"101"}, orderIds.OrderIds)

				key = dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"101", "102"}, orderIds.OrderIds)
			},
		},
		{
			name:            "pass - reverse record added after successful offer extends",
			existingDymName: dymName,
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "102",
				AssetId:                dymName.Name,
				AssetType:              dymnstypes.TypeName,
				Buyer:                  buyerA,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			dymName:                     dymName.Name,
			buyer:                       buyerA,
			offer:                       minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:          "102",
			originalModuleBalance:       sdk.NewInt(0),
			originalBuyerBalance:        sdk.NewInt(1),
			originalAnotherBuyerBalance: minOfferPriceCoin.Amount,
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 2)

				err := s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, buyerA, "102")
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, dymName.Name, dymnstypes.TypeName, "102")
				s.Require().NoError(err)
			},
			wantErr:        false,
			wantBuyOrderId: "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    dymName.Name,
				AssetType:  dymnstypes.TypeName,
				Buyer:      buyerA,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"102"}, orderIds.OrderIds)

				offers, err := s.dymNsKeeper.GetBuyOrdersOfDymName(s.ctx, dymName.Name)
				s.Require().NoError(err)
				s.Equal("102", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"102"}, orderIds.OrderIds)

				offers, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, buyerA)
				s.Require().NoError(err)
				s.Equal("102", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   dymName.Name,
					AssetType: dymnstypes.TypeName,
					Buyer:     anotherBuyerA,
					Offer:     minOfferPriceCoin,
				})
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Equal("103", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(anotherBuyerA))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"103"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(buyerA))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"102"}, orderIds.OrderIds)

				key = dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"102", "103"}, orderIds.OrderIds)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			setupParams(s)

			if tt.originalModuleBalance.IsPositive() {
				s.mintToModuleAccount2(tt.originalModuleBalance)
			}

			if tt.originalBuyerBalance.IsPositive() {
				s.mintToAccount2(tt.buyer, tt.originalBuyerBalance)
			}

			if !tt.originalAnotherBuyerBalance.IsNil() && tt.originalAnotherBuyerBalance.IsPositive() {
				s.mintToAccount2(anotherBuyerA, tt.originalAnotherBuyerBalance)
			}

			if tt.existingDymName != nil {
				err := s.dymNsKeeper.SetDymName(s.ctx, *tt.existingDymName)
				s.Require().NoError(err)
			}

			if tt.existingOffer != nil {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *tt.existingOffer)
				s.Require().NoError(err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{
				AssetId:         tt.dymName,
				AssetType:       dymnstypes.TypeName,
				Buyer:           tt.buyer,
				ContinueOrderId: tt.existingBuyOrderId,
				Offer:           tt.offer,
			})

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.wantLaterOffer.Id)
					s.Require().NotNil(laterOffer)
					s.Equal(*tt.wantLaterOffer, *laterOffer)
				}

				laterModuleBalance := s.moduleBalance2()
				s.Equal(tt.wantLaterModuleBalance.String(), laterModuleBalance.String())

				laterBuyerBalance := s.balance2(tt.buyer)
				s.Equal(tt.wantLaterBuyerBalance.String(), laterBuyerBalance.String())

				s.Less(tt.wantMinConsumeGas, s.ctx.GasMeter().GasConsumed())

				if tt.existingDymName != nil {
					originalDymName := *tt.existingDymName
					laterDymName := s.dymNsKeeper.GetDymName(s.ctx, tt.dymName)
					s.Require().NotNil(laterDymName)
					s.Equal(originalDymName, *laterDymName, "Dym-Name record should not be changed")
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				s.Require().Nil(resp)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Equal(tt.wantBuyOrderId, resp.OrderId)
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_PlaceBuyOrder_Alias() {
	const minOfferPrice = 5

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	minOfferPriceCoin := sdk.NewCoin(s.priceDenom(), sdk.NewInt(minOfferPrice).Mul(priceMultiplier))

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	creator_3_asAnotherBuyer := testAddr(3).bech32()

	setupParams := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Price.MinOfferPrice = minOfferPriceCoin.Amount
			// force enable trading
			moduleParams.Misc.EnableTradingName = true
			moduleParams.Misc.EnableTradingAlias = true
			return moduleParams
		})
	}

	type rollapp struct {
		rollAppID string
		creator   string
		alias     string
	}

	rollApp_1_by1_asSrc := rollapp{
		rollAppID: "rollapp_1-1",
		creator:   creator_1_asOwner,
		alias:     "one",
	}
	rollApp_2_by2_asDest := rollapp{
		rollAppID: "rollapp_2-1",
		creator:   creator_2_asBuyer,
		alias:     "two",
	}
	rollApp_3_by3_asDest_noAlias := rollapp{
		rollAppID: "rollapp_3-1",
		creator:   creator_3_asAnotherBuyer,
		alias:     "",
	}
	rollApp_4_by1_asDest_noAlias := rollapp{
		rollAppID: "rollapp_4-2",
		creator:   creator_1_asOwner,
		alias:     "",
	}

	tests := []struct {
		name                        string
		existingRollApps            []rollapp
		existingOffer               *dymnstypes.BuyOrder
		alias                       string
		buyer                       string
		dstRollAppId                string // destination RollApp ID
		offer                       sdk.Coin
		existingBuyOrderId          string
		originalModuleBalance       sdkmath.Int
		originalBuyerBalance        sdkmath.Int
		originalAnotherBuyerBalance sdkmath.Int
		preRunSetupFunc             func(s *KeeperTestSuite)
		wantErr                     bool
		wantErrContains             string
		wantBuyOrderId              string
		wantLaterOffer              *dymnstypes.BuyOrder
		wantLaterModuleBalance      sdkmath.Int
		wantLaterBuyerBalance       sdkmath.Int
		wantMinConsumeGas           sdk.Gas
		afterTestFunc               func(s *KeeperTestSuite)
	}{
		{
			name:                  "pass - can place offer",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin,
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantErr:               false,
			wantBuyOrderId:        "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.AddRaw(5),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:                  "fail - can not place offer of alias which presents in params",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin,
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "some-chain",
						Aliases: []string{rollApp_1_by1_asSrc.alias},
					},
				}
				err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(5),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
		},
		{
			name:             "pass - can extends offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(0),
			originalBuyerBalance:  sdk.NewInt(1),
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "fail - can NOT extend offer of alias which presents in params",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(0),
			originalBuyerBalance:  sdk.NewInt(1),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "some-chain",
						Aliases: []string{rollApp_1_by1_asSrc.alias},
					},
				}
				err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
				s.Require().NoError(err)
			},
			wantErr:         true,
			wantErrContains: "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantBuyOrderId:  "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - can NOT extends offer of type mis-match",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeName,
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "102",
			originalModuleBalance: sdk.NewInt(0),
			originalBuyerBalance:  sdk.NewInt(1),
			wantErr:               true,
			wantErrContains:       "asset type mismatch with existing offer",
			wantBuyOrderId:        "102",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "102",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeName,
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      1,
		},
		{
			name:             "pass - can extends offer with counterparty offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  sdk.NewInt(6),
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			wantLaterModuleBalance: sdk.NewInt(2),
			wantLaterBuyerBalance:  sdk.NewInt(5),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - can extends offer with offer equals to counterparty offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(3)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(2),
			originalBuyerBalance:  sdk.NewInt(6),
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(3)),
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			wantLaterModuleBalance: sdk.NewInt(5),
			wantLaterBuyerBalance:  sdk.NewInt(3),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - can extends offer with offer greater than counterparty offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(5)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  sdk.NewInt(7),
			wantErr:               false,
			wantBuyOrderId:        "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(5)),
				CounterpartyOfferPrice: uptr.To(minOfferPriceCoin.AddAmount(sdk.NewInt(3))),
			},
			wantLaterModuleBalance: sdk.NewInt(6),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - extends an existing offer only take the extra amount instead of all",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
			existingBuyOrderId:    "201",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  sdk.NewInt(3),
			wantErr:               false,
			wantBuyOrderId:        "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
			},
			wantLaterModuleBalance: sdk.NewInt(7),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "fail - can NOT place offer if trading Alias is disabled",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin,
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Misc.EnableTradingAlias = false
				err := s.dymNsKeeper.SetParams(s.ctx, moduleParams)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "trading of Alias is disabled",
			wantBuyOrderId:         "",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(5),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - reject offer for non-existing Alias",
			existingRollApps:       []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			alias:                  "void",
			buyer:                  rollApp_2_by2_asDest.creator,
			dstRollAppId:           rollApp_2_by2_asDest.rollAppID,
			offer:                  minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "alias is not in-used: void: not found",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:                  "pass - can place offer buy own alias, different RollApp",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_4_by1_asDest_noAlias},
			existingOffer:         nil,
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_1_by1_asSrc.creator,
			dstRollAppId:          rollApp_4_by1_asDest_noAlias.rollAppID,
			offer:                 minOfferPriceCoin,
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(5),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantErr:               false,
			wantBuyOrderId:        "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_4_by1_asDest_noAlias.rollAppID},
				Buyer:      rollApp_1_by1_asSrc.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.AddRaw(5),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:             "fail - offer denom must match params",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			alias:            rollApp_1_by1_asSrc.alias,
			buyer:            rollApp_2_by2_asDest.creator,
			dstRollAppId:     rollApp_2_by2_asDest.rollAppID,
			offer: sdk.Coin{
				Denom:  "u" + s.priceDenom(),
				Amount: sdk.NewInt(minOfferPrice),
			},
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "invalid offer denomination, only accept",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:                   "fail - offer price can not lower than min defined in params",
			existingRollApps:       []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			alias:                  rollApp_1_by1_asSrc.alias,
			buyer:                  rollApp_2_by2_asDest.creator,
			dstRollAppId:           rollApp_2_by2_asDest.rollAppID,
			offer:                  minOfferPriceCoin.SubAmount(sdk.NewInt(1)),
			originalModuleBalance:  sdk.NewInt(1),
			originalBuyerBalance:   minOfferPriceCoin.Amount,
			wantErr:                true,
			wantErrContains:        "offer price must be greater than or equal to",
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
			},
		},
		{
			name:             "pass - if NOT continue offer, create another and charges full offer price",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "",
			originalModuleBalance: sdk.NewInt(2),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:        false,
			wantBuyOrderId: "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.AddRaw(1).AddRaw(2),
			wantLaterBuyerBalance:  sdk.NewInt(1),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
		},
		{
			name:             "fail - continue a non-existing offer",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "202",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "Buy-Order ID: 202: not found",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but not yours",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  creator_3_asAnotherBuyer, // not the buyer
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator, // not the existing offer's buyer
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "201",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "not the owner of the offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      creator_3_asAnotherBuyer,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but the Alias mismatch",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                "another",
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "201",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "alias mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    "another",
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but mis-match offer denom",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:        "201",
				AssetId:   rollApp_1_by1_asSrc.alias,
				AssetType: dymnstypes.TypeAlias,
				Params:    []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:     rollApp_2_by2_asDest.creator,
				OfferPrice: sdk.Coin{
					Denom:  "u" + s.priceDenom(),
					Amount: sdk.NewInt(minOfferPrice),
				},
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:    "201",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer denomination mismatch with existing offer",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:        "201",
				AssetId:   rollApp_1_by1_asSrc.alias,
				AssetType: dymnstypes.TypeAlias,
				Params:    []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:     rollApp_2_by2_asDest.creator,
				OfferPrice: sdk.Coin{
					Denom:  "u" + s.priceDenom(),
					Amount: sdk.NewInt(minOfferPrice),
				},
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but new offer less than previous",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(1)), // less
			existingBuyOrderId:    "201",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)), // keep
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "202"))
			},
		},
		{
			name:             "fail - continue an existing offer but new offer equals to previous",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "201",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				CounterpartyOfferPrice: nil,
			},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_2_by2_asDest.creator,
			dstRollAppId:          rollApp_2_by2_asDest.rollAppID,
			offer:                 minOfferPriceCoin.AddAmount(sdk.NewInt(2)), // same
			existingBuyOrderId:    "201",
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 1)
			},
			wantErr:         true,
			wantErrContains: "offer price must be greater than existing offer price",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Len(s.dymNsKeeper.GetAllBuyOrders(s.ctx), 1)
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "202"))
			},
		},
		{
			name:                  "fail - destination RollApp is not found",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 creator_2_asBuyer,
			dstRollAppId:          "nah_0-0",
			offer:                 minOfferPriceCoin,
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
			},
			wantErr:                true,
			wantErrContains:        "destination Roll-App does not exists",
			wantBuyOrderId:         "201",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "201"))
			},
		},
		{
			name:                  "fail - destination RollApp is not found",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc, rollApp_3_by3_asDest_noAlias},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 creator_2_asBuyer,
			dstRollAppId:          rollApp_3_by3_asDest_noAlias.rollAppID,
			offer:                 minOfferPriceCoin,
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
			},
			wantErr:                true,
			wantErrContains:        "not the owner of the RollApp",
			wantBuyOrderId:         "201",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "201"))
			},
		},
		{
			name:                  "fail - destination RollApp is the same as source",
			existingRollApps:      []rollapp{rollApp_1_by1_asSrc},
			alias:                 rollApp_1_by1_asSrc.alias,
			buyer:                 rollApp_1_by1_asSrc.creator,
			dstRollAppId:          rollApp_1_by1_asSrc.rollAppID,
			offer:                 minOfferPriceCoin,
			originalModuleBalance: sdk.NewInt(1),
			originalBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
			},
			wantErr:                true,
			wantErrContains:        "destination Roll-App ID is the same as the source",
			wantBuyOrderId:         "201",
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  minOfferPriceCoin.Amount.AddRaw(2),
			wantMinConsumeGas:      1,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.Require().Empty(s.dymNsKeeper.GetAllBuyOrders(s.ctx))
				s.Require().Nil(s.dymNsKeeper.GetBuyOrder(s.ctx, "201"))
			},
		},
		{
			name:                        "pass - reverse record added after successful offer",
			existingRollApps:            []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest, rollApp_3_by3_asDest_noAlias},
			existingOffer:               nil,
			alias:                       rollApp_1_by1_asSrc.alias,
			buyer:                       rollApp_2_by2_asDest.creator,
			dstRollAppId:                rollApp_2_by2_asDest.rollAppID,
			offer:                       minOfferPriceCoin,
			existingBuyOrderId:          "",
			originalModuleBalance:       sdk.NewInt(5),
			originalBuyerBalance:        minOfferPriceCoin.Amount.AddRaw(2),
			originalAnotherBuyerBalance: minOfferPriceCoin.Amount,
			wantErr:                     false,
			wantBuyOrderId:              "201",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "201",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin,
			},
			wantLaterModuleBalance: minOfferPriceCoin.Amount.AddRaw(5),
			wantLaterBuyerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasPutBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"201"}, orderIds.OrderIds)

				offers, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, rollApp_1_by1_asSrc.alias)
				s.Require().NoError(err)
				s.Equal("201", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"201"}, orderIds.OrderIds)

				offers, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, rollApp_2_by2_asDest.creator)
				s.Require().NoError(err)
				s.Equal("201", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   rollApp_1_by1_asSrc.alias,
					AssetType: dymnstypes.TypeAlias,
					Params:    []string{rollApp_3_by3_asDest_noAlias.rollAppID},
					Buyer:     rollApp_3_by3_asDest_noAlias.creator,
					Offer:     minOfferPriceCoin,
				})
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Equal("202", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_3_by3_asDest_noAlias.creator))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"202"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"201"}, orderIds.OrderIds)

				key = dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"201", "202"}, orderIds.OrderIds)
			},
		},
		{
			name:             "pass - reverse record added after successful offer extends",
			existingRollApps: []rollapp{rollApp_1_by1_asSrc, rollApp_2_by2_asDest, rollApp_3_by3_asDest_noAlias},
			existingOffer: &dymnstypes.BuyOrder{
				Id:                     "202",
				AssetId:                rollApp_1_by1_asSrc.alias,
				AssetType:              dymnstypes.TypeAlias,
				Params:                 []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:                  rollApp_2_by2_asDest.creator,
				OfferPrice:             minOfferPriceCoin,
				CounterpartyOfferPrice: nil,
			},
			alias:                       rollApp_1_by1_asSrc.alias,
			buyer:                       rollApp_2_by2_asDest.creator,
			dstRollAppId:                rollApp_2_by2_asDest.rollAppID,
			offer:                       minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			existingBuyOrderId:          "202",
			originalModuleBalance:       sdk.NewInt(0),
			originalBuyerBalance:        sdk.NewInt(1),
			originalAnotherBuyerBalance: minOfferPriceCoin.Amount,
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.dymNsKeeper.SetCountBuyOrders(s.ctx, 2)

				err := s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, rollApp_2_by2_asDest.creator, "202")
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, rollApp_1_by1_asSrc.alias, dymnstypes.TypeAlias, "202")
				s.Require().NoError(err)
			},
			wantErr:        false,
			wantBuyOrderId: "202",
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         "202",
				AssetId:    rollApp_1_by1_asSrc.alias,
				AssetType:  dymnstypes.TypeAlias,
				Params:     []string{rollApp_2_by2_asDest.rollAppID},
				Buyer:      rollApp_2_by2_asDest.creator,
				OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(1)),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterBuyerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"202"}, orderIds.OrderIds)

				offers, err := s.dymNsKeeper.GetBuyOrdersOfAlias(s.ctx, rollApp_1_by1_asSrc.alias)
				s.Require().NoError(err)
				s.Equal("202", offers[0].Id)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"202"}, orderIds.OrderIds)

				offers, err = s.dymNsKeeper.GetBuyOrdersByBuyer(s.ctx, rollApp_2_by2_asDest.creator)
				s.Require().NoError(err)
				s.Equal("202", offers[0].Id)

				resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{
					AssetId:   rollApp_1_by1_asSrc.alias,
					AssetType: dymnstypes.TypeAlias,
					Params:    []string{rollApp_3_by3_asDest_noAlias.rollAppID},
					Buyer:     rollApp_3_by3_asDest_noAlias.creator,
					Offer:     minOfferPriceCoin,
				})
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Equal("203", resp.OrderId)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_3_by3_asDest_noAlias.creator))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"203"}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(rollApp_2_by2_asDest.creator))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"202"}, orderIds.OrderIds)

				key = dymnstypes.AliasToBuyOrderIdsRvlKey(rollApp_1_by1_asSrc.alias)
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Equal([]string{"202", "203"}, orderIds.OrderIds)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			setupParams(s)

			if tt.originalModuleBalance.IsPositive() {
				s.mintToModuleAccount2(tt.originalModuleBalance)
			}

			if tt.originalBuyerBalance.IsPositive() {
				s.mintToAccount2(tt.buyer, tt.originalBuyerBalance)
			}

			if !tt.originalAnotherBuyerBalance.IsNil() && tt.originalAnotherBuyerBalance.IsPositive() {
				s.mintToAccount2(creator_3_asAnotherBuyer, tt.originalAnotherBuyerBalance)
			}

			for _, existingRollApp := range tt.existingRollApps {
				s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
					RollappId: existingRollApp.rollAppID,
					Owner:     existingRollApp.creator,
				})
				if existingRollApp.alias != "" {
					err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, existingRollApp.rollAppID, existingRollApp.alias)
					s.Require().NoError(err)
				}
			}

			if tt.existingOffer != nil {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *tt.existingOffer)
				s.Require().NoError(err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).PlaceBuyOrder(s.ctx, &dymnstypes.MsgPlaceBuyOrder{
				AssetId:         tt.alias,
				AssetType:       dymnstypes.TypeAlias,
				Buyer:           tt.buyer,
				Params:          []string{tt.dstRollAppId},
				ContinueOrderId: tt.existingBuyOrderId,
				Offer:           tt.offer,
			})

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.wantLaterOffer.Id)
					s.Require().NotNil(laterOffer)
					s.Equal(*tt.wantLaterOffer, *laterOffer)
				}

				laterModuleBalance := s.moduleBalance2()
				s.Equal(tt.wantLaterModuleBalance.String(), laterModuleBalance.String())

				laterBuyerBalance := s.balance2(tt.buyer)
				s.Equal(tt.wantLaterBuyerBalance.String(), laterBuyerBalance.String())

				s.Less(tt.wantMinConsumeGas, s.ctx.GasMeter().GasConsumed())

				for _, existingRollApp := range tt.existingRollApps {
					rollApp, found := s.rollAppKeeper.GetRollapp(s.ctx, existingRollApp.rollAppID)
					s.Require().True(found)
					s.Equal(existingRollApp.creator, rollApp.Owner)
					if existingRollApp.alias != "" {
						s.requireRollApp(existingRollApp.rollAppID).HasAlias(existingRollApp.alias)
					} else {
						s.requireRollApp(existingRollApp.rollAppID).HasNoAlias()
					}
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				s.Require().Nil(resp)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
			s.Equal(tt.wantBuyOrderId, resp.OrderId)
		})
	}
}
