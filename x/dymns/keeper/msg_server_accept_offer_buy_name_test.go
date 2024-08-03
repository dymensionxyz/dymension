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
	"github.com/stretchr/testify/require"
)

func Test_msgServer_AcceptOfferBuyName(t *testing.T) {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const minOfferPrice = 5
	const daysProhibitSell = 30

	buyerA := testAddr(1).bech32()
	ownerA := testAddr(2).bech32()
	anotherOwnerA := testAddr(3).bech32()

	setupTest := func() (dymnskeeper.Keeper, dymnskeeper.BankKeeper, sdk.Context) {
		dk, bk, _, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now)

		moduleParams := dk.GetParams(ctx)
		// price
		moduleParams.Price.PriceDenom = denom
		moduleParams.Price.MinOfferPrice = sdk.NewInt(minOfferPrice)
		// misc
		moduleParams.Misc.ProhibitSellDuration = daysProhibitSell * 24 * time.Hour
		// submit
		err := dk.SetParams(ctx, moduleParams)
		require.NoError(t, err)

		return dk, bk, ctx
	}

	t.Run("reject if message not pass validate basic", func(t *testing.T) {
		dk, _, ctx := setupTest()

		requireErrorFContains(t, func() error {
			_, err := dymnskeeper.NewMsgServerImpl(dk).AcceptOfferBuyName(ctx, &dymnstypes.MsgAcceptOfferBuyName{})
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

	offer := &dymnstypes.OfferToBuy{
		Id:         "1",
		Name:       dymName.Name,
		Buyer:      buyerA,
		OfferPrice: dymnsutils.TestCoin(minOfferPrice),
	}

	tests := []struct {
		name                   string
		existingDymName        *dymnstypes.DymName
		existingOffer          *dymnstypes.OfferToBuy
		offerId                string
		owner                  string
		minAccept              sdk.Coin
		originalModuleBalance  int64
		originalOwnerBalance   int64
		preRunSetupFunc        func(ctx sdk.Context, dk dymnskeeper.Keeper)
		wantErr                bool
		wantErrContains        string
		wantLaterOffer         *dymnstypes.OfferToBuy
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
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
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
				offerIds := dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
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
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(dymName.Name)
				offerIds := dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Empty(t, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
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
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
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
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         offer.Id,
				Name:       offer.Name,
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
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
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
				offerIds := dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)
			},
			wantErr: false,
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         offer.Id,
				Name:       offer.Name,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
			afterTestFunc: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				key := dymnstypes.DymNameToOfferIdsRvlKey(dymName.Name)
				offerIds := dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
				require.Equal(t, []string{offer.Id}, offerIds.OfferIds)

				key = dymnstypes.BuyerToOfferIdsRvlKey(sdk.MustAccAddressFromBech32(offer.Buyer))
				offerIds = dk.GenericGetReverseLookupOfferToBuyIdsRecord(ctx, key)
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
			wantLaterOffer: &dymnstypes.OfferToBuy{
				Id:         offer.Id,
				Name:       offer.Name,
				Buyer:      offer.Buyer,
				OfferPrice: offer.OfferPrice,
				CounterpartyOfferPrice: func() *sdk.Coin {
					coin := offer.OfferPrice.Add(dymnsutils.TestCoin(1))
					return &coin
				}(),
			},
			wantLaterModuleBalance: 1,
			wantLaterOwnerBalance:  2,
			wantMinConsumeGas:      dymnstypes.OpGasUpdateOffer,
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
			name:                   "fail - offer not found",
			existingDymName:        dymName,
			existingOffer:          nil,
			offerId:                "1",
			owner:                  ownerA,
			minAccept:              dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Offer-To-Buy: 1: not found",
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
			offerId:                "673264823",
			owner:                  ownerA,
			minAccept:              dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance:  1,
			originalOwnerBalance:   2,
			wantErr:                true,
			wantErrContains:        "Offer-To-Buy: 673264823: not found",
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
			wantErrContains:        fmt.Sprintf("Dym-Name: %s: not found", offer.Name),
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
			wantErrContains:        fmt.Sprintf("Dym-Name: %s: not found", offer.Name),
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
			existingOffer: func() *dymnstypes.OfferToBuy {
				return &dymnstypes.OfferToBuy{
					Id:         "1",
					Name:       dymName.Name,
					Buyer:      ownerA,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice),
				}
			}(),
			offerId:               "1",
			owner:                 ownerA,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "cannot accept own offer",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.OfferToBuy {
				return &dymnstypes.OfferToBuy{
					Id:         "1",
					Name:       dymName.Name,
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
			existingOffer: func() *dymnstypes.OfferToBuy {
				return &dymnstypes.OfferToBuy{
					Id:    "1",
					Name:  dymName.Name,
					Buyer: buyerA,
					OfferPrice: sdk.Coin{
						Denom:  denom,
						Amount: sdk.NewInt(minOfferPrice),
					},
				}
			}(),
			offerId: "1",
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
			wantLaterOffer: func() *dymnstypes.OfferToBuy {
				return &dymnstypes.OfferToBuy{
					Id:    "1",
					Name:  dymName.Name,
					Buyer: buyerA,
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
			existingOffer: func() *dymnstypes.OfferToBuy {
				return &dymnstypes.OfferToBuy{
					Id:         "1",
					Name:       dymName.Name,
					Buyer:      buyerA,
					OfferPrice: dymnsutils.TestCoin(minOfferPrice + 2),
				}
			}(),
			offerId:               "1",
			owner:                 ownerA,
			minAccept:             dymnsutils.TestCoin(minOfferPrice),
			originalModuleBalance: 1,
			originalOwnerBalance:  2,
			wantErr:               true,
			wantErrContains:       "amount must be greater than or equals to the offer price",
			wantLaterDymName:      dymName,
			wantLaterOffer: func() *dymnstypes.OfferToBuy {
				return &dymnstypes.OfferToBuy{
					Id:         "1",
					Name:       dymName.Name,
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
				err := dk.SetOfferToBuy(ctx, *tt.existingOffer)
				require.NoError(t, err)

				err = dk.AddReverseMappingBuyerToOfferToBuyRecord(ctx, tt.existingOffer.Buyer, tt.existingOffer.Id)
				require.NoError(t, err)

				err = dk.AddReverseMappingDymNameToOfferToBuy(ctx, tt.existingOffer.Name, tt.existingOffer.Id)
				require.NoError(t, err)
			}

			if tt.preRunSetupFunc != nil {
				tt.preRunSetupFunc(ctx, dk)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(dk).AcceptOfferBuyName(ctx, &dymnstypes.MsgAcceptOfferBuyName{
				OfferId:   tt.offerId,
				Owner:     tt.owner,
				MinAccept: tt.minAccept,
			})

			defer func() {
				if t.Failed() {
					return
				}

				if tt.wantLaterOffer != nil {
					laterOffer := dk.GetOfferToBuy(ctx, tt.wantLaterOffer.Id)
					require.NotNil(t, laterOffer)
					require.Equal(t, *tt.wantLaterOffer, *laterOffer)
				} else {
					laterOffer := dk.GetOfferToBuy(ctx, tt.offerId)
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
