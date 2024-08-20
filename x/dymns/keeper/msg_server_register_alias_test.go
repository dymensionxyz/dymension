package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (s *KeeperTestSuite) Test_msgServer_RegisterAlias() {
	denom := s.coin(0).Denom
	const price1L = 6
	const price2L = 5
	const price3L = 4
	const price4L = 3
	const price5PlusL = 2

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	rollApp1 := *newRollApp("rollapp_1-1").WithOwner(testAddr(1).bech32()).WithBech32("one").WithAlias("one")
	rollApp2 := *newRollApp("rollapp_2-2").WithOwner(testAddr(2).bech32()).WithBech32("two").WithAlias("two")
	rollApp3WithoutAlias := *newRollApp("rollapp_3-3").WithBech32("three").WithOwner(testAddr(3).bech32())

	buyerNotOwnedAnyRollApp := testAddr(4).bech32()

	s.persistRollApp(rollApp1)
	s.persistRollApp(rollApp2)
	s.persistRollApp(rollApp3WithoutAlias)

	const reservedAliasInParams = "reserved"

	const freeAlias8L = "accepted"

	const originalModuleBalance int64 = 88

	s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
		moduleParams.Price.AliasPriceSteps = []sdkmath.Int{
			sdk.NewInt(price1L).Mul(priceMultiplier),
			sdk.NewInt(price2L).Mul(priceMultiplier),
			sdk.NewInt(price3L).Mul(priceMultiplier),
			sdk.NewInt(price4L).Mul(priceMultiplier),
			sdk.NewInt(price5PlusL).Mul(priceMultiplier),
		}
		moduleParams.Price.PriceDenom = denom
		// reserved
		moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: "famous_9999-9",
				Aliases: []string{reservedAliasInParams},
			},
		}

		return moduleParams
	})

	s.mintToModuleAccount(originalModuleBalance)

	dymNameOwnerAcc := testAddr(5)
	dymName := dymnstypes.DymName{
		Name:       "my-name",
		Owner:      dymNameOwnerAcc.bech32(),
		Controller: dymNameOwnerAcc.bech32(),
		ExpireAt:   s.now.Unix() + 100,
	}
	s.setDymNameWithFunctionsAfter(dymName)

	s.SaveCurrentContext()

	s.Run("reject if message not pass validate basic", func() {
		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).RegisterAlias(s.ctx, &dymnstypes.MsgRegisterAlias{})
		s.Require().ErrorContains(err, gerrc.ErrInvalidArgument.Error())
	})

	tests := []struct {
		name                           string
		msg                            dymnstypes.MsgRegisterAlias
		buyerBalance                   int64
		preRunSetup                    func(s *KeeperTestSuite)
		wantErr                        bool
		wantErrContains                string
		wantLaterAliasLinkedTo         string
		wantLaterBuyerBalance          int64
		wantLaterAliasesOwnedByRollApp []string
	}{
		{
			name: "pass - can register",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          freeAlias8L,
				RollappId:      rollApp1.rollAppId,
				Owner:          rollApp1.owner,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price5PlusL)),
			},
			buyerBalance:                   price5PlusL + 1,
			wantErr:                        false,
			wantLaterAliasLinkedTo:         rollApp1.rollAppId,
			wantLaterBuyerBalance:          1,
			wantLaterAliasesOwnedByRollApp: append(rollApp1.aliases, freeAlias8L),
		},
		{
			name: "pass - can register, 2L",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          "oh",
				RollappId:      rollApp1.rollAppId,
				Owner:          rollApp1.owner,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price2L)),
			},
			buyerBalance:                   price2L + 1,
			wantErr:                        false,
			wantLaterAliasLinkedTo:         rollApp1.rollAppId,
			wantLaterBuyerBalance:          1,
			wantLaterAliasesOwnedByRollApp: append(rollApp1.aliases, "oh"),
		},
		{
			name: "pass - can register to RollApp that not owned any alias",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          "oh",
				RollappId:      rollApp3WithoutAlias.rollAppId,
				Owner:          rollApp3WithoutAlias.owner,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price2L)),
			},
			buyerBalance:                   price2L + 1,
			wantErr:                        false,
			wantLaterAliasLinkedTo:         rollApp3WithoutAlias.rollAppId,
			wantLaterBuyerBalance:          1,
			wantLaterAliasesOwnedByRollApp: []string{"oh"},
		},
		{
			name: "pass - deduct correct amount of balance",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          "oh",
				RollappId:      rollApp3WithoutAlias.rollAppId,
				Owner:          rollApp3WithoutAlias.owner,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price2L)),
			},
			buyerBalance:                   price1L,
			wantErr:                        false,
			wantLaterAliasLinkedTo:         rollApp3WithoutAlias.rollAppId,
			wantLaterBuyerBalance:          price1L - price2L,
			wantLaterAliasesOwnedByRollApp: []string{"oh"},
		},
		{
			name: "fail - reject if buyer does not have enough balance to register",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          freeAlias8L,
				RollappId:      rollApp1.rollAppId,
				Owner:          rollApp1.owner,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price5PlusL)),
			},
			buyerBalance:                   price5PlusL - 1,
			wantErr:                        true,
			wantErrContains:                "insufficient funds",
			wantLaterAliasLinkedTo:         "",
			wantLaterBuyerBalance:          price5PlusL - 1,
			wantLaterAliasesOwnedByRollApp: rollApp1.aliases,
		},
		{
			name:            "fail - reject bad request",
			msg:             dymnstypes.MsgRegisterAlias{},
			wantErr:         true,
			wantErrContains: gerrc.ErrInvalidArgument.Error(),
		},
		{
			name: "fail - reject alias occupied by another",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          rollApp2.alias,
				RollappId:      rollApp1.rollAppId,
				Owner:          rollApp1.owner,
				ConfirmPayment: sdk.NewCoin(denom, s.moduleParams().Price.GetAliasPrice(rollApp2.alias)),
			},
			buyerBalance:                   price1L,
			wantErr:                        true,
			wantErrContains:                "alias already in use or preserved",
			wantLaterAliasLinkedTo:         rollApp2.rollAppId,
			wantLaterBuyerBalance:          price1L,
			wantLaterAliasesOwnedByRollApp: rollApp1.aliases,
		},
		{
			name: "fail - reject alias reserved in params",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          reservedAliasInParams,
				RollappId:      rollApp1.rollAppId,
				Owner:          rollApp1.owner,
				ConfirmPayment: sdk.NewCoin(denom, s.moduleParams().Price.GetAliasPrice(reservedAliasInParams)),
			},
			buyerBalance:                   price1L,
			wantErr:                        true,
			wantErrContains:                "alias already in use or preserved",
			wantLaterAliasLinkedTo:         "",
			wantLaterBuyerBalance:          price1L,
			wantLaterAliasesOwnedByRollApp: rollApp1.aliases,
		},
		{
			name: "fail - reject if RollApp not found",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          freeAlias8L,
				RollappId:      "nah_0-0",
				Owner:          buyerNotOwnedAnyRollApp,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price5PlusL)),
			},
			buyerBalance:                   price1L,
			wantErr:                        true,
			wantErrContains:                "not found",
			wantLaterAliasLinkedTo:         "",
			wantLaterBuyerBalance:          price1L,
			wantLaterAliasesOwnedByRollApp: []string{},
		},
		{
			name: "fail - don't charge if tx failed",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          freeAlias8L,
				RollappId:      "nah_0-0",
				Owner:          buyerNotOwnedAnyRollApp,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price5PlusL)),
			},
			buyerBalance:          price1L,
			wantErr:               true,
			wantErrContains:       "not found",
			wantLaterBuyerBalance: price1L,
		},
		{
			name: "fail - reject if not owner of the RollApp",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          freeAlias8L,
				RollappId:      rollApp1.rollAppId,
				Owner:          buyerNotOwnedAnyRollApp,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price5PlusL)),
			},
			buyerBalance:                   price1L,
			wantErr:                        true,
			wantErrContains:                "not the owner of the RollApp",
			wantLaterAliasLinkedTo:         "",
			wantLaterBuyerBalance:          price1L,
			wantLaterAliasesOwnedByRollApp: rollApp1.aliases,
		},
		{
			name: "fail - reject if confirm payment does not match the actual amount",
			msg: dymnstypes.MsgRegisterAlias{
				Alias:          freeAlias8L,
				RollappId:      rollApp1.rollAppId,
				Owner:          rollApp1.owner,
				ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price1L)),
			},
			buyerBalance:                   price1L,
			wantErr:                        true,
			wantErrContains:                "actual payment is is different with provided by user",
			wantLaterAliasLinkedTo:         "",
			wantLaterBuyerBalance:          price1L,
			wantLaterAliasesOwnedByRollApp: rollApp1.aliases,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.buyerBalance > 0 {
				s.Require().NotEmpty(tt.msg.Owner)
				s.mintToAccount2(tt.msg.Owner, sdk.NewInt(tt.buyerBalance).Mul(priceMultiplier))
			}

			if tt.preRunSetup != nil {
				tt.preRunSetup(s)
			}

			if !tt.msg.ConfirmPayment.IsNil() {
				tt.msg.ConfirmPayment = sdk.NewCoin(denom, sdk.NewInt(tt.msg.ConfirmPayment.Amount.Int64()).Mul(priceMultiplier))
			}
			_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).RegisterAlias(s.ctx, &tt.msg)

			defer func() {
				if s.T().Failed() {
					return
				}

				s.Equal(originalModuleBalance, s.moduleBalance(), "module balance should not change because of burn")

				if tt.msg.Owner != "" && dymnsutils.IsValidBech32AccountAddress(tt.msg.Owner, true) {
					laterBuyerBalance := s.balance2(tt.msg.Owner)
					s.Equal(
						sdk.NewInt(tt.wantLaterBuyerBalance).Mul(priceMultiplier).String(),
						laterBuyerBalance.String(),
					)
				}
			}()

			defer func() {
				rollAppId, found := s.dymNsKeeper.GetRollAppIdByAlias(s.ctx, tt.msg.Alias)
				if tt.wantLaterAliasLinkedTo == "" {
					s.Falsef(found, "alias should not be linked to any RollApp but got: %s", rollAppId)
				} else {
					s.True(found, "alias should be linked to a RollApp")
					s.Equal(tt.wantLaterAliasLinkedTo, rollAppId, "alias should be linked to the RollApp")
				}

				if dymnsutils.IsValidAlias(tt.msg.Alias) {
					if s.dymNsKeeper.IsRollAppId(s.ctx, tt.msg.RollappId) {
						if len(tt.wantLaterAliasesOwnedByRollApp) == 0 {
							s.requireRollApp(tt.msg.RollappId).HasNoAlias()
						} else {
							s.requireRollApp(tt.msg.RollappId).HasAlias(tt.wantLaterAliasesOwnedByRollApp...)
						}
					}
				}
			}()

			if tt.wantErr {
				s.Require().ErrorContains(err, tt.wantErrContains)
				return
			}

			s.Require().NoError(err)

			rollApp, found := s.rollAppKeeper.GetRollapp(s.ctx, tt.msg.RollappId)
			s.Require().True(found)

			outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, fmt.Sprintf("%s@%s", dymName.Name, tt.msg.Alias))
			s.Require().NoError(err)
			s.Equal(dymNameOwnerAcc.bech32C(rollApp.Bech32Prefix), outputAddr, "resolution should be correct")

			if len(tt.wantLaterAliasesOwnedByRollApp) == 1 {
				outputDNA, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, dymNameOwnerAcc.bech32C(rollApp.Bech32Prefix), rollApp.RollappId)
				s.Require().NoError(err)
				s.Require().NotEmpty(outputDNA, "should have value")
				s.Equal(fmt.Sprintf("%s@%s", dymName.Name, tt.msg.Alias), outputDNA[0].String(), "reverse resolution should be correct")
			}
		})
	}

	s.Run("pass - new alias should be appended to the tails of the list", func() {
		s.RefreshContext()

		s.mintToAccount2(rollApp1.owner, sdk.NewInt(price1L).Mul(priceMultiplier))

		_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).RegisterAlias(s.ctx, &dymnstypes.MsgRegisterAlias{
			Alias:          freeAlias8L,
			RollappId:      rollApp1.rollAppId,
			Owner:          rollApp1.owner,
			ConfirmPayment: sdk.NewCoin(denom, sdk.NewInt(price5PlusL).Mul(priceMultiplier)),
		})
		s.Require().NoError(err)

		s.requireRollApp(rollApp1.rollAppId).HasAliasesWithOrder(append(rollApp1.aliases, freeAlias8L)...)
	})
}

func (s *KeeperTestSuite) TestEstimateRegisterAlias() {
	const denom = "atom"
	const price1L int64 = 9
	const price2L int64 = 8
	const price3L int64 = 7
	const price4L int64 = 6
	const price5PlusL int64 = 5

	// the number values used in this test will be multiplied by this value
	priceMultiplier := sdk.NewInt(1e18)

	priceParams := dymnstypes.DefaultPriceParams()
	priceParams.PriceDenom = denom
	priceParams.AliasPriceSteps = []sdkmath.Int{
		sdk.NewInt(price1L).Mul(priceMultiplier),
		sdk.NewInt(price2L).Mul(priceMultiplier),
		sdk.NewInt(price3L).Mul(priceMultiplier),
		sdk.NewInt(price4L).Mul(priceMultiplier),
		sdk.NewInt(price5PlusL).Mul(priceMultiplier),
	}

	tests := []struct {
		name      string
		alias     string
		wantPrice int64
	}{
		{
			name:      "1 letter",
			alias:     "a",
			wantPrice: price1L,
		},
		{
			name:      "2 letters",
			alias:     "oh",
			wantPrice: price2L,
		},
		{
			name:      "3 letters",
			alias:     "dog",
			wantPrice: price3L,
		},
		{
			name:      "4 letters",
			alias:     "pool",
			wantPrice: price4L,
		},
		{
			name:      "5 letters",
			alias:     "human",
			wantPrice: price5PlusL,
		},
		{
			name:      "6 letters",
			alias:     "planet",
			wantPrice: price5PlusL,
		},
		{
			name:      "5+ letters",
			alias:     "universe",
			wantPrice: price5PlusL,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := dymnskeeper.EstimateRegisterAlias(
				tt.alias,
				priceParams,
			)
			s.Equal(
				sdkmath.NewInt(tt.wantPrice).Mul(priceMultiplier).String(),
				got.Price.Amount.String(),
			)
			s.Equal(denom, got.Price.Denom)
		})
	}
}
