package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_AcceptBuyOrder_Type_DymName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5
	const daysProhibitSell = 30

	buyerA := testAddr(1).bech32()
	ownerA := testAddr(2).bech32()
	anotherOwnerA := testAddr(3).bech32()

	setupTest := func() (dymnskeeper.Keeper, dymnstypes.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.PriceDenom = denom
		moduleParams.Price.MinOfferPrice = sdk.NewInt(minOfferPrice)
		// misc
		moduleParams.Misc.ProhibitSellDuration = daysProhibitSell * 24 * time.Hour
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).AcceptBuyOrder(ctx, &dymnstypes.MsgAcceptBuyOrder{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	dymName := &dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Add(daysProhibitSell * 24 * time.Hour).Add(time.Second).Unix(),
	}

	sameDymNameButOwnedByAnother := &dymnstypes.DymName{
		Name:       dymName.Name,
		Owner:      anotherOwnerA,
		Controller: anotherOwnerA,
		ExpireAt:   dymName.ExpireAt,
	}

	offer := &dymnstypes.BuyOffer{
		Id:         "101",
		GoodsId:    dymName.Name,
		Type:       dymnstypes.NameOrder,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	tests := []struct {
		name                   string
		existingDymName        *dymnstypes.DymName
		existingOffer          *dymnstypes.BuyOffer
		offerId                string
		owner                  string
		minAccept              sdk.Coin
		originalModuleBalance  int64
		originalOwnerBalance   int64
		preRunSetupFunc        func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOffer
		wantLaterDymName       *dymnstypes.DymName
		wantLaterModuleBalance int64
		wantLaterOwnerBalance  int64
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                  "pass - can accept offer (match)",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer:        nil,
			wantLaterDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      offer.Buyer,
				Controller: offer.Buyer,
				ExpireAt:   dymName.ExpireAt,
			},
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
		},
		{
			name:                  "pass - after match offer, reverse records of the offer are removed",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(dymName.Name)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)
			},
			wantErr:        false,
			wantLaterOffer: nil,
			wantLaterDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      offer.Buyer,
				Controller: offer.Buyer,
				ExpireAt:   dymName.ExpireAt,
			},
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(dymName.Name)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Empty(t, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Empty(t, offerIds.OfferIds)
			},
		},
		{
			name:                  "pass - after match offer, reverse records of the Dym-Name are updated",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				// reverse record still linked to owner before transaction
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames := dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				// no reverse record for buyer (the later owner) before transaction
				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)
			},
			wantErr:        false,
			wantLaterOffer: nil,
			wantLaterDymName: &dymnstypes.DymName{
				Name:       dymName.Name,
				Owner:      offer.Buyer,
				Controller: offer.Buyer,
				ExpireAt:   dymName.ExpireAt,
			},
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offer.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				// reverse record to later owner (buyer) are created after transaction
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames := dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				// reverse record to previous owner are removed after transaction
				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)
			},
		},
		{
			name:                  "pass - (negotiation) when price not match offer price, raise the counterparty offer price",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.Add(dymnsutils.TestCoin(1)),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer: &dymnstypes.BuyOffer{
				Id:         offer.Id,
				GoodsId:    offer.GoodsId,
				Type:       dymnstypes.NameOrder,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
		},
		{
			name:                  "pass - after put negotiation price, reverse records of the offer are preserved",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.Add(dymnsutils.TestCoin(1)),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(dymName.Name)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOffer{
				Id:         offer.Id,
				GoodsId:    offer.GoodsId,
				Type:       dymnstypes.NameOrder,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(dymName.Name)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)
			},
		},
		{
			name:                  "pass - after put negotiation price, reverse records of the Dym-Name are preserved",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice.Add(dymnsutils.TestCoin(1)),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames := dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOffer{
				Id:         offer.Id,
				GoodsId:    offer.GoodsId,
				Type:       dymnstypes.NameOrder,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(dymName.Owner)
				dymNames := dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(dymName.Owner)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(dymName.Owner))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Equal(t, []string{dymName.Name}, dymNames.DymNames)

				key = dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(offer.Buyer)
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(offer.Buyer)))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)

				key = dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx, key)
				require.Empty(t, dymNames.DymNames)
			},
		},
		{
			name:                  "fail - can NOT accept offer when trading Dym-Name is disabled",
			existingDymName:       dymName,
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 dymName.Owner,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: offer.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Misc.EnableTradingName = false
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                true,
			wantErrContains:        "trading of Dym-Name is disabled",
			wantLaterOffer:         offer,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: offer.OfferPrice.Amount.Int64(),
			wantLaterOwnerBalance:  0,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingDymName:        dymName,
			existingOffer:          nil,
			offerId:                "101",
			owner:                  ownerA,
			minAccept:              dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order: 101: not found",
			wantLaterOffer:         nil,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingDymName:        dymName,
			existingOffer:          offer,
			offerId:                "10673264823",
			owner:                  ownerA,
			minAccept:              dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order: 10673264823: not found",
			wantLaterOffer:         offer,
			wantLaterDymName:       dymName,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - Dym-Name not found",
			existingDymName:        nil,
			existingOffer:          offer,
			offerId:                offer.Id,
			owner:                  ownerA,
			minAccept:              offer.OfferPrice,
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        fmt.Sprintf("Dym-Name: %s: not found", offer.GoodsId),
			wantLaterOffer:         offer,
			wantLaterDymName:       nil,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name: "fail - expired Dym-Name",
			existingDymName: func() *dymnstypes.DymName {
				return &dymnstypes.DymName{
					Name:       dymName.Name,
					Owner:      dymName.Owner,
					Controller: dymName.Controller,
					ExpireAt:   now.Unix() - 1,
				}
			}(),
			existingOffer:          offer,
			offerId:                offer.Id,
			owner:                  dymName.Owner,
			minAccept:              offer.OfferPrice,
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        fmt.Sprintf("Dym-Name: %s: not found", offer.GoodsId),
			wantLaterOffer:         offer,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not accept offer of Dym-Name owned by another",
			existingDymName:        sameDymNameButOwnedByAnother,
			existingOffer:          offer,
			offerId:                offer.Id,
			owner:                  ownerA,
			minAccept:              offer.OfferPrice,
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "not the owner of the Dym-Name",
			wantLaterDymName:       sameDymNameButOwnedByAnother,
			wantLaterOffer:         offer,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name: "fail - can not accept offer if Dym-Name expiration less than grace period",
			existingDymName: func() *dymnstypes.DymName {
				return &dymnstypes.DymName{
					Name:       dymName.Name,
					Owner:      dymName.Owner,
					Controller: dymName.Owner,
					ExpireAt:   now.Unix() + 1,
				}
			}(),
			existingOffer:         offer,
			offerId:               offer.Id,
			owner:                 ownerA,
			minAccept:             offer.OfferPrice,
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "duration before Dym-Name expiry, prohibited to sell",
			wantLaterDymName: func() *dymnstypes.DymName {
				return &dymnstypes.DymName{
					Name:       dymName.Name,
					Owner:      dymName.Owner,
					Controller: dymName.Owner,
					ExpireAt:   now.Unix() + 1,
				}
			}(),
			wantLaterOffer:         offer,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:            "fail - can not accept own offer",
			existingDymName: dymName,
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         "101",
					GoodsId:    dymName.Name,
					Type:       dymnstypes.NameOrder,
					Buyer:      ownerA,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			offerId:               "101",
			owner:                 ownerA,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "cannot accept own offer",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         "101",
					GoodsId:    dymName.Name,
					Type:       dymnstypes.NameOrder,
					Buyer:      ownerA,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:            "fail - offer price denom != accept price denom",
			existingDymName: dymName,
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:      "101",
					GoodsId: dymName.Name,
					Type:    dymnstypes.NameOrder,
					Buyer:   buyerA,
					OfferPrice: sdk.Coin{
						Denom:  denom,
						Amount: sdk.NewInt(minOfferPrice),
					},
				}
			}(),
			offerId: "101",
			owner:   ownerA,
			minAccept: sdk.Coin{
				Denom:  "u" + denom,
				Amount: sdk.NewInt(minOfferPrice),
			},
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "denom must be the same as the offer price",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:      "101",
					GoodsId: dymName.Name,
					Type:    dymnstypes.NameOrder,
					Buyer:   buyerA,
					OfferPrice: sdk.Coin{
						Denom:  denom,
						Amount: sdk.NewInt(minOfferPrice),
					},
				}
			}(),
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:            "fail - accept price lower than offer price",
			existingDymName: dymName,
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         "101",
					GoodsId:    dymName.Name,
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
				}
			}(),
			offerId:               "101",
			owner:                 ownerA,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "amount must be greater than or equals to the offer price",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         "101",
					GoodsId:    dymName.Name,
					Type:       dymnstypes.NameOrder,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
				}
			}(),
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, bk, ctx := setupTest()

			if tt.originalModuleBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalModuleBalance),
				)
				require.NoError(t, err)
			}

			if tt.originalOwnerBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalOwnerBalance),
				)
				require.NoError(t, err)

				err = bk.SendCoinsFromModuleToAccount(
					ctx,
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(tt.owner),
					dymnsutils.TestCoins(tt.originalOwnerBalance),
				)
				require.NoError(t, err)
			}

			if tt.existingDymName != nil {
				setDymNameWithFunctionsAfter(ctx, *tt.existingDymName, t, dk)
			}

			if tt.existingOffer != nil {
				err := dk.SetBuyOffer(ctx, *tt.existingOffer)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingGoodsIdToBuyOffer(ctx, tt.existingOffer.GoodsId, tt.existingOffer.Type, tt.existingOffer.Id)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).AcceptBuyOrder(ctx, &dymnstypes.MsgAcceptBuyOrder{
				OfferId:   tt.offerId,
				Owner:     tt.owner,
				MinAccept: tt.minAccept,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetBuyOffer(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := dk.GetBuyOffer(ctx, tt.offerId)
					require.Nil(t, laterOffer)
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, denom)
				require.Equal(t, tt.wantLaterModuleBalance, laterModuleBalance.Amount.Int64())

				laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.owner), denom)
				require.Equal(t, tt.wantLaterOwnerBalance, laterBuyerBalance.Amount.Int64())

				require.Less(t, tt.wantMinConsumeGas, ctx.GasMeter().GasConsumed())

				if tt.wantLaterDymName != nil {
					laterDymName := dk.GetDymName(ctx, tt.wantLaterDymName.Name)
					require.NotNil(t, laterDymName)
					require.Equal(t, *tt.wantLaterDymName, *laterDymName)
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(ctx, dk)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

//goland:noinspection GoSnakeCaseUsage
func Test_msgServer_AcceptBuyOrder_Type_Alias(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5

	creator_1_asOwner := testAddr(1).bech32()
	creator_2_asBuyer := testAddr(2).bech32()
	anotherAcc := testAddr(3)

	type rollapp struct {
		rollAppId string
		creator   string
		aliases   []string
	}

	rollApp_One_By1_SingleAlias := rollapp{
		rollAppId: "rollapp_1-1",
		creator:   creator_1_asOwner,
		aliases:   []string{"one1"},
	}
	rollApp_Two_By2_SingleAlias := rollapp{
		rollAppId: "rollapp_2-2",
		creator:   creator_2_asBuyer,
		aliases:   []string{"two1"},
	}
	rollApp_Three_By1_MultipleAliases := rollapp{
		rollAppId: "rollapp_3-1",
		creator:   creator_1_asOwner,
		aliases:   []string{"three1", "three2"},
	}
	rollApp_Four_By2_MultipleAliases := rollapp{
		rollAppId: "rollapp_4-2",
		creator:   creator_2_asBuyer,
		aliases:   []string{"four1", "four2", "four3"},
	}

	setupTest := func() (sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper, dymnstypes.BankKeeper) {
		dk, bk, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.PriceDenom = denom
		moduleParams.Price.MinOfferPrice = sdk.NewInt(minOfferPrice)
		// misc
		moduleParams.Misc.EnableTradingName = true
		moduleParams.Misc.EnableTradingAlias = true
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return ctx, dk, rk, bk
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		ctx, dk, _, _ := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).AcceptBuyOrder(ctx, &dymnstypes.MsgAcceptBuyOrder{})
			return err
		}, gerrc.ErrInvalidArgument.Error())
	})

	offerAliasOfRollAppOne := &dymnstypes.BuyOffer{
		Id:         dymnstypes.CreateBuyOfferId(dymnstypes.AliasOrder, 1),
		GoodsId:    rollApp_One_By1_SingleAlias.aliases[0],
		Type:       dymnstypes.AliasOrder,
		Params:     []string{rollApp_Two_By2_SingleAlias.rollAppId},
		Buyer:      rollApp_Two_By2_SingleAlias.creator,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	offerNonExistingAlias := &dymnstypes.BuyOffer{
		Id:         dymnstypes.CreateBuyOfferId(dymnstypes.AliasOrder, 2),
		GoodsId:    "nah",
		Type:       dymnstypes.AliasOrder,
		Params:     []string{rollApp_Two_By2_SingleAlias.rollAppId},
		Buyer:      rollApp_Two_By2_SingleAlias.creator,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	offerAliasForNonExistingRollApp := &dymnstypes.BuyOffer{
		Id:         dymnstypes.CreateBuyOfferId(dymnstypes.AliasOrder, 1),
		GoodsId:    rollApp_One_By1_SingleAlias.aliases[0],
		Type:       dymnstypes.AliasOrder,
		Params:     []string{"nah_0-0"},
		Buyer:      creator_2_asBuyer,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	tests := []struct {
		name                   string
		existingRollApps       []rollapp
		existingOffer          *dymnstypes.BuyOffer
		offerId                string
		owner                  string
		minAccept              sdk.Coin
		originalModuleBalance  int64
		originalOwnerBalance   int64
		preRunSetupFunc        func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.BuyOffer
		wantLaterRollApps      []rollapp
		wantLaterModuleBalance int64
		wantLaterOwnerBalance  int64
		wantMinConsumeGas      sdk.Gas
		afterTestFunc          func(ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name:                  "pass - can accept offer (match)",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
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
					aliases:   append(rollApp_Two_By2_SingleAlias.aliases, offerAliasOfRollAppOne.GoodsId),
				},
			},
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
		},
		{
			name:                  "pass - after match offer, reverse records of the offer are removed",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.AliasToOfferIdsRvlKey(offerAliasOfRollAppOne.GoodsId)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offerAliasOfRollAppOne.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offerAliasOfRollAppOne.Id}, offerIds.OfferIds)
			},
			wantErr:                false,
			wantLaterOffer:         nil,
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(offerAliasOfRollAppOne.GoodsId)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Empty(t, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Empty(t, offerIds.OfferIds)
			},
		},
		{
			name:                  "pass - after match offer, linking between RollApps and the alias are updated",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				requireAliasLinkedToRollApp(
					offerAliasOfRollAppOne.GoodsId,
					rollApp_One_By1_SingleAlias.rollAppId,
					t, ctx, dk,
				)
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
					aliases:   append(rollApp_Two_By2_SingleAlias.aliases, offerAliasOfRollAppOne.GoodsId),
				},
			},
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				requireAliasLinkedToRollApp(
					offerAliasOfRollAppOne.GoodsId,
					rollApp_Two_By2_SingleAlias.rollAppId, // changed
					t, ctx, dk,
				)

				requireRollAppHasNoAlias(
					rollApp_One_By1_SingleAlias.rollAppId, // linking removed from previous RollApp
					t, ctx, dk,
				)
			},
		},
		{
			name:                  "pass - (negotiation) when price not match offer price, raise the counterparty offer price",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.Add(dymnsutils.TestCoin(1)),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			preRunSetupFunc:       nil,
			wantErr:               false,
			wantLaterOffer: &dymnstypes.BuyOffer{
				Id:         offerAliasOfRollAppOne.Id,
				GoodsId:    offerAliasOfRollAppOne.GoodsId,
				Type:       offerAliasOfRollAppOne.Type,
				Params:     offerAliasOfRollAppOne.Params,
				Buyer:      offerAliasOfRollAppOne.Buyer,
				OfferPrice: offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offerAliasOfRollAppOne.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
		},
		{
			name:                  "pass - after put negotiation price, reverse records of the offer are preserved",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.Add(dymnsutils.TestCoin(1)),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.AliasToOfferIdsRvlKey(offerAliasOfRollAppOne.GoodsId)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offerAliasOfRollAppOne.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offerAliasOfRollAppOne.Id}, offerIds.OfferIds)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOffer{
				Id:         offerAliasOfRollAppOne.Id,
				GoodsId:    offerAliasOfRollAppOne.GoodsId,
				Type:       offerAliasOfRollAppOne.Type,
				Params:     offerAliasOfRollAppOne.Params,
				Buyer:      offerAliasOfRollAppOne.Buyer,
				OfferPrice: offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offerAliasOfRollAppOne.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				// the same as before

				key := dymnstypes.AliasToOfferIdsRvlKey(offerAliasOfRollAppOne.GoodsId)
				offerIds := dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offerAliasOfRollAppOne.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offerAliasOfRollAppOne.Buyer))
				offerIds = dk.GenericGetReverseLookupBuyOfferIdsRecord(ctx, key)
				require.Equal(t, []string{offerAliasOfRollAppOne.Id}, offerIds.OfferIds)
			},
		},
		{
			name:                  "pass - after put negotiation price, original linking between RollApp and alias are preserved",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice.Add(dymnsutils.TestCoin(1)),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				requireAliasLinkedToRollApp(
					offerAliasOfRollAppOne.GoodsId, rollApp_One_By1_SingleAlias.rollAppId,
					t, ctx, dk,
				)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.BuyOffer{
				Id:         offerAliasOfRollAppOne.Id,
				GoodsId:    offerAliasOfRollAppOne.GoodsId,
				Type:       offerAliasOfRollAppOne.Type,
				Params:     offerAliasOfRollAppOne.Params,
				Buyer:      offerAliasOfRollAppOne.Buyer,
				OfferPrice: offerAliasOfRollAppOne.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offerAliasOfRollAppOne.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				requireAliasLinkedToRollApp(
					offerAliasOfRollAppOne.GoodsId, rollApp_One_By1_SingleAlias.rollAppId,
					t, ctx, dk,
				)
			},
		},
		{
			name: "fail - not accept offer if alias presents in params",
			existingRollApps: []rollapp{
				rollApp_One_By1_SingleAlias,
				rollApp_Two_By2_SingleAlias,
			},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "some-chain",
						Aliases: []string{offerAliasOfRollAppOne.GoodsId},
					},
				}
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:         true,
			wantErrContains: "prohibited to trade aliases which is reserved for chain-id or alias in module params",
			wantLaterOffer:  offerAliasOfRollAppOne,
			wantLaterRollApps: []rollapp{
				rollApp_One_By1_SingleAlias,
				rollApp_Two_By2_SingleAlias,
			},
			wantLaterModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantLaterOwnerBalance:  0,
			wantMinConsumeGas:      1,
		},
		{
			name:                  "fail - can NOT accept offer when trading Alias is disabled",
			existingRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:         offerAliasOfRollAppOne,
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
			preRunSetupFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Misc.EnableTradingAlias = false
				err := dk.SetParams(ctx, moduleParams)
				require.NoError(t, err)
			},
			wantErr:                true,
			wantErrContains:        "trading of Alias is disabled",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantLaterOwnerBalance:  0,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          nil,
			offerId:                "201",
			owner:                  rollApp_One_By1_SingleAlias.creator,
			minAccept:              dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order: 201: not found",
			wantLaterOffer:         nil,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - offer not found",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerAliasOfRollAppOne,
			offerId:                "20673264823",
			owner:                  rollApp_One_By1_SingleAlias.creator,
			minAccept:              dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Buy-Order: 20673264823: not found",
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - Alias not found",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerNonExistingAlias, // offer non-existing alias
			offerId:                offerNonExistingAlias.Id,
			owner:                  rollApp_One_By1_SingleAlias.creator,
			minAccept:              offerNonExistingAlias.OfferPrice,
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "alias is not in-used",
			wantLaterOffer:         offerNonExistingAlias,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - destination RollApp not exists",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerAliasForNonExistingRollApp, // offer for non-existing RollApp
			offerId:                offerAliasForNonExistingRollApp.Id,
			owner:                  rollApp_One_By1_SingleAlias.creator,
			minAccept:              offerAliasForNonExistingRollApp.OfferPrice,
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "invalid destination Roll-App ID",
			wantLaterOffer:         offerAliasForNonExistingRollApp,
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:                   "fail - can not accept offer of Alias owned by another",
			existingRollApps:       []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer:          offerAliasOfRollAppOne,
			offerId:                offerAliasOfRollAppOne.Id,
			owner:                  anotherAcc.bech32(),
			minAccept:              offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "not the owner of the RollApp",
			wantLaterRollApps:      []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer:         offerAliasOfRollAppOne,
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - can not accept own offer",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         offerAliasOfRollAppOne.Id,
					GoodsId:    offerAliasOfRollAppOne.GoodsId,
					Type:       offerAliasOfRollAppOne.Type,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      creator_1_asOwner,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 creator_1_asOwner,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "cannot accept own offer",
			wantLaterRollApps:     []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         offerAliasOfRollAppOne.Id,
					GoodsId:    offerAliasOfRollAppOne.GoodsId,
					Type:       offerAliasOfRollAppOne.Type,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      creator_1_asOwner,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - offer price denom != accept price denom",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:      offerAliasOfRollAppOne.Id,
					GoodsId: offerAliasOfRollAppOne.GoodsId,
					Type:    offerAliasOfRollAppOne.Type,
					Params:  offerAliasOfRollAppOne.Params,
					Buyer:   offerAliasOfRollAppOne.Buyer,
					OfferPrice: sdk.Coin{
						Denom:  denom,
						Amount: sdk.NewInt(minOfferPrice),
					},
				}
			}(),
			offerId: offerAliasOfRollAppOne.Id,
			owner:   rollApp_One_By1_SingleAlias.creator,
			minAccept: sdk.Coin{
				Denom:  "u" + denom,
				Amount: sdk.NewInt(minOfferPrice),
			},
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "denom must be the same as the offer price",
			wantLaterRollApps:     []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:      offerAliasOfRollAppOne.Id,
					GoodsId: offerAliasOfRollAppOne.GoodsId,
					Type:    offerAliasOfRollAppOne.Type,
					Params:  offerAliasOfRollAppOne.Params,
					Buyer:   offerAliasOfRollAppOne.Buyer,
					OfferPrice: sdk.Coin{
						Denom:  denom,
						Amount: sdk.NewInt(minOfferPrice),
					},
				}
			}(),
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:             "fail - accept price lower than offer price",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         offerAliasOfRollAppOne.Id,
					GoodsId:    offerAliasOfRollAppOne.GoodsId,
					Type:       offerAliasOfRollAppOne.Type,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      offerAliasOfRollAppOne.Buyer,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
				}
			}(),
			offerId:               offerAliasOfRollAppOne.Id,
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "amount must be greater than or equals to the offer price",
			wantLaterRollApps:     []rollapp{rollApp_One_By1_SingleAlias, rollApp_Two_By2_SingleAlias},
			wantLaterOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         offerAliasOfRollAppOne.Id,
					GoodsId:    offerAliasOfRollAppOne.GoodsId,
					Type:       offerAliasOfRollAppOne.Type,
					Params:     offerAliasOfRollAppOne.Params,
					Buyer:      offerAliasOfRollAppOne.Buyer,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
				}
			}(),
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      1,
		},
		{
			name:             "pass - accept offer transfer alias from One-Alias-RollApp to Multiple-Alias-RollApp",
			existingRollApps: []rollapp{rollApp_One_By1_SingleAlias, rollApp_Four_By2_MultipleAliases},
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         "201",
					GoodsId:    rollApp_One_By1_SingleAlias.aliases[0],
					Type:       dymnstypes.AliasOrder,
					Params:     []string{rollApp_Four_By2_MultipleAliases.rollAppId},
					Buyer:      creator_2_asBuyer,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			offerId:               "201",
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             offerAliasOfRollAppOne.OfferPrice,
			originalModuleBalance: offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			originalOwnerBalance:  0,
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
					aliases:   append(rollApp_Four_By2_MultipleAliases.aliases, offerAliasOfRollAppOne.GoodsId),
				},
			},
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  offerAliasOfRollAppOne.OfferPrice.Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
		},
		{
			name:             "pass - accept offer transfer alias from Multiple-Alias-RollApp to One-Alias-RollApp",
			existingRollApps: []rollapp{rollApp_Three_By1_MultipleAliases, rollApp_Two_By2_SingleAlias},
			existingOffer: func() *dymnstypes.BuyOffer {
				return &dymnstypes.BuyOffer{
					Id:         "201",
					GoodsId:    rollApp_Three_By1_MultipleAliases.aliases[0],
					Type:       dymnstypes.AliasOrder,
					Params:     []string{rollApp_Two_By2_SingleAlias.rollAppId},
					Buyer:      creator_2_asBuyer,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			offerId:               "201",
			owner:                 rollApp_One_By1_SingleAlias.creator,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: dymnsutils.TestCoin(minOfferPrice).Amount.Int64(),
			originalOwnerBalance:  0,
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
			wantLaterModuleBalance: 0,
			wantLaterOwnerBalance:  dymnsutils.TestCoin(minOfferPrice).Amount.Int64(),
			wantMinConsumeGas:      dymnstypes.OpGasUpdateBuyOffer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, dk, rk, bk := setupTest()

			if tt.originalModuleBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalModuleBalance),
				)
				require.NoError(t, err)
			}

			if tt.originalOwnerBalance > 0 {
				err := bk.MintCoins(
					ctx,
					dymnstypes.ModuleName,
					dymnsutils.TestCoins(tt.originalOwnerBalance),
				)
				require.NoError(t, err)

				err = bk.SendCoinsFromModuleToAccount(
					ctx,
					dymnstypes.ModuleName, sdk.MustAccAddressFromBech32(tt.owner),
					dymnsutils.TestCoins(tt.originalOwnerBalance),
				)
				require.NoError(t, err)
			}

			for _, rollApp := range tt.existingRollApps {
				rk.SetRollapp(ctx, rollapptypes.Rollapp{
					RollappId: rollApp.rollAppId,
					Owner:     rollApp.creator,
				})
				for _, alias := range rollApp.aliases {
					err := dk.SetAliasForRollAppId(ctx, rollApp.rollAppId, alias)
					require.NoError(t, err)
				}
			}

			if tt.existingOffer != nil {
				err := dk.SetBuyOffer(ctx, *tt.existingOffer)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToBuyOfferRecord(ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingGoodsIdToBuyOffer(ctx, tt.existingOffer.GoodsId, tt.existingOffer.Type, tt.existingOffer.Id)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).AcceptBuyOrder(ctx, &dymnstypes.MsgAcceptBuyOrder{
				OfferId:   tt.offerId,
				Owner:     tt.owner,
				MinAccept: tt.minAccept,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetBuyOffer(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := dk.GetBuyOffer(ctx, tt.offerId)
					require.Nil(t, laterOffer)
				}

				laterModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, denom)
				require.Equal(t, tt.wantLaterModuleBalance, laterModuleBalance.Amount.Int64())

				laterBuyerBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.owner), denom)
				require.Equal(t, tt.wantLaterOwnerBalance, laterBuyerBalance.Amount.Int64())

				require.Less(t, tt.wantMinConsumeGas, ctx.GasMeter().GasConsumed())

				for _, wantLaterRollApp := range tt.wantLaterRollApps {
					rollApp, found := rk.GetRollapp(ctx, wantLaterRollApp.rollAppId)
					require.True(t, found)
					if len(wantLaterRollApp.aliases) == 0 {
						requireRollAppHasNoAlias(rollApp.RollappId, t, ctx, dk)
					} else {
						for _, alias := range wantLaterRollApp.aliases {
							requireAliasLinkedToRollApp(alias, rollApp.RollappId, t, ctx, dk)
						}
					}
				}

				if tt.afterTestFunc != nil {
					tt.afterTestFunc(ctx, dk)
				}
			}()

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}
