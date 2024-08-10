package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

func (s *KeeperTestSuite) Test_msgServer_RegisterName() {
	now := time.Now().UTC()

	denom := dymnsutils.TestCoin(0).Denom
	const firstYearPrice1L = 6
	const firstYearPrice2L = 5
	const firstYearPrice3L = 4
	const firstYearPrice4L = 3
	const firstYearPrice5PlusL = 2
	const extendsPrice = 1
	const gracePeriod = 30

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	buyerA := testAddr(1).bech32()
	previousOwnerA := testAddr(2).bech32()
	anotherA := testAddr(3).bech32()

	const preservedDymName = "preserved"

	preservedAddr1a := testAddr(4).bech32()
	preservedAddr2a := testAddr(5).bech32()

	setupParams := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Price.NamePriceSteps = []sdkmath.Int{
				sdk.NewInt(firstYearPrice1L).Mul(priceMultiplier),
				sdk.NewInt(firstYearPrice2L).Mul(priceMultiplier),
				sdk.NewInt(firstYearPrice3L).Mul(priceMultiplier),
				sdk.NewInt(firstYearPrice4L).Mul(priceMultiplier),
				sdk.NewInt(firstYearPrice5PlusL).Mul(priceMultiplier),
			}
			moduleParams.Price.PriceExtends = sdk.NewInt(extendsPrice).Mul(priceMultiplier)
			moduleParams.Price.PriceDenom = denom
			// misc
			moduleParams.Misc.GracePeriodDuration = gracePeriod * 24 * time.Hour
			// preserved
			moduleParams.PreservedRegistration.ExpirationEpoch = now.Add(time.Hour).Unix()
			moduleParams.PreservedRegistration.PreservedDymNames = []dymnstypes.PreservedDymName{
				{
					DymName:            preservedDymName,
					WhitelistedAddress: preservedAddr1a,
				},
				{
					DymName:            preservedDymName,
					WhitelistedAddress: preservedAddr2a,
				},
			}

			return moduleParams
		})
	}

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).RegisterName(s.ctx, &dymnstypes.MsgRegisterName{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	const originalModuleBalance int64 = 88

	tests := []struct {
		name                    string
		buyer                   string
		originalBalance         int64
		duration                int64
		confirmPayment          sdk.Coin
		contact                 string
		customDymName           string
		existingDymName         *dymnstypes.DymName
		setupHistoricalData     bool
		preRunSetup             func(s *KeeperTestSuite)
		wantLaterDymName        *dymnstypes.DymName
		wantErr                 bool
		wantErrContains         string
		wantLaterBalance        int64
		wantPruneHistoricalData bool
	}{
		{
			name:            "pass - can register, new Dym-Name",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			contact:         "contact@example.com",
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Contact:    "contact@example.com",
			},
			wantLaterBalance: 3,
		},
		{
			name:            "fail - not allow to takeover a non-expired Dym-Name",
			buyer:           buyerA,
			originalBalance: 1,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			contact:         "contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Add(time.Hour).Unix(),
				Contact:    "existing@example.com",
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Add(time.Hour).Unix(),
				Contact:    "existing@example.com",
			},
			wantErr:          true,
			wantErrContains:  "unauthenticated",
			wantLaterBalance: 1,
		},
		{
			name:            "fail - not allow to takeover an expired Dym-Name which in grace period",
			buyer:           buyerA,
			originalBalance: 1,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			wantErr:          true,
			wantErrContains:  "can be taken over after",
			wantLaterBalance: 1,
		},
		{
			name:             "fail - not enough balance to pay for the Dym-Name",
			buyer:            buyerA,
			originalBalance:  1,
			duration:         2,
			confirmPayment:   dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			wantErr:          true,
			wantErrContains:  "insufficient funds",
			wantLaterBalance: 1,
		},
		{
			name:             "fail - mis-match confirm payment",
			buyer:            buyerA,
			originalBalance:  firstYearPrice5PlusL + extendsPrice + 3,
			duration:         2,
			confirmPayment:   dymnsutils.TestCoin(1),
			wantErr:          true,
			wantErrContains:  "actual payment is is different with provided by user",
			wantLaterBalance: firstYearPrice5PlusL + extendsPrice + 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 5+ letters, multiple years",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice*2 + 3,
			duration:        3,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice*2),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*3,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 5+ letters, 1 year",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 4 letters, multiple years",
			buyer:           buyerA,
			customDymName:   "kids",
			originalBalance: firstYearPrice4L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice4L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 4 letters, 1 year",
			buyer:           buyerA,
			customDymName:   "kids",
			originalBalance: firstYearPrice4L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice4L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 3 letters, multiple years",
			buyer:           buyerA,
			customDymName:   "abc",
			originalBalance: firstYearPrice3L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice3L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 3 letters, 1 year",
			buyer:           buyerA,
			customDymName:   "abc",
			originalBalance: firstYearPrice3L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice3L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 2 letters, multiple years",
			buyer:           buyerA,
			customDymName:   "ab",
			originalBalance: firstYearPrice2L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice2L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 2 letters, 1 year",
			buyer:           buyerA,
			customDymName:   "ab",
			originalBalance: firstYearPrice2L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice2L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 1 letter, multiple years",
			buyer:           buyerA,
			customDymName:   "a",
			originalBalance: firstYearPrice1L + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice1L + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - deduct balance for new Dym-Name, 1 letter, 1 year",
			buyer:           buyerA,
			customDymName:   "a",
			originalBalance: firstYearPrice1L + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice1L),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - can extend owned Dym-Name, not expired",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1 + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - when extend owned non-expired Dym-Name, keep config and historical data",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
				Contact: "existing@example.com",
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1 + 86400*365*2,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
				Contact: "existing@example.com",
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: false,
		},
		{
			name:            "pass - when extend owned non-expired Dym-Name, keep config and historical data, update contact if provided",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
				Contact: "existing@example.com",
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1 + 86400*365*2,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
				Contact: "new-contact@example.com",
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: false,
		},
		{
			name:            "pass - can renew owned Dym-Name, expired",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - can renew owned Dym-Name, expired, update contact if provided",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Contact:    "new-contact@example.com",
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - when renew previously-owned expired Dym-Name, reset config",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   5,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: true,
		},
		{
			name:            "pass - when renew previously-owned expired Dym-Name, reset config, update contact if provided",
			buyer:           buyerA,
			originalBalance: extendsPrice*2 + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(extendsPrice * 2),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   5,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
			},
			setupHistoricalData: true,
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
				Contact:    "new-contact@example.com",
			},
			wantLaterBalance:        3,
			wantPruneHistoricalData: true,
		},
		{
			name:            "pass - can take over an expired Dym-Name after grace period has passed",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   1,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - take over an expired when ownership changed, reset config",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
				Contact: "old-contact@example.com",
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - take over an expired when ownership changed, reset config, update contact if provided",
			buyer:           buyerA,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			contact:         "new-contact@example.com",
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: buyerA,
				}},
				Contact: "old-contact@example.com",
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*2,
				Configs:    nil,
				Contact:    "new-contact@example.com",
			},
			wantLaterBalance: 3,
		},
		{
			name:            "fail - not enough balance to take over an expired Dym-Name after grace period has passed",
			buyer:           buyerA,
			originalBalance: 1,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			existingDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   3,
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   3,
			},
			wantErr:          true,
			wantErrContains:  "insufficient funds",
			wantLaterBalance: 1,
		},
		{
			name:            "pass - address in the preserved Dym-Name list, can still buy other Dym-Names",
			buyer:           preservedAddr1a,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      preservedAddr1a,
				Controller: preservedAddr1a,
				ExpireAt:   now.Unix() + 86400*365*1,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - only whitelisted address can buy preserved Dym-Name, addr 1",
			buyer:           preservedAddr1a,
			customDymName:   preservedDymName,
			originalBalance: firstYearPrice5PlusL + extendsPrice + 3,
			duration:        2,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL + extendsPrice),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      preservedAddr1a,
				Controller: preservedAddr1a,
				ExpireAt:   now.Unix() + 86400*365*2,
			},
			wantLaterBalance: 3,
		},
		{
			name:            "pass - only whitelisted address can buy preserved Dym-Name, addr 2",
			buyer:           preservedAddr2a,
			customDymName:   preservedDymName,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      preservedAddr2a,
				Controller: preservedAddr2a,
				ExpireAt:   now.Unix() + 86400*365*1,
			},
			wantLaterBalance: 3,
		},
		{
			name:             "fail - only whitelisted address can buy preserved Dym-Name, reject others",
			buyer:            buyerA,
			customDymName:    preservedDymName,
			originalBalance:  firstYearPrice5PlusL + 3,
			duration:         1,
			confirmPayment:   dymnsutils.TestCoin(firstYearPrice5PlusL),
			wantErr:          true,
			wantErrContains:  "only able to be registered by specific addresses",
			wantLaterBalance: firstYearPrice5PlusL + 3,
		},
		{
			name:            "pass - after preserved expiration date, anyone can buy preserved Dym-Name",
			buyer:           buyerA,
			customDymName:   preservedDymName,
			originalBalance: firstYearPrice5PlusL + 3,
			duration:        1,
			confirmPayment:  dymnsutils.TestCoin(firstYearPrice5PlusL),
			preRunSetup: func(s *KeeperTestSuite) {
				s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
					moduleParams.PreservedRegistration.ExpirationEpoch = now.Add(-time.Hour).Unix()
					return moduleParams
				})
			},
			wantLaterDymName: &dymnstypes.DymName{
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 86400*365*1,
			},
			wantLaterBalance: 3,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			setupParams(s)

			s.mintToModuleAccount2(sdk.NewInt(originalModuleBalance).Mul(priceMultiplier))

			if tt.originalBalance > 0 {
				s.mintToAccount2(tt.buyer, sdk.NewInt(tt.originalBalance).Mul(priceMultiplier))
			}

			useRecordName := "my-name"
			if tt.customDymName != "" {
				useRecordName = tt.customDymName
			}

			if tt.existingDymName != nil {
				tt.existingDymName.Name = useRecordName
				err := s.dymNsKeeper.SetDymName(s.ctx, *tt.existingDymName)
				s.Require().NoError(err)

				if tt.setupHistoricalData {
					so1 := dymnstypes.SellOrder{
						AssetId:   useRecordName,
						AssetType: dymnstypes.TypeName,
						ExpireAt:  now.Unix() - 1,
						MinPrice:  dymnsutils.TestCoin(1),
					}
					err := s.dymNsKeeper.SetSellOrder(s.ctx, so1)
					s.Require().NoError(err)

					err = s.dymNsKeeper.MoveSellOrderToHistorical(s.ctx, useRecordName, dymnstypes.TypeName)
					s.Require().NoError(err)

					so2 := dymnstypes.SellOrder{
						AssetId:   useRecordName,
						AssetType: dymnstypes.TypeName,
						ExpireAt:  tt.existingDymName.ExpireAt - 1,
						MinPrice:  dymnsutils.TestCoin(1),
						SellPrice: dymnsutils.TestCoinP(2),
						HighestBid: &dymnstypes.SellOrderBid{
							Bidder: anotherA,
							Price:  dymnsutils.TestCoin(2),
						},
					}
					err = s.dymNsKeeper.SetSellOrder(s.ctx, so2)
					s.Require().NoError(err)

					s.Require().Len(
						s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, useRecordName, dymnstypes.TypeName), 1,
					)
				}
			} else {
				s.Require().False(tt.setupHistoricalData, "bad setup testcase")
			}
			if tt.wantLaterDymName != nil {
				tt.wantLaterDymName.Name = useRecordName
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			resp, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).RegisterName(s.ctx, &dymnstypes.MsgRegisterName{
				Name:           useRecordName,
				Duration:       tt.duration,
				Owner:          tt.buyer,
				ConfirmPayment: sdk.NewCoin(tt.confirmPayment.Denom, tt.confirmPayment.Amount.Mul(priceMultiplier)),
				Contact:        tt.contact,
			})
			laterDymName := s.dymNsKeeper.GetDymName(s.ctx, useRecordName)

			defer func() {
				laterBalance := s.balance2(tt.buyer)
				s.Equal(
					sdk.NewInt(tt.wantLaterBalance).Mul(priceMultiplier).String(),
					laterBalance.String(),
				)
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				s.Nil(resp)

				defer func() {
					laterModuleBalance := s.moduleBalance2()
					s.Equal(
						sdk.NewInt(originalModuleBalance).Mul(priceMultiplier).String(),
						laterModuleBalance.String(),
						"module account balance should not be changed",
					)
				}()

				if tt.existingDymName != nil {
					s.Equal(tt.existingDymName.Name, laterDymName.Name, "should not change existing record")
					s.Require().NotNil(tt.wantLaterDymName, "bad setup testcase")
					s.Equal(*tt.wantLaterDymName, *laterDymName)
				} else {
					s.Nil(laterDymName)
					s.Nil(tt.wantLaterDymName, "bad setup testcase")
				}

				if tt.setupHistoricalData {
					s.NotNil(
						s.dymNsKeeper.GetSellOrder(s.ctx, useRecordName, dymnstypes.TypeName),
						"sell order must be kept",
					)
					s.Len(
						s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, useRecordName, dymnstypes.TypeName),
						1,
						"historical data must be kept",
					)
				}
				return
			}

			s.Require().NoError(err)
			s.NotNil(resp)

			defer func() {
				laterModuleBalance := s.moduleBalance2()
				s.Equal(
					sdk.NewInt(originalModuleBalance).Mul(priceMultiplier).String(), laterModuleBalance.String(),
					"token should be burned",
				)
			}()

			s.NotNil(laterDymName)
			s.NotNil(tt.wantLaterDymName, "bad setup testcase")
			s.Equal(*tt.wantLaterDymName, *laterDymName)

			if tt.setupHistoricalData {
				if tt.wantPruneHistoricalData {
					s.Nil(
						s.dymNsKeeper.GetSellOrder(s.ctx, useRecordName, dymnstypes.TypeName),
						"sell order must be pruned",
					)
					s.Empty(
						s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, useRecordName, dymnstypes.TypeName),
						"historical data must be pruned",
					)

					if tt.existingDymName.Owner != laterDymName.Owner {
						ownedByPreviousOwner, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, tt.existingDymName.Owner)
						s.Require().NoError(err)
						s.Empty(ownedByPreviousOwner, "reverse mapping should be removed")

						mappedDymNamesByPreviousOwner, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, tt.existingDymName.Owner)
						s.Require().NoError(err)
						s.Empty(mappedDymNamesByPreviousOwner, "reverse mapping should be removed")

						mappedDymNamesByPreviousOwner, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx,
							sdk.MustAccAddressFromBech32(tt.existingDymName.Owner).Bytes(),
						)
						s.Require().NoError(err)
						s.Empty(mappedDymNamesByPreviousOwner, "reverse mapping should be removed")
					}
				} else {
					s.NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, useRecordName, dymnstypes.TypeName), "sell order must be kept")
					s.Len(s.dymNsKeeper.GetHistoricalSellOrders(s.ctx, useRecordName, dymnstypes.TypeName), 1, "historical data must be kept")
				}
			} else {
				s.False(tt.wantPruneHistoricalData, "bad setup testcase")
			}

			ownedByBuyer, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, tt.buyer)
			s.Require().NoError(err)
			s.Len(ownedByBuyer, 1, "reverse mapping should be set")
			s.Equal(useRecordName, ownedByBuyer[0].Name)

			mappedDymNamesByBuyer, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, tt.buyer)
			s.Require().NoError(err)
			s.Len(mappedDymNamesByBuyer, 1, "reverse mapping should be set")
			s.Equal(useRecordName, mappedDymNamesByBuyer[0].Name)

			mappedDymNamesByBuyer, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx,
				sdk.MustAccAddressFromBech32(tt.buyer).Bytes(),
			)
			s.Require().NoError(err)
			s.Len(mappedDymNamesByBuyer, 1, "reverse mapping should be set")
			s.Equal(useRecordName, mappedDymNamesByBuyer[0].Name)
		})
	}
}

func (s *KeeperTestSuite) TestEstimateRegisterName() {
	now := time.Now()

	const denom = "atom"
	const price1L int64 = 9
	const price2L int64 = 8
	const price3L int64 = 7
	const price4L int64 = 6
	const price5PlusL int64 = 5
	const extendsPrice int64 = 4

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	params := dymnstypes.DefaultParams()
	params.Price.PriceDenom = denom
	params.Price.NamePriceSteps = []sdkmath.Int{
		sdk.NewInt(price1L).Mul(priceMultiplier),
		sdk.NewInt(price2L).Mul(priceMultiplier),
		sdk.NewInt(price3L).Mul(priceMultiplier),
		sdk.NewInt(price4L).Mul(priceMultiplier),
		sdk.NewInt(price5PlusL).Mul(priceMultiplier),
	}
	params.Price.PriceExtends = sdk.NewInt(extendsPrice).Mul(priceMultiplier)

	buyerA := testAddr(1).bech32()
	previousOwnerA := testAddr(2).bech32()

	tests := []struct {
		name               string
		dymName            string
		existingDymName    *dymnstypes.DymName
		newOwner           string
		duration           int64
		wantFirstYearPrice int64
		wantExtendPrice    int64
	}{
		{
			name:               "new registration, 1 letter, 1 year",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    0,
		},
		{
			name:               "new registration, 1 letter, 2 years",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:               "new registration, 1 letter, N years",
			dymName:            "a",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * (99 - 1),
		},
		{
			name:               "new registration, 6 letters, 1 year",
			dymName:            "bridge",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    0,
		},
		{
			name:               "new registration, 6 letters, 2 years",
			dymName:            "bridge",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:               "new registration, 5+ letters, N years",
			dymName:            "central",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * (99 - 1),
		},
		{
			name:    "extends same owner, 1 letter, 1 year",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:    "extends same owner, 1 letter, 2 years",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "extends same owner, 1 letter, N years",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 99,
		},
		{
			name:    "extends same owner, 6 letters, 1 year",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice,
		},
		{
			name:    "extends same owner, 6 letters, 2 years",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "extends same owner, 5+ letters, N years",
			dymName: "central",
			existingDymName: &dymnstypes.DymName{
				Name:       "central",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() + 1,
			},
			newOwner:           buyerA,
			duration:           99,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 99,
		},
		{
			name:    "extends expired, same owner, 5+ letters, 2 years",
			dymName: "central",
			existingDymName: &dymnstypes.DymName{
				Name:       "central",
				Owner:      buyerA,
				Controller: buyerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           2,
			wantFirstYearPrice: 0,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "take-over, 1 letter, 1 year",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    0,
		},
		{
			name:    "take-over, 1 letter, 3 years",
			dymName: "a",
			existingDymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price1L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:    "take-over, 6 letters, 1 year",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           1,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    0,
		},
		{
			name:    "take-over, 6 letters, 3 years",
			dymName: "bridge",
			existingDymName: &dymnstypes.DymName{
				Name:       "bridge",
				Owner:      previousOwnerA,
				Controller: previousOwnerA,
				ExpireAt:   now.Unix() - 1,
			},
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 2 letters",
			dymName:            "aa",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price2L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 3 letters",
			dymName:            "aaa",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price3L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 4 letters",
			dymName:            "geek",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price4L,
			wantExtendPrice:    extendsPrice * 2,
		},
		{
			name:               "new registration, 5 letters",
			dymName:            "human",
			existingDymName:    nil,
			newOwner:           buyerA,
			duration:           3,
			wantFirstYearPrice: price5PlusL,
			wantExtendPrice:    extendsPrice * 2,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := dymnskeeper.EstimateRegisterName(
				params,
				tt.dymName,
				tt.existingDymName,
				tt.newOwner,
				tt.duration,
			)
			s.Equal(
				sdkmath.NewInt(tt.wantFirstYearPrice).Mul(priceMultiplier).String(),
				got.FirstYearPrice.Amount.String(),
			)
			s.Equal(
				sdkmath.NewInt(tt.wantExtendPrice).Mul(priceMultiplier).String(),
				got.ExtendPrice.Amount.String(),
			)
			s.Equal(
				sdkmath.NewInt(tt.wantFirstYearPrice+tt.wantExtendPrice).Mul(priceMultiplier).String(),
				got.TotalPrice.Amount.String(),
				"total price must be equals to sum of first year and extend price",
			)
			s.Equal(denom, got.FirstYearPrice.Denom)
			s.Equal(denom, got.ExtendPrice.Denom)
			s.Equal(denom, got.TotalPrice.Denom)
		})
	}
}
