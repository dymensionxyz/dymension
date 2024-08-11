package keeper_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/dymensionxyz/sdk-utils/utils/uptr"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (s *KeeperTestSuite) Test_msgServer_AcceptBuyOrder_Type_DymName() {
	const minOfferPrice = 5
	const daysProhibitSell = 30

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	minOfferPriceCoin := sdk.NewCoin(s.priceDenom(), sdk.NewInt(minOfferPrice).Mul(priceMultiplier))

	buyerA := testAddr(1).bech32()
	ownerA := testAddr(2).bech32()
	anotherOwnerA := testAddr(3).bech32()

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.MinOfferPrice = minOfferPriceCoin.Amount
		moduleParams.Misc.ProhibitSellDuration = daysProhibitSell * 24 * time.Hour
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		return moduleParams
	})
	s.MakeAnchorContext()

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).AcceptBuyOrder(s.ctx, &dymnstypes.MsgAcceptBuyOrder{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Add(daysProhibitSell * 24 * time.Hour).Add(time.Second).Unix(),
	}

	sameDymNameButOwnedByAnother := &dymnstypes.DymName{
		Name:       dymName.Name,
		Owner:      anotherOwnerA,
		Controller: anotherOwnerA,
		ExpireAt:   dymName.ExpireAt,
	}

	offer := &dymnstypes.BuyOrder{
		Id:         "101",
		AssetId:    dymName.Name,
		AssetType:  dymnstypes.TypeName,
		Buyer:      buyerA,
		OfferPrice: minOfferPriceCoin,
	}

	tests := []struct {
		name                   string
		existingDymName        *dymnstypes.DymName
		existingOffer          *dymnstypes.BuyOrder
		buyOrderId             string
		owner                  string
		minAccept              sdk.Coin
		originalModuleBalance  sdkmath.Int
		originalOwnerBalance   sdkmath.Int
		preRunSetupFunc        func(s *KeeperTestSuite)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOrder
		wantLaterDymName       *dymnstypes.DymName
		wantLaterModuleBalance sdkmath.Int
		wantLaterOwnerBalance  sdkmath.Int
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(s *KeeperTestSuite)
	}{
		{
			name:                  "pass - can accept offer (match)",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer:        nil,
			wantLaterDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      offer.Buyer,
				Controller: offer.Buyer,
				ExpireAt:   dymName.ExpireAt,
			},
			wantLaterModuleBalance: sdkmath.ZeroInt(),
			wantLaterOwnerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "pass - after match offer, reverse records of the offer are removed",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offer.Id}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offer.Id}, orderIds.OrderIds)
			},
			wantErr:        false,
			wantLaterOffer: nil,
			wantLaterDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      offer.Buyer,
				Controller: offer.Buyer,
				ExpireAt:   dymName.ExpireAt,
			},
			wantLaterModuleBalance: sdkmath.ZeroInt(),
			wantLaterOwnerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Empty(orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Empty(orderIds.OrderIds)
			},
		},
		{
			name:                  "pass - after match offer, reverse records of the Dym-Name are updated",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				// reverse record still linked to owner before transaction
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames := s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				// no reverse record for buyer (the later owner) before transaction
				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)
			},
			wantErr:        false,
			wantLaterOffer: nil,
			wantLaterDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      offer.Buyer,
				Controller: offer.Buyer,
				ExpireAt:   dymName.ExpireAt,
			},
			wantLaterModuleBalance: sdkmath.ZeroInt(),
			wantLaterOwnerBalance:  offer.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				// reverse record to later owner (buyer) are created after transaction
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames := s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				// reverse record to previous owner are removed after transaction
				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)
			},
		},
		{
			name:                  "pass - (negotiation) when price not match offer price, raise the counterparty offer price",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         offer.Id,
				AssetId:    offer.AssetId,
				AssetType:  dymnstypes.TypeName,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.AddAmount(sdk.NewInt(1))
					return &coin
				}(),
			},
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "pass - after put negotiation price, reverse records of the offer are preserved",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offer.Id}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offer.Id}, orderIds.OrderIds)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         offer.Id,
				AssetId:    offer.AssetId,
				AssetType:  dymnstypes.TypeName,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.AddAmount(sdk.NewInt(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(dymName.Name)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offer.Id}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offer.Id}, orderIds.OrderIds)
			},
		},
		{
			name:                  "pass - after put negotiation price, reverse records of the Dym-Name are preserved",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames := s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         offer.Id,
				AssetId:    offer.AssetId,
				AssetType:  dymnstypes.TypeName,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.AddAmount(sdk.NewInt(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames := s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Equal([]string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx, key)
				s.Require().Empty(dymNames.DymNames)
			},
		},
		{
			name:                  "fail - can NOT accept offer when trading Dym-Name is disabled",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Misc.EnableTradingName = false
					return moduleParams
				})
			},
			wantErr:                true,
			wantErrContains:        "trading of Dym-Name is disabled",
			wantLaterOffer:         offer,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: offer.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingDymName:        dymName,
			existingOffer:          nil,
			buyOrderId:             "101",
			owner:                  ownerA,
			minAccept:              minOfferPriceCoin,
			originalModuleBalance:  sdkmath.NewInt(1).Mul(priceMultiplier),
			originalOwnerBalance:   sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:                true,
			wantErrContains:        "Buy-Order: 101: not found",
			wantLaterOffer:         nil,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingDymName:        dymName,
			existingOffer:          offer,
			buyOrderId:             "10673264823",
			owner:                  ownerA,
			minAccept:              minOfferPriceCoin,
			originalModuleBalance:  sdkmath.NewInt(1).Mul(priceMultiplier),
			originalOwnerBalance:   sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:                true,
			wantErrContains:        "Buy-Order: 10673264823: not found",
			wantLaterOffer:         offer,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - Dym-Name not found",
			existingDymName:        nil,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			owner:                  ownerA,
			minAccept:              offer.OfferPrice,
			originalModuleBalance:  sdkmath.NewInt(1).Mul(priceMultiplier),
			originalOwnerBalance:   sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:                true,
			wantErrContains:        fmt.Sprintf("Dym-Name: %s: not found", offer.AssetId),
			wantLaterOffer:         offer,
			wantLaterDymName:       nil,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name: "fail - expired Dym-Name",
			existingDymName: func() *dymnstypes.DymName {
				return &dymnstypes.DymName{
					Name:       dymName.Name,
					Owner:      dymName.Owner,
					Controller: dymName.Controller,
					ExpireAt:   s.now.Unix() - 1,
				}
			}(),
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			owner:                  dymName.Owner,
			minAccept:              offer.OfferPrice,
			originalModuleBalance:  sdkmath.NewInt(1).Mul(priceMultiplier),
			originalOwnerBalance:   sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:                true,
			wantErrContains:        fmt.Sprintf("Dym-Name: %s: not found", offer.AssetId),
			wantLaterOffer:         offer,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not accept offer of Dym-Name owned by another",
			existingDymName:        sameDymNameButOwnedByAnother,
			existingOffer:          offer,
			buyOrderId:             offer.Id,
			owner:                  ownerA,
			minAccept:              offer.OfferPrice,
			originalModuleBalance:  sdkmath.NewInt(1).Mul(priceMultiplier),
			originalOwnerBalance:   sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:                true,
			wantErrContains:        "not the owner of the Dym-Name",
			wantLaterDymName:       sameDymNameButOwnedByAnother,
			wantLaterOffer:         offer,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name: "fail - can not accept offer if Dym-Name expiration less than grace period",
			existingDymName: func() *dymnstypes.DymName {
				return &dymnstypes.DymName{
					Name:       dymName.Name,
					Owner:      dymName.Owner,
					Controller: dymName.Owner,
					ExpireAt:   s.now.Add(time.Hour).Unix(),
				}
			}(),
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 ownerA,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:               true,
			wantErrContains:       "duration before Dym-Name expiry, prohibited to sell",
			wantLaterDymName: func() *dymnstypes.DymName {
				return &dymnstypes.DymName{
					Name:       dymName.Name,
					Owner:      dymName.Owner,
					Controller: dymName.Owner,
					ExpireAt:   s.now.Add(time.Hour).Unix(),
				}
			}(),
			wantLaterOffer:         offer,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:            "fail - can not accept own offer",
			existingDymName: dymName,
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         "101",
					AssetId:    dymName.Name,
					AssetType:  dymnstypes.TypeName,
					Buyer:      ownerA,
					OfferPrice: minOfferPriceCoin,
				}
			}(),
			buyOrderId:            "101",
			owner:                 ownerA,
			minAccept:             minOfferPriceCoin,
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:               true,
			wantErrContains:       "cannot accept own offer",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         "101",
					AssetId:    dymName.Name,
					AssetType:  dymnstypes.TypeName,
					Buyer:      ownerA,
					OfferPrice: minOfferPriceCoin,
				}
			}(),
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:            "fail - offer price denom != accept price denom",
			existingDymName: dymName,
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:        "101",
					AssetId:   dymName.Name,
					AssetType: dymnstypes.TypeName,
					Buyer:     buyerA,
					OfferPrice: sdk.Coin{
						Denom:  s.priceDenom(),
						Amount: sdk.NewInt(minOfferPrice).Mul(priceMultiplier),
					},
				}
			}(),
			buyOrderId: "101",
			owner:      ownerA,
			minAccept: sdk.Coin{
				Denom:  "u" + s.priceDenom(),
				Amount: sdk.NewInt(minOfferPrice).Mul(priceMultiplier),
			},
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:               true,
			wantErrContains:       "denom must be the same as the offer price",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.BuyOrder {
				// unchanged
				return &dymnstypes.BuyOrder{
					Id:        "101",
					AssetId:   dymName.Name,
					AssetType: dymnstypes.TypeName,
					Buyer:     buyerA,
					OfferPrice: sdk.Coin{
						Denom:  s.priceDenom(),
						Amount: sdk.NewInt(minOfferPrice).Mul(priceMultiplier),
					},
				}
			}(),
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:            "fail - accept price lower than offer price",
			existingDymName: dymName,
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         "101",
					AssetId:    dymName.Name,
					AssetType:  dymnstypes.TypeName,
					Buyer:      buyerA,
					OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				}
			}(),
			buyOrderId:            "101",
			owner:                 ownerA,
			minAccept:             minOfferPriceCoin,
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantErr:               true,
			wantErrContains:       "amount must be greater than or equals to the offer price",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         "101",
					AssetId:    dymName.Name,
					AssetType:  dymnstypes.TypeName,
					Buyer:      buyerA,
					OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				}
			}(),
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      1,
		},
		{
			name:                  "fail - prohibited to accept offer if a Sell-Order is active",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(
					s.ctx,
					s.newDymNameSellOrder(dymName.Name).
						WithMinPrice(1).
						WithExpiry(s.now.Add(time.Hour).Unix()).
						Build(),
				)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "must cancel the sell order first",
			wantLaterOffer:         offer,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: offer.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                  "fail - prohibited to accept offer if a Sell-Order is active, regardless the SO is expired",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(
					s.ctx,
					s.newDymNameSellOrder(dymName.Name).
						WithMinPrice(1).
						Expired().
						Build(),
				)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "must cancel the sell order first",
			wantLaterOffer:         offer,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: offer.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                  "pass - can negotiate when a Sell-Order is active",
			existingDymName:       dymName,
			existingOffer:         offer,
			buyOrderId:            offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			originalOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(
					s.ctx,
					s.newDymNameSellOrder(dymName.Name).
						WithMinPrice(1).
						Expired().
						Build(),
				)
				s.Require().NoError(err)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:         offer.Id,
				AssetId:    offer.AssetId,
				AssetType:  dymnstypes.TypeName,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.AddAmount(sdk.NewInt(1))
					return &coin
				}(),
			},
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: sdkmath.OneInt().Mul(priceMultiplier),
			wantLaterOwnerBalance:  sdkmath.NewInt(2).Mul(priceMultiplier),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.UseAnchorContext()

			if tt.originalModuleBalance.IsPositive() {
				s.mintToModuleAccount2(tt.originalModuleBalance)
			}

			if tt.originalOwnerBalance.IsPositive() {
				s.mintToAccount2(tt.owner, tt.originalOwnerBalance)
			}

			if tt.existingDymName != nil {
				s.setDymNameWithFunctionsAfter(*tt.existingDymName)
			}

			if tt.existingOffer != nil {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *tt.existingOffer)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, tt.existingOffer.AssetId, tt.existingOffer.AssetType, tt.existingOffer.Id)
				s.Require().NoError(err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).AcceptBuyOrder(s.ctx, &dymnstypes.MsgAcceptBuyOrder{
				OrderId:   tt.buyOrderId,
				Owner:     tt.owner,
				MinAccept: tt.minAccept,
			})

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.wantLaterOffer.Id)
					s.Require().NotNil(laterOffer)
					s.Require().Equal(*tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.buyOrderId)
					s.Require().Nil(laterOffer)
				}

				laterModuleBalance := s.moduleBalance2()
				s.Equal(tt.wantLaterModuleBalance.String(), laterModuleBalance.String())

				laterBuyerBalance := s.balance2(tt.owner)
				s.Equal(tt.wantLaterOwnerBalance.String(), laterBuyerBalance.String())

				s.Less(tt.wantMinConsumeGas, s.ctx.GasMeter().GasConsumed())

				if tt.wantLaterDymName != nil {
					laterDymName := s.dymNsKeeper.GetDymName(s.ctx, tt.wantLaterDymName.Name)
					s.Require().NotNil(laterDymName)
					s.Require().Equal(*tt.wantLaterDymName, *laterDymName)
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.NotNil(resp)
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func (s *KeeperTestSuite) Test_msgServer_AcceptBuyOrder_Type_Alias() {
	const minOfferPrice = 5

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	minOfferPriceCoin := sdk.NewCoin(s.priceDenom(), sdk.NewInt(minOfferPrice).Mul(priceMultiplier))

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	anotherAcc := testAddr(3)

	rollApp_One_By1_SingleAlias := *newRollApp("rollapp_1-1").
		WithOwner(creator_1_asOwner).
		WithAlias("one1")
	rollApp_Two_By2_SingleAlias := *newRollApp("rollapp_2-2").
		WithOwner(creator_2_asBuyer).
		WithAlias("two1")
	rollApp_Three_By1_MultipleAliases := *newRollApp("rollapp_3-1").
		WithOwner(creator_1_asOwner).
		WithAlias("three1").WithAlias("three2")
	rollApp_Four_By2_MultipleAliases := *newRollApp("rollapp_4-2").
		WithOwner(creator_2_asBuyer).
		WithAlias("four1").WithAlias("four2").WithAlias("four3")

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.MinOfferPrice = minOfferPriceCoin.Amount
		// force enable trading
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		return moduleParams
	})
	s.MakeAnchorContext()

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).AcceptBuyOrder(s.ctx, &dymnstypes.MsgAcceptBuyOrder{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	offerAliasOfRollAppOne := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 1),
		AssetId:    rollApp_One_By1_SingleAlias.aliases[0],
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{rollApp_Two_By2_SingleAlias.rollAppId},
		Buyer:      rollApp_Two_By2_SingleAlias.owner,
		OfferPrice: minOfferPriceCoin,
	}

	offerNonExistingAlias := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 2),
		AssetId:    "nah",
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{rollApp_Two_By2_SingleAlias.rollAppId},
		Buyer:      rollApp_Two_By2_SingleAlias.owner,
		OfferPrice: minOfferPriceCoin,
	}

	offerAliasForNonExistingRollApp := &dymnstypes.BuyOrder{
		Id:         dymnstypes.CreateBuyOrderId(dymnstypes.TypeAlias, 1),
		AssetId:    rollApp_One_By1_SingleAlias.aliases[0],
		AssetType:  dymnstypes.TypeAlias,
		Params:     []string{"nah_0-0"},
		Buyer:      creator_2_asBuyer,
		OfferPrice: minOfferPriceCoin,
	}

	tests := []struct {
		name                   string
		existingRollApps       []rollapp
		existingOffer          *dymnstypes.BuyOrder
		buyOrderId             string
		owner                  string
		minAccept              sdk.Coin
		originalModuleBalance  sdkmath.Int
		originalOwnerBalance   sdkmath.Int
		preRunSetupFunc        func(s *KeeperTestSuite)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOrder
		wantLaterRollApps      []rollapp
		wantLaterModuleBalance sdkmath.Int
		wantLaterOwnerBalance  sdkmath.Int
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(s *KeeperTestSuite)
	}{
		{
			name:                  "pass - can accept offer (match)",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer:        nil,
			wantLaterRollApps: []rollapp{
				{
					rollAppId: rollApp_One_By1_SingleAlias.rollAppId,
					aliases:   []string{},
				},
				{
					rollAppId: rollApp_Two_By2_SingleAlias.rollAppId,
					aliases:   append(rollApp_Two_By2_SingleAlias.aliases, offerAliasOfRollAppOne.AssetId),
				},
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "pass - after match offer, reverse records of the offer are removed",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.AliasToBuyOrderIdsRvlKey(offerAliasOfRollAppOne.AssetId)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offerAliasOfRollAppOne.Id}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offerAliasOfRollAppOne.Id}, orderIds.OrderIds)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.DymNameToBuyOrderIdsRvlKey(offerAliasOfRollAppOne.AssetId)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Empty(orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Empty(orderIds.OrderIds)
			},
		},
		{
			name:                  "pass - after match offer, linking between RollApps and the alias are updated",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.requireAlias(offerAliasOfRollAppOne.AssetId).
					LinkedToRollApp(rollApp_One_By1_SingleAlias.rollAppId)
			},
			wantErr:        false,
			wantLaterOffer: nil,
			wantLaterRollApps: []rollapp{
				{
					rollAppId: rollApp_One_By1_SingleAlias.rollAppId,
					aliases:   []string{},
				},
				{
					rollAppId: rollApp_Two_By2_SingleAlias.rollAppId,
					aliases:   append(rollApp_Two_By2_SingleAlias.aliases, offerAliasOfRollAppOne.AssetId),
				},
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				s.requireAlias(offerAliasOfRollAppOne.AssetId).
					LinkedToRollApp(rollApp_Two_By2_SingleAlias.rollAppId) // changed

				s.requireRollApp(rollApp_One_By1_SingleAlias.rollAppId).
					HasNoAlias() // link removed from previous RollApp
			},
		},
		{
			name:                  "pass - (negotiation) when price not match offer price, raise the counterparty offer price",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     offerAliasOfRollAppOne.Id,
				AssetId:                offerAliasOfRollAppOne.AssetId,
				AssetType:              offerAliasOfRollAppOne.AssetType,
				Params:                 offerAliasOfRollAppOne.Params,
				Buyer:                  offerAliasOfRollAppOne.Buyer,
				OfferPrice:             offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: uptr.To(offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1))),
			},
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "pass - after put negotiation price, reverse records of the offer are preserved",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				key := dymnstypes.AliasToBuyOrderIdsRvlKey(offerAliasOfRollAppOne.AssetId)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offerAliasOfRollAppOne.Id}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offerAliasOfRollAppOne.Id}, orderIds.OrderIds)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     offerAliasOfRollAppOne.Id,
				AssetId:                offerAliasOfRollAppOne.AssetId,
				AssetType:              offerAliasOfRollAppOne.AssetType,
				Params:                 offerAliasOfRollAppOne.Params,
				Buyer:                  offerAliasOfRollAppOne.Buyer,
				OfferPrice:             offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: uptr.To(offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1))),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				// the same as before

				key := dymnstypes.AliasToBuyOrderIdsRvlKey(offerAliasOfRollAppOne.AssetId)
				orderIds := s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offerAliasOfRollAppOne.Id}, orderIds.OrderIds)

				key = dymnstypes.BuyerToOrderIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				orderIds = s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(s.ctx, key)
				s.Require().Equal([]string{offerAliasOfRollAppOne.Id}, orderIds.OrderIds)
			},
		},
		{
			name:                  "pass - after put negotiation price, original linking between RollApp and alias are preserved",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.requireAlias(offerAliasOfRollAppOne.AssetId).LinkedToRollApp(rollApp_One_By1_SingleAlias.rollAppId)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     offerAliasOfRollAppOne.Id,
				AssetId:                offerAliasOfRollAppOne.AssetId,
				AssetType:              offerAliasOfRollAppOne.AssetType,
				Params:                 offerAliasOfRollAppOne.Params,
				Buyer:                  offerAliasOfRollAppOne.Buyer,
				OfferPrice:             offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: uptr.To(offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1))),
			},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
			afterTestFunc: func(s *KeeperTestSuite) {
				// unchanged
				s.requireAlias(offerAliasOfRollAppOne.AssetId).LinkedToRollApp(rollApp_One_By1_SingleAlias.rollAppId)
			},
		},
		{
			name: "fail - not accept offer if alias presents in params",
			existingRollApps: []rollapp{
				rollApp_One_By1_SingleAlias,
				rollApp_Two_By2_SingleAlias,
			},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
						{
							ChainId: "some-chain",
							Aliases: []string{offerAliasOfRollAppOne.AssetId},
						},
					}
					return moduleParams
				})
			},
			wantErr:         true,
			wantErrContains: "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantLaterOffer:  offerAliasOfRollAppOne,
			wantLaterRollApps: []rollapp{
				rollApp_One_By1_SingleAlias,
				rollApp_Two_By2_SingleAlias,
			},
			wantLaterModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                  "fail - can NOT accept offer when trading Alias is disabled",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.Misc.EnableTradingAlias = false
					return moduleParams
				})
			},
			wantErr:                true,
			wantErrContains:        "trading of Alias is disabled",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          nil,
			buyOrderId:             "201",
			owner:                  rollApp_One_By1_SingleAlias.owner,
			minAccept:              minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalOwnerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "Buy-Order: 201: not found",
			wantLaterOffer:         nil,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             "20673264823",
			owner:                  rollApp_One_By1_SingleAlias.owner,
			minAccept:              minOfferPriceCoin,
			originalModuleBalance:  sdk.NewInt(1),
			originalOwnerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "Buy-Order: 20673264823: not found",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - Alias not found",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerNonExistingAlias, // offer non-existing alias
			buyOrderId:             offerNonExistingAlias.Id,
			owner:                  rollApp_One_By1_SingleAlias.owner,
			minAccept:              offerNonExistingAlias.OfferPrice,
			originalModuleBalance:  sdk.NewInt(1),
			originalOwnerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "alias is not in-used",
			wantLaterOffer:         offerNonExistingAlias,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - destination RollApp not exists",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerAliasForNonExistingRollApp, // offer for non-existing RollApp
			buyOrderId:             offerAliasForNonExistingRollApp.Id,
			owner:                  rollApp_One_By1_SingleAlias.owner,
			minAccept:              offerAliasForNonExistingRollApp.OfferPrice,
			originalModuleBalance:  sdk.NewInt(1),
			originalOwnerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "invalid destination Roll-App ID",
			wantLaterOffer:         offerAliasForNonExistingRollApp,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not accept offer of Alias owned by another",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerAliasOfRollAppOne,
			buyOrderId:             offerAliasOfRollAppOne.Id,
			owner:                  anotherAcc.bech32(),
			minAccept:              offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance:  sdk.NewInt(1),
			originalOwnerBalance:   sdk.NewInt(2),
			wantErr:                true,
			wantErrContains:        "not the owner of the RollApp",
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - can not accept own offer",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         offerAliasOfRollAppOne.Id,
					AssetId:    offerAliasOfRollAppOne.AssetId,
					AssetType:  offerAliasOfRollAppOne.AssetType,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      creator_1_asOwner,
					OfferPrice: minOfferPriceCoin,
				}
			}(),
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 creator_1_asOwner,
			minAccept:             minOfferPriceCoin,
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			wantErr:               true,
			wantErrContains:       "cannot accept own offer",
			wantLaterRollApps:     []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         offerAliasOfRollAppOne.Id,
					AssetId:    offerAliasOfRollAppOne.AssetId,
					AssetType:  offerAliasOfRollAppOne.AssetType,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      creator_1_asOwner,
					OfferPrice: minOfferPriceCoin,
				}
			}(),
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - offer price denom != accept price denom",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:        offerAliasOfRollAppOne.Id,
					AssetId:   offerAliasOfRollAppOne.AssetId,
					AssetType: offerAliasOfRollAppOne.AssetType,
					Params:    offerAliasOfRollAppOne.Params,
					Buyer:     offerAliasOfRollAppOne.Buyer,
					OfferPrice: sdk.Coin{
						Denom:  s.priceDenom(),
						Amount: minOfferPriceCoin.Amount,
					},
				}
			}(),
			buyOrderId: offerAliasOfRollAppOne.Id,
			owner:      rollApp_One_By1_SingleAlias.owner,
			minAccept: sdk.Coin{
				Denom:  "u" + s.priceDenom(),
				Amount: minOfferPriceCoin.Amount,
			},
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			wantErr:               true,
			wantErrContains:       "denom must be the same as the offer price",
			wantLaterRollApps:     []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer: func() *dymnstypes.BuyOrder {
				// unchanged
				return &dymnstypes.BuyOrder{
					Id:        offerAliasOfRollAppOne.Id,
					AssetId:   offerAliasOfRollAppOne.AssetId,
					AssetType: offerAliasOfRollAppOne.AssetType,
					Params:    offerAliasOfRollAppOne.Params,
					Buyer:     offerAliasOfRollAppOne.Buyer,
					OfferPrice: sdk.Coin{
						Denom:  s.priceDenom(),
						Amount: minOfferPriceCoin.Amount,
					},
				}
			}(),
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - accept price lower than offer price",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         offerAliasOfRollAppOne.Id,
					AssetId:    offerAliasOfRollAppOne.AssetId,
					AssetType:  offerAliasOfRollAppOne.AssetType,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      offerAliasOfRollAppOne.Buyer,
					OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				}
			}(),
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             minOfferPriceCoin,
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			wantErr:               true,
			wantErrContains:       "amount must be greater than or equals to the offer price",
			wantLaterRollApps:     []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         offerAliasOfRollAppOne.Id,
					AssetId:    offerAliasOfRollAppOne.AssetId,
					AssetType:  offerAliasOfRollAppOne.AssetType,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      offerAliasOfRollAppOne.Buyer,
					OfferPrice: minOfferPriceCoin.AddAmount(sdk.NewInt(2)),
				}
			}(),
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      1,
		},
		{
			name:             "pass - accept offer transfer alias from One-Alias-RollApp to Multiple-Alias-RollApp",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Four_By2_MultipleAliases},
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         "201",
					AssetId:    rollApp_One_By1_SingleAlias.aliases[0],
					AssetType:  dymnstypes.TypeAlias,
					Params:     []string{rollApp_Four_By2_MultipleAliases.rollAppId},
					Buyer:      creator_2_asBuyer,
					OfferPrice: minOfferPriceCoin,
				}
			}(),
			buyOrderId:            "201",
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer:        nil,
			wantLaterRollApps: []rollapp{
				{
					rollAppId: rollApp_One_By1_SingleAlias.rollAppId,
					aliases:   []string{},
				},
				{
					rollAppId: rollApp_Four_By2_MultipleAliases.rollAppId,
					aliases:   append(rollApp_Four_By2_MultipleAliases.aliases, offerAliasOfRollAppOne.AssetId),
				},
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:             "pass - accept offer transfer alias from Multiple-Alias-RollApp to One-Alias-RollApp",
			existingRollApps: []rollapp{rollApp_Three_By1_MultipleAliases, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOrder {
				return &dymnstypes.BuyOrder{
					Id:         "201",
					AssetId:    rollApp_Three_By1_MultipleAliases.aliases[0],
					AssetType:  dymnstypes.TypeAlias,
					Params:     []string{rollApp_Two_By2_SingleAlias.rollAppId},
					Buyer:      creator_2_asBuyer,
					OfferPrice: minOfferPriceCoin,
				}
			}(),
			buyOrderId:            "201",
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             minOfferPriceCoin,
			originalModuleBalance: minOfferPriceCoin.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer:        nil,
			wantLaterRollApps: []rollapp{
				{
					rollAppId: rollApp_Three_By1_MultipleAliases.rollAppId,
					aliases:   rollApp_Three_By1_MultipleAliases.aliases[1:],
				},
				{
					rollAppId: rollApp_Two_By2_SingleAlias.rollAppId,
					aliases:   append(rollApp_Two_By2_SingleAlias.aliases, rollApp_Three_By1_MultipleAliases.aliases[0]),
				},
			},
			wantLaterModuleBalance: sdk.NewInt(0),
			wantLaterOwnerBalance:  minOfferPriceCoin.Amount,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
		{
			name:                  "fail - prohibit to accept offer when a Sell-Order is active",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(
					s.ctx,
					s.newAliasSellOrder(offerAliasOfRollAppOne.AssetId).
						WithExpiry(s.now.Add(time.Hour).Unix()).
						WithMinPrice(1).
						Build(),
				)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "must cancel the sell order first",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                  "fail - prohibit to accept offer when a Sell-Order is active, regardless the SO is expired",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			originalOwnerBalance:  sdk.NewInt(0),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(
					s.ctx,
					s.newAliasSellOrder(offerAliasOfRollAppOne.AssetId).
						Expired().
						WithMinPrice(1).
						Build(),
				)
				s.Require().NoError(err)
			},
			wantErr:                true,
			wantErrContains:        "must cancel the sell order first",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount,
			wantLaterOwnerBalance:  sdk.NewInt(0),
			wantMinConsumeGas:      1,
		},
		{
			name:                  "pass - can negotiate when a Sell-Order is active",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			buyOrderId:            offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.owner,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1)),
			originalModuleBalance: sdk.NewInt(1),
			originalOwnerBalance:  sdk.NewInt(2),
			preRunSetupFunc: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.SetSellOrder(
					s.ctx,
					s.newAliasSellOrder(offerAliasOfRollAppOne.AssetId).
						Expired().
						WithMinPrice(1).
						Build(),
				)
				s.Require().NoError(err)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOrder{
				Id:                     offerAliasOfRollAppOne.Id,
				AssetId:                offerAliasOfRollAppOne.AssetId,
				AssetType:              offerAliasOfRollAppOne.AssetType,
				Params:                 offerAliasOfRollAppOne.Params,
				Buyer:                  offerAliasOfRollAppOne.Buyer,
				OfferPrice:             offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: uptr.To(offerAliasOfRollAppOne.OfferPrice.AddAmount(sdk.NewInt(1))),
			},
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: sdk.NewInt(1),
			wantLaterOwnerBalance:  sdk.NewInt(2),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOrder,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.UseAnchorContext()

			if tt.originalModuleBalance.IsPositive() {
				s.mintToModuleAccount2(tt.originalModuleBalance)
			}

			if tt.originalOwnerBalance.IsPositive() {
				s.mintToAccount2(tt.owner, tt.originalOwnerBalance)
			}

			for _, rollApp := range tt.existingRollApps {
				s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
					RollappId: rollApp.rollAppId,
					Owner:     rollApp.owner,
				})
				for _, alias := range rollApp.aliases {
					err := s.dymNsKeeper.SetAliasForRollAppId(s.ctx, rollApp.rollAppId, alias)
					s.Require().NoError(err)
				}
			}

			if tt.existingOffer != nil {
				err := s.dymNsKeeper.SetBuyOrder(s.ctx, *tt.existingOffer)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingBuyerToBuyOrderRecord(s.ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				s.Require().NoError(err)

				err = s.dymNsKeeper.AddReverseMappingAssetIdToBuyOrder(s.ctx, tt.existingOffer.AssetId, tt.existingOffer.AssetType, tt.existingOffer.Id)
				s.Require().NoError(err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).AcceptBuyOrder(s.ctx, &dymnstypes.MsgAcceptBuyOrder{
				OrderId:   tt.buyOrderId,
				Owner:     tt.owner,
				MinAccept: tt.minAccept,
			})

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.wantLaterOffer.Id)
					s.Require().NotNil(laterOffer)
					s.Require().Equal(*tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := s.dymNsKeeper.GetBuyOrder(s.ctx, tt.buyOrderId)
					s.Require().Nil(laterOffer)
				}

				laterModuleBalance := s.moduleBalance2()
				s.Require().Equal(tt.wantLaterModuleBalance.String(), laterModuleBalance.String())

				laterBuyerBalance := s.balance2(tt.owner)
				s.Require().Equal(tt.wantLaterOwnerBalance.String(), laterBuyerBalance.String())

				s.Less(tt.wantMinConsumeGas, s.ctx.GasMeter().GasConsumed())

				for _, wantLaterRollApp := range tt.wantLaterRollApps {
					rollApp, found := s.rollAppKeeper.GetRollapp(s.ctx, wantLaterRollApp.rollAppId)
					s.Require().True(found)
					if len(wantLaterRollApp.aliases) == 0 {
						s.requireRollApp(rollApp.RollappId).HasNoAlias()
					} else {
						s.requireRollApp(rollApp.RollappId).HasAlias(wantLaterRollApp.aliases...)
					}
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(s)
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(resp)
		})
	}
}
