package keeper_test

import (
	"strings"
	"time"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/ethereum/go-ethereum/common"
)

func (s *KeeperTestSuite) TestKeeper_GetSetDeleteDymName() {
	ownerA := testAddr(1).bech32()

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: ownerA,
		}},
	}

	s.setDymNameWithFunctionsAfter(dymName)

	s.Run("event should be fired", func() {
		events := s.ctx.EventManager().Events()
		s.Require().NotEmpty(events)

		for _, event := range events {
			if event.Type == dymnstypes.EventTypeSetDymName {
				return
			}
		}

		s.T().Errorf("event %s not found", dymnstypes.EventTypeSetDymName)
	})

	s.Run("Dym-Name should be equals to original", func() {
		s.Require().Equal(dymName, *s.dymNsKeeper.GetDymName(s.ctx, dymName.Name))
	})

	s.Run("delete", func() {
		err := s.dymNsKeeper.DeleteDymName(s.ctx, dymName.Name)
		s.Require().NoError(err)
		s.Require().Nil(s.dymNsKeeper.GetDymName(s.ctx, dymName.Name))

		s.Run("reverse mapping should be deleted, check by key", func() {
			ownedBy := s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx,
				dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(ownerA)),
			)
			s.Require().NoError(err)
			s.Require().Empty(ownedBy, "reverse mapping should be removed")

			dymNames := s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx,
				dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(ownerA),
			)
			s.Require().NoError(err)
			s.Require().Empty(dymNames, "reverse mapping should be removed")

			dymNames = s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(s.ctx,
				dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(ownerA))),
			)
			s.Require().NoError(err)
			s.Require().Empty(dymNames, "reverse mapping should be removed")
		})

		s.Run("reverse mapping should be deleted, check by get", func() {
			ownedBy, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
			s.Require().NoError(err)
			s.Require().Empty(ownedBy, "reverse mapping should be removed")

			dymNames, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, ownerA)
			s.Require().NoError(err)
			s.Require().Empty(dymNames, "reverse mapping should be removed")

			dymNames, err = s.dymNsKeeper.GetDymNamesContainsFallbackAddress(s.ctx, sdk.MustAccAddressFromBech32(ownerA).Bytes())
			s.Require().NoError(err)
			s.Require().Empty(dymNames, "reverse mapping should be removed")
		})
	})

	s.Run("can not set invalid Dym-Name", func() {
		s.Require().Error(s.dymNsKeeper.SetDymName(s.ctx, dymnstypes.DymName{}))
	})

	s.Run("get returns nil if non-exists", func() {
		s.Require().Nil(s.dymNsKeeper.GetDymName(s.ctx, "non-exists"))
	})

	s.Run("delete a non-exists Dym-Name should be ok", func() {
		err := s.dymNsKeeper.DeleteDymName(s.ctx, "non-exists")
		s.Require().NoError(err)
	})
}

func (s *KeeperTestSuite) TestKeeper_BeforeAfterDymNameOwnerChanged() {
	s.Run("BeforeDymNameOwnerChanged can be called on non-existing Dym-Name without error", func() {
		s.Require().NoError(s.dymNsKeeper.BeforeDymNameOwnerChanged(s.ctx, "non-exists"))
	})

	s.Run("AfterDymNameOwnerChanged should returns error when calling on non-existing Dym-Name", func() {
		err := s.dymNsKeeper.AfterDymNameOwnerChanged(s.ctx, "non-exists")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Dym-Name: non-exists: not found")
	})

	ownerA := testAddr(1).bech32()

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   time.Now().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: ownerA,
		}},
	}

	s.Run("BeforeDymNameOwnerChanged will remove the reverse mapping owned-by", func() {
		s.setDymNameWithFunctionsAfter(dymName)

		owned, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
		s.Require().NoError(err)
		s.Require().Len(owned, 1)

		s.Require().NoError(s.dymNsKeeper.BeforeDymNameOwnerChanged(s.ctx, dymName.Name))

		owned, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
		s.Require().NoError(err)
		s.Require().Empty(owned)
	})

	s.Run("after run BeforeDymNameOwnerChanged, Dym-Name must be kept", func() {
		s.setDymNameWithFunctionsAfter(dymName)

		s.Require().NoError(s.dymNsKeeper.BeforeDymNameOwnerChanged(s.ctx, dymName.Name))

		s.Require().Equal(dymName, *s.dymNsKeeper.GetDymName(s.ctx, dymName.Name))
	})

	s.Run("AfterDymNameOwnerChanged will add the reverse mapping owned-by", func() {
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))

		owned, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
		s.Require().NoError(err)
		s.Require().Empty(owned)

		s.Require().NoError(s.dymNsKeeper.AfterDymNameOwnerChanged(s.ctx, dymName.Name))

		owned, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, ownerA)
		s.Require().NoError(err)
		s.Require().Len(owned, 1)
	})

	s.Run("after run AfterDymNameOwnerChanged, Dym-Name must be kept", func() {
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))

		s.Require().NoError(s.dymNsKeeper.AfterDymNameOwnerChanged(s.ctx, dymName.Name))

		s.Require().Equal(dymName, *s.dymNsKeeper.GetDymName(s.ctx, dymName.Name))
	})
}

func (s *KeeperTestSuite) TestKeeper_BeforeAfterDymNameConfigChanged() {
	s.Run("BeforeDymNameConfigChanged can be called on non-existing Dym-Name without error", func() {
		s.Require().NoError(s.dymNsKeeper.BeforeDymNameConfigChanged(s.ctx, "non-exists"))
	})

	s.Run("AfterDymNameConfigChanged should returns error when calling on non-existing Dym-Name", func() {
		err := s.dymNsKeeper.AfterDymNameConfigChanged(s.ctx, "non-exists")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Dym-Name: non-exists: not found")
	})

	ownerAcc := testAddr(1)
	controllerAcc := testAddr(2)
	icaAcc := testICAddr(3)

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerAcc.bech32(),
		Controller: controllerAcc.bech32(),
		ExpireAt:   time.Now().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{
			{
				Type:  dymnstypes.DymNameConfigType_DCT_NAME,
				Path:  "controller",
				Value: controllerAcc.bech32(),
			}, {
				Type:  dymnstypes.DymNameConfigType_DCT_NAME,
				Path:  "ica",
				Value: icaAcc.bech32(),
			},
		},
	}

	s.Run("BeforeDymNameConfigChanged will remove the reverse mapping address", func() {
		// do setup test

		s.setDymNameWithFunctionsAfter(dymName)

		s.requireConfiguredAddress(ownerAcc.bech32()).mappedDymNames(dymName.Name)
		s.requireConfiguredAddress(controllerAcc.bech32()).mappedDymNames(dymName.Name)
		s.requireConfiguredAddress(icaAcc.bech32()).mappedDymNames(dymName.Name)
		s.requireFallbackAddress(ownerAcc.bytes()).mappedDymNames(dymName.Name)
		s.requireFallbackAddress(controllerAcc.bytes()).notMappedToAnyDymName()
		s.requireFallbackAddress(icaAcc.bytes()).notMappedToAnyDymName()

		// do test

		s.Require().NoError(s.dymNsKeeper.BeforeDymNameConfigChanged(s.ctx, dymName.Name))

		s.requireConfiguredAddress(ownerAcc.bech32()).notMappedToAnyDymName()
		s.requireConfiguredAddress(controllerAcc.bech32()).notMappedToAnyDymName()
		s.requireConfiguredAddress(icaAcc.bech32()).notMappedToAnyDymName()
		s.requireFallbackAddress(ownerAcc.bytes()).notMappedToAnyDymName()
		s.requireFallbackAddress(controllerAcc.bytes()).notMappedToAnyDymName()
		s.requireFallbackAddress(icaAcc.bytes()).notMappedToAnyDymName()
	})

	s.Run("after run BeforeDymNameConfigChanged, Dym-Name must be kept", func() {
		s.setDymNameWithFunctionsAfter(dymName)

		s.Require().NoError(s.dymNsKeeper.BeforeDymNameConfigChanged(s.ctx, dymName.Name))

		s.Require().Equal(dymName, *s.dymNsKeeper.GetDymName(s.ctx, dymName.Name))
	})

	s.Run("AfterDymNameConfigChanged will add the reverse mapping address", func() {
		// do setup test

		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))

		s.requireConfiguredAddress(ownerAcc.bech32()).notMappedToAnyDymName()
		s.requireConfiguredAddress(controllerAcc.bech32()).notMappedToAnyDymName()
		s.requireConfiguredAddress(icaAcc.bech32()).notMappedToAnyDymName()
		s.requireFallbackAddress(ownerAcc.bytes()).notMappedToAnyDymName()
		s.requireFallbackAddress(controllerAcc.bytes()).notMappedToAnyDymName()
		s.requireFallbackAddress(icaAcc.bytes()).notMappedToAnyDymName()

		// do test

		s.Require().NoError(s.dymNsKeeper.AfterDymNameConfigChanged(s.ctx, dymName.Name))

		s.requireConfiguredAddress(ownerAcc.bech32()).mappedDymNames(dymName.Name)
		s.requireConfiguredAddress(controllerAcc.bech32()).mappedDymNames(dymName.Name)
		s.requireConfiguredAddress(icaAcc.bech32()).mappedDymNames(dymName.Name)
		s.requireFallbackAddress(ownerAcc.bytes()).mappedDymNames(dymName.Name)
		s.requireFallbackAddress(controllerAcc.bytes()).notMappedToAnyDymName()
		s.requireFallbackAddress(icaAcc.bytes()).notMappedToAnyDymName()
	})

	s.Run("after run AfterDymNameConfigChanged, Dym-Name must be kept", func() {
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))

		s.Require().NoError(s.dymNsKeeper.AfterDymNameConfigChanged(s.ctx, dymName.Name))

		s.Require().Equal(dymName, *s.dymNsKeeper.GetDymName(s.ctx, dymName.Name))
	})
}

func (s *KeeperTestSuite) TestKeeper_GetDymNameWithExpirationCheck() {
	now := time.Now().UTC()

	s.ctx = s.ctx.WithBlockTime(now)

	s.Run("returns nil if not exists", func() {
		s.Require().Nil(s.dymNsKeeper.GetDymNameWithExpirationCheck(s.ctx, "non-exists"))
	})

	ownerA := testAddr(1).bech32()

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Unix() + 100,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: ownerA,
		}},
	}

	err := s.dymNsKeeper.SetDymName(s.ctx, dymName)
	s.Require().NoError(err)

	s.Run("returns if not expired", func() {
		s.Require().NotNil(s.dymNsKeeper.GetDymNameWithExpirationCheck(s.ctx, dymName.Name))
	})

	s.Run("returns nil if expired", func() {
		dymName.ExpireAt = s.ctx.BlockTime().Unix() - 1000
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))
		s.Require().Nil(s.dymNsKeeper.GetDymNameWithExpirationCheck(
			s.ctx.WithBlockTime(time.Unix(dymName.ExpireAt+1, 0)), dymName.Name,
		))
	})
}

func (s *KeeperTestSuite) TestKeeper_GetAllDymNamesAndNonExpiredDymNames() {
	now := time.Now().UTC()

	s.ctx = s.ctx.WithBlockTime(now)

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()
	owner3a := testAddr(3).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: owner1a,
		}},
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: owner2a,
		}},
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName2))

	dymName3 := dymnstypes.DymName{
		Name:       "c",
		Owner:      owner3a,
		Controller: owner3a,
		ExpireAt:   s.now.Add(-time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: owner3a,
		}},
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName3))

	listNonExpired := s.dymNsKeeper.GetAllNonExpiredDymNames(s.ctx)
	s.Require().Len(listNonExpired, 2)
	s.Require().Contains(listNonExpired, dymName1)
	s.Require().Contains(listNonExpired, dymName2)
	s.Require().NotContains(listNonExpired, dymName3, "should not include expired Dym-Name")

	listAll := s.dymNsKeeper.GetAllDymNames(s.ctx)
	s.Require().Len(listAll, 3)
	s.Require().Contains(listAll, dymName1)
	s.Require().Contains(listAll, dymName2)
	s.Require().Contains(listAll, dymName3, "should include expired Dym-Name")
}

func (s *KeeperTestSuite) TestKeeper_GetDymNamesOwnedBy() {
	now := time.Now().UTC()

	s.ctx = s.ctx.WithBlockTime(now)

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()

	dymName11 := dymnstypes.DymName{
		Name:       "n11",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
	}
	s.setDymNameWithFunctionsAfter(dymName11)

	dymName12 := dymnstypes.DymName{
		Name:       "n12",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
	}
	s.setDymNameWithFunctionsAfter(dymName12)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
	}
	s.setDymNameWithFunctionsAfter(dymName21)

	s.Run("returns owned Dym-Names", func() {
		ownedBy, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, owner1a)
		s.Require().NoError(err)
		s.requireDymNameList(ownedBy, []string{dymName11.Name, dymName12.Name})
	})

	s.Run("returns owned Dym-Names with filtered expiration", func() {
		dymName12.ExpireAt = s.now.Add(-time.Hour).Unix()
		s.setDymNameWithFunctionsAfter(dymName12)

		ownedBy, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, owner1a)
		s.Require().NoError(err)
		s.requireDymNameList(ownedBy, []string{dymName11.Name})
	})
}

func (s *KeeperTestSuite) TestKeeper_PruneDymName() {
	now := time.Now().UTC()

	s.ctx = s.ctx.WithBlockTime(now)

	s.Run("prune non-exists Dym-Name should be ok", func() {
		s.Require().NoError(s.dymNsKeeper.PruneDymName(s.ctx, "non-exists"))
	})

	ownerA := testAddr(1).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
	}

	s.Run("able to prune non-expired Dym-Name", func() {
		s.setDymNameWithFunctionsAfter(dymName1)
		s.Require().NotNil(s.dymNsKeeper.GetDymName(s.ctx, dymName1.Name))

		s.Require().NoError(s.dymNsKeeper.PruneDymName(s.ctx, dymName1.Name))
		s.Require().Nil(s.dymNsKeeper.GetDymName(s.ctx, dymName1.Name))
	})

	// re-setup record
	s.setDymNameWithFunctionsAfter(dymName1)
	s.Require().NotNil(s.dymNsKeeper.GetDymName(s.ctx, dymName1.Name))
	owned, err := s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, dymName1.Owner)
	s.Require().NoError(err)
	s.Require().Len(owned, 1)

	// setup active SO
	so := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  s.now.Add(time.Hour).Unix(),
		MinPrice:  s.coin(100),
	}
	err = s.dymNsKeeper.SetSellOrder(s.ctx, so)
	s.Require().NoError(err)
	s.Require().NotNil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName))

	// prune
	err = s.dymNsKeeper.PruneDymName(s.ctx, dymName1.Name)
	s.Require().NoError(err)

	s.Require().Nil(s.dymNsKeeper.GetDymName(s.ctx, dymName1.Name), "Dym-Name should be removed")

	owned, err = s.dymNsKeeper.GetDymNamesOwnedBy(s.ctx, dymName1.Owner)
	s.Require().NoError(err)
	s.Require().Empty(owned, "reserve mapping should be removed")

	s.Require().Nil(s.dymNsKeeper.GetSellOrder(s.ctx, dymName1.Name, dymnstypes.TypeName), "active SO should be removed")
}

//goland:noinspection SpellCheckingInspection
func (s *KeeperTestSuite) TestKeeper_ResolveByDymNameAddress() {
	addr1a := testAddr(1).bech32()

	addr2Acc := testAddr(2)
	addr2a := addr2Acc.bech32()

	addr3a := testAddr(3).bech32()

	generalSetupAlias := func(s *KeeperTestSuite) {
		s.updateModuleParams(func(moduleParams dymnstypes.Params) dymnstypes.Params {
			moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
				{
					ChainId: s.chainId,
					Aliases: []string{"dym", "dymension"},
				},
				{
					ChainId: "blumbus_111-1",
					Aliases: []string{"bb", "blumbus"},
				},
				{
					ChainId: "froopyland_100-1",
					Aliases: nil,
				},
			}
			return moduleParams
		})
	}

	tests := []struct {
		name              string
		dymName           *dymnstypes.DymName
		preSetup          func(*KeeperTestSuite)
		dymNameAddress    string
		wantError         bool
		wantErrContains   string
		wantOutputAddress string
		postTest          func(*KeeperTestSuite)
	}{
		{
			name: "success, no sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: addr3a,
				}},
			},
			dymNameAddress:    "a.dymension_1100-1",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, no sub name, chain-id, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: addr3a,
				}},
			},
			dymNameAddress:    "a@dymension_1100-1",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr3a,
				}},
			},
			dymNameAddress:    "b.a.dymension_1100-1",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, sub name, chain-id, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr3a,
				}},
			},
			dymNameAddress:    "b.a@dymension_1100-1",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, multi-sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}},
			},
			dymNameAddress:    "c.b.a.dymension_1100-1",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, multi-sub name, chain-id, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}},
			},
			dymNameAddress:    "c.b.a@dymension_1100-1",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, RollApp chain-ID",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "1122",
					Value:   addr2Acc.bech32C("nim"),
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "1122",
					Path:    "b",
					Value:   addr2Acc.bech32C("nim"),
				}},
			},
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(*newRollApp("nim_1122-1"))
			},
			dymNameAddress:    "a@nim_1122-1",
			wantOutputAddress: addr2Acc.bech32C("nim"),
			postTest:          func(s *KeeperTestSuite) {},
		},
		{
			name: "success, no sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a.dym",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, no sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a@dym",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "b.a.dym",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "b.a@dym",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, multi-sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a.dym",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, match multiple alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr2a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a.dymension",
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				var outputAddr string
				var err error

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "c.b.a.dym")
				s.Require().NoError(err)
				s.Require().Equal(addr3a, outputAddr)
			},
		},
		{
			name: "success, multi-sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a@dym",
			wantOutputAddress: addr3a,
		},
		{
			name: "success, multi-sub config, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr2a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr1a,
				}},
			},
			preSetup:          nil,
			dymNameAddress:    "c.b.a.dymension_1100-1",
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				var outputAddr string
				var err error

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a.dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a@dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a@dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr1a, outputAddr)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@dym")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "no resolution found")

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "non-exists.a@dymension_1100-1")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "no resolution found")
			},
		},
		{
			name: "success, multi-sub config, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr2a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr1a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a@dym",
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				var outputAddr string
				var err error

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a.dym")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a.dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a@dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a@dym")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@dym")
				s.Require().NoError(err)
				s.Require().Equal(addr1a, outputAddr)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "non-exists.a@dym")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "no resolution found")
			},
		},
		{
			name: "success, alias of RollApp",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "1122",
					Value:   addr2Acc.bech32C("nim"),
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "1122",
					Path:    "b",
					Value:   addr2Acc.bech32C("nim"),
				}},
			},
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(*newRollApp("nim_1122-1").WithBech32("nim").WithAlias("nim"))
			},
			dymNameAddress:    "a@nim",
			wantOutputAddress: addr2Acc.bech32C("nim"),
			postTest: func(s *KeeperTestSuite) {
				// should be able to resolve if multiple aliases attached to the same RollApp

				aliases := []string{"nim1", "nim2", "nim3"}
				for _, alias := range aliases {
					s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "nim_1122-1", alias))
				}

				for _, alias := range aliases {
					outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@"+alias)
					s.Require().NoError(err)
					s.Require().Equal(addr2Acc.bech32C("nim"), outputAddr)

					outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.a@"+alias)
					s.Require().NoError(err)
					s.Require().Equal(addr2Acc.bech32C("nim"), outputAddr)
				}
			},
		},
		{
			name: "lookup through multiple sub-domains",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "b",
					Value: addr3a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr3a,
				}},
			},
			preSetup: func(s *KeeperTestSuite) {
				dymNameB := dymnstypes.DymName{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr2a,
					ExpireAt:   s.now.Unix() + 100,
					Configs: []dymnstypes.DymNameConfig{{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Path:  "b",
						Value: addr2a,
					}, {
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Path:  "",
						Value: addr2a,
					}},
				}
				s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymNameB))
			},
			dymNameAddress:    "b.a.dymension_1100-1",
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				var outputAddr string
				var err error

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b@dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "b.b.dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)
			},
		},
		{
			name: "matching by chain-id, no alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "b",
					Value:   addr2a,
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "",
					Value:   addr2a,
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr3a,
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "",
					Value:   addr3a,
				}},
			},
			dymNameAddress:    "a.blumbus_111-1",
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				var outputAddr string
				var err error

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.blumbus_111-1")
				s.Require().NoError(err)
				s.Require().Equal(addr3a, outputAddr)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@bb")
				s.Require().Error(err)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@blumbus")
				s.Require().Error(err)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dym")
				s.Require().Error(err)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dymension")
				s.Require().Error(err)
			},
		},
		{
			name: "matching by chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "b",
					Value:   addr2a,
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "",
					Value:   addr2a,
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr3a,
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "",
					Value:   addr3a,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a.blumbus_111-1",
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				var outputAddr string
				var err error

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.blumbus_111-1")
				s.Require().NoError(err)
				s.Require().Equal(addr3a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@bb")
				s.Require().NoError(err)
				s.Require().Equal(addr3a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@blumbus")
				s.Require().NoError(err)
				s.Require().Equal(addr3a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dymension_1100-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dym")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dymension")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)
			},
		},
		{
			name: "not configured sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "c.b",
					Value: addr3a,
				}, {
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr2a,
				}},
			},
			dymNameAddress:  "b.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "no resolution found",
		},
		{
			name: "when Dym-Name does not exists",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr3a,
				}},
			},
			dymNameAddress:  "b@dym",
			wantError:       true,
			wantErrContains: "Dym-Name: b: not found",
		},
		{
			name: "resolve to owner when no Dym-Name config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs:    nil,
			},
			dymNameAddress:    "a.dymension_1100-1",
			wantError:         false,
			wantOutputAddress: addr1a,
		},
		{
			name: "resolve to non-bech32/non-hex",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "another",
						Path:    "",
						Value:   "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5",
					},
				},
			},
			dymNameAddress:    "a.another",
			wantError:         false,
			wantOutputAddress: "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5",
		},
		{
			name: "resolve to non-bech32/non-hex, with sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "another",
						Path:    "sub1",
						Value:   "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5",
					},
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "another",
						Path:    "sub2",
						Value:   "Ae2tdPwUPEZFSi1cTyL1ZL6bgixhc2vSy5heg6Zg9uP7PpumkAJ82Qprt8b",
					},
				},
			},
			dymNameAddress:    "sub2.a.another",
			wantError:         false,
			wantOutputAddress: "Ae2tdPwUPEZFSi1cTyL1ZL6bgixhc2vSy5heg6Zg9uP7PpumkAJ82Qprt8b",
			postTest: func(s *KeeperTestSuite) {
				list, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, "Ae2tdPwUPEZFSi1cTyL1ZL6bgixhc2vSy5heg6Zg9uP7PpumkAJ82Qprt8b", "another")
				s.Require().NoError(err)
				s.Require().Len(list, 1)
				s.Require().Equal("sub2.a@another", list[0].String())

				list, err = s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5", "another")
				s.Require().NoError(err)
				s.Require().Len(list, 1)
				s.Require().Equal("sub1.a@another", list[0].String())
			},
		},
		{
			name: "resolve to owner when no default (without sub-name) Dym-Name config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Path:  "sub",
						Value: addr3a,
					},
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "blumbus_111-1",
						Path:    "",
						Value:   addr2a,
					},
				},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a.dymension_1100-1",
			wantError:         false,
			wantOutputAddress: addr1a,
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "sub.a.dym")
				s.Require().NoError(err)
				s.Require().Equal(addr3a, outputAddr)

				_, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "non-exists.a.dym")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), "no resolution found")

				outputAddr, err = s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@bb")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)
			},
		},
		{
			name: "do not fallback for sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs:    nil,
			},
			dymNameAddress:  "sub.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "no resolution found",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a.dymension_1100-1")
				s.Require().NoError(err, "should fallback if not sub-name")
				s.Require().Equal(addr1a, outputAddr)
			},
		},
		{
			name: "should not resolve for expired Dym-Name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() - 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr3a,
				}},
			},
			dymNameAddress:  "a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "Dym-Name: a: not found",
		},
		{
			name: "should not resolve for expired Dym-Name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() - 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr3a,
				}},
			},
			dymNameAddress:  "a.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "Dym-Name: a: not found",
		},
		{
			name: "should not resolve if input addr is invalid",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_DCT_NAME,
					Path:  "",
					Value: addr3a,
				}},
			},
			dymNameAddress:  "a@a.dymension_1100-1",
			wantError:       true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name: "if alias collision with configured record, priority configuration",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "blumbus_111-1",
						Path:    "",
						Value:   addr2a,
					},
					{
						Type:    dymnstypes.DymNameConfigType_DCT_NAME,
						ChainId: "blumbus",
						Path:    "",
						Value:   addr3a,
					},
				},
			},
			preSetup: func(s *KeeperTestSuite) {
				params := s.dymNsKeeper.GetParams(s.ctx)
				params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "blumbus_111-1",
						Aliases: []string{"blumbus"},
					},
				}
				err := s.dymNsKeeper.SetParams(s.ctx, params)
				s.Require().NoError(err)
			},
			dymNameAddress:    "a.blumbus",
			wantError:         false,
			wantOutputAddress: addr3a,
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "a@blumbus_111-1")
				s.Require().NoError(err)
				s.Require().Equal(addr2a, outputAddr)
			},
		},
		{
			name:              "resolve extra format 0x1234...6789@dym",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "0x1234567890123456789012345678901234567890.dym")
				s.Require().NoError(err)
				s.Require().Equal("dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96", outputAddr)
			},
		},
		{
			name:              "resolve extra format 0x1234...6789@dym, do not resolve if chain-id is unknown",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890@unknown-1",
			wantError:         true,
			wantErrContains:   "Dym-Name: 0x1234567890123456789012345678901234567890: not found",
			wantOutputAddress: "",
		},
		{
			name:    "resolve extra format 0x1234...6789@dym, do not resolve if chain-id is not RollApp, even tho alias was defined",
			dymName: nil,
			preSetup: func(s *KeeperTestSuite) {
				params := s.dymNsKeeper.GetParams(s.ctx)
				params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "blumbus_111-1",
						Aliases: []string{"blumbus"},
					},
				}
				err := s.dymNsKeeper.SetParams(s.ctx, params)
				s.Require().NoError(err)
			},
			dymNameAddress:    "0x1234567890123456789012345678901234567890@blumbus",
			wantError:         true,
			wantErrContains:   "Dym-Name: 0x1234567890123456789012345678901234567890: not found",
			wantOutputAddress: "",
		},
		{
			name:              "resolve extra format 0x1234...6789@dym, Interchain Account",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890123456789012345678901234@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "0x1234567890123456789012345678901234567890123456789012345678901234.dym")
				s.Require().NoError(err)
				s.Require().Equal("dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul", outputAddr)
			},
		},
		{
			name:              "resolve extra format nim1...@dym, cross bech32 format",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9.dym")
				s.Require().NoError(err)
				s.Require().Equal("dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96", outputAddr)
			},
		},
		{
			name: "fallback resolve follow default config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr1a,
				ExpireAt:   s.now.Unix() + 100,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: addr2Acc.bech32(),
					},
				},
			},
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(
					*newRollApp("nim_1122-1").WithBech32("nim").WithAlias("nim"),
				)
			},
			dymNameAddress:    "a@nim",
			wantError:         false,
			wantOutputAddress: addr2Acc.bech32C("nim"),
			postTest:          nil,
		},
		{
			name:    "resolve extra format 0x1234...6789@nim (RollApp)",
			dymName: nil,
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(*newRollApp("nim_1122-1").WithBech32("nim").WithAlias("nim"))
			},
			dymNameAddress:    "0x1234567890123456789012345678901234567890@nim_1122-1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "0x1234567890123456789012345678901234567890.nim")
				s.Require().NoError(err)
				s.Require().Equal("nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "resolve extra format 0x1234...6789@nim1 (RollApp), alternative alias among multiple aliases for RollApp",
			dymName: nil,
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(*newRollApp("nim_1122-1").WithBech32("nim").WithAlias("nim"))
				s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "nim_1122-1", "nim1"))
				s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "nim_1122-1", "nim2"))
			},
			dymNameAddress:    "0x1234567890123456789012345678901234567890@nim1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "0x1234567890123456789012345678901234567890.nim")
				s.Require().NoError(err)
				s.Require().Equal("nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "resolve extra format dym1...@nim (RollApp), cross bech32 format",
			dymName: nil,
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(*newRollApp("nim_1122-1").WithBech32("nim").WithAlias("nim"))
			},
			dymNameAddress:    "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@nim_1122-1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96.nim")
				s.Require().NoError(err)
				s.Require().Equal("nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "resolve extra format dym1...@nim1 (RollApp), cross bech32 format, alternative alias among multiple aliases for RollApp",
			dymName: nil,
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(*newRollApp("nim_1122-1").WithBech32("nim").WithAlias("nim"))
				s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "nim_1122-1", "nim1"))
				s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "nim_1122-1", "nim2"))
			},
			dymNameAddress:    "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@nim1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(s *KeeperTestSuite) {
				outputAddr, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96.nim")
				s.Require().NoError(err)
				s.Require().Equal("nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "try resolve extra format dym1...@rollapp, cross bech32 format, but RollApp does not have bech32 configured",
			dymName: nil,
			preSetup: func(s *KeeperTestSuite) {
				s.persistRollApp(
					*newRollApp("rollapp_1-1"),
				)
			},
			dymNameAddress:  "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@rollapp_1-1",
			wantError:       true,
			wantErrContains: "not found",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			if tt.preSetup != nil {
				tt.preSetup(s)
			}

			if tt.dymName != nil {
				s.setDymNameWithFunctionsAfter(*tt.dymName)
			}

			outputAddress, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, tt.dymNameAddress)

			defer func() {
				if s.T().Failed() {
					return
				}

				if tt.postTest != nil {
					tt.postTest(s)
				}
			}()

			if tt.wantError {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tt.wantOutputAddress, outputAddress)
		})
	}

	s.Run("mixed tests", func() {
		s.RefreshContext()

		bech32Addr := func(no uint64) string {
			return testAddr(no).bech32()
		}

		// setup alias
		moduleParams := s.dymNsKeeper.GetParams(s.ctx)
		moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: s.chainId,
				Aliases: []string{"dym"},
			},
			{
				ChainId: "blumbus_111-1",
				Aliases: []string{"bb"},
			},
			{
				ChainId: "froopyland_100-1",
				Aliases: nil,
			},
			{
				ChainId: "cosmoshub-4",
				Aliases: []string{"cosmos"},
			},
		}
		s.Require().NoError(s.dymNsKeeper.SetParams(s.ctx, moduleParams))

		// setup Dym-Names
		dymName1 := dymnstypes.DymName{
			Name:       "name1",
			Owner:      bech32Addr(1),
			Controller: bech32Addr(2),
			ExpireAt:   s.now.Unix() + 100,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   bech32Addr(3),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s2",
					Value:   bech32Addr(4),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "a.s5",
					Value:   bech32Addr(5),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   bech32Addr(6),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "c.b",
					Value:   bech32Addr(7),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "juno-1",
					Path:    "",
					Value:   bech32Addr(8),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "juno-1",
					Path:    "a.b.c",
					Value:   bech32Addr(9),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "cosmoshub-4",
					Path:    "",
					Value:   bech32Addr(10),
				},
			},
		}
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName1))

		dymName2 := dymnstypes.DymName{
			Name:       "name2",
			Owner:      bech32Addr(100),
			Controller: bech32Addr(101),
			ExpireAt:   s.now.Unix() + 100,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   bech32Addr(103),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s2",
					Value:   bech32Addr(104),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "a.s5",
					Value:   bech32Addr(105),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   bech32Addr(106),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "c.b",
					Value:   bech32Addr(107),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "juno-1",
					Path:    "",
					Value:   bech32Addr(108),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "juno-1",
					Path:    "a.b.c",
					Value:   bech32Addr(109),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "froopyland_100-1",
					Path:    "a",
					Value:   bech32Addr(110),
				},
			},
		}
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName2))

		dymName3 := dymnstypes.DymName{
			Name:       "name3",
			Owner:      bech32Addr(200),
			Controller: bech32Addr(201),
			ExpireAt:   s.now.Unix() + 100,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   bech32Addr(203),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s2",
					Value:   bech32Addr(204),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "a.s5",
					Value:   bech32Addr(205),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   bech32Addr(206),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "blumbus_111-1",
					Path:    "c.b",
					Value:   bech32Addr(207),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "juno-1",
					Path:    "",
					Value:   bech32Addr(208),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "juno-1",
					Path:    "a.b.c",
					Value:   bech32Addr(209),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "froopyland_100-1",
					Path:    "a",
					Value:   bech32Addr(210),
				},
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "cosmoshub-4",
					Path:    "a",
					Value:   bech32Addr(211),
				},
			},
		}
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName3))

		dymName4 := dymnstypes.DymName{
			Name:       "name4",
			Owner:      "dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn",
			Controller: bech32Addr(301),
			ExpireAt:   s.now.Unix() + 100,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   bech32Addr(302),
				},
			},
		}
		s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName4))

		rollAppNim := rollapptypes.Rollapp{
			RollappId: "nim_1122-1",
			Owner:     bech32Addr(1122),
		}
		s.persistRollApp(*newRollApp(rollAppNim.RollappId).WithBech32("nim").WithAlias("nim"))
		rollAppNim, found := s.rollAppKeeper.GetRollapp(s.ctx, rollAppNim.RollappId)
		s.Require().True(found)

		tc := func(name, chainIdOrAlias string) input {
			return newInputTestcase(name, chainIdOrAlias, s)
		}

		tc("name1", s.chainId).WithSubName("s1").RequireResolveTo(bech32Addr(3))
		tc("name1", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(3))
		tc("name1", s.chainId).WithSubName("s2").RequireResolveTo(bech32Addr(4))
		tc("name1", "dym").WithSubName("s2").RequireResolveTo(bech32Addr(4))
		tc("name1", s.chainId).WithSubName("a.s5").RequireResolveTo(bech32Addr(5))
		tc("name1", "dym").WithSubName("a.s5").RequireResolveTo(bech32Addr(5))
		tc("name1", s.chainId).WithSubName("none").RequireNotResolve()
		tc("name1", "dym").WithSubName("none").RequireNotResolve()
		tc("name1", "blumbus_111-1").WithSubName("b").RequireResolveTo(bech32Addr(6))
		tc("name1", "bb").WithSubName("b").RequireResolveTo(bech32Addr(6))
		tc("name1", "blumbus_111-1").WithSubName("c.b").RequireResolveTo(bech32Addr(7))
		tc("name1", "bb").WithSubName("c.b").RequireResolveTo(bech32Addr(7))
		tc("name1", "blumbus_111-1").WithSubName("none").RequireNotResolve()
		tc("name1", "bb").WithSubName("none").RequireNotResolve()
		tc("name1", "juno-1").RequireResolveTo(bech32Addr(8))
		tc("name1", "juno-1").WithSubName("a.b.c").RequireResolveTo(bech32Addr(9))
		tc("name1", "juno-1").WithSubName("none").RequireNotResolve()
		tc("name1", "cosmoshub-4").RequireResolveTo(bech32Addr(10))
		tc("name1", "cosmos").RequireResolveTo(bech32Addr(10))

		tc("name2", s.chainId).WithSubName("s1").RequireResolveTo(bech32Addr(103))
		tc("name2", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(103))
		tc("name2", s.chainId).WithSubName("s2").RequireResolveTo(bech32Addr(104))
		tc("name2", "dym").WithSubName("s2").RequireResolveTo(bech32Addr(104))
		tc("name2", s.chainId).WithSubName("a.s5").RequireResolveTo(bech32Addr(105))
		tc("name2", "dym").WithSubName("a.s5").RequireResolveTo(bech32Addr(105))
		tc("name2", s.chainId).WithSubName("none").RequireNotResolve()
		tc("name2", "dym").WithSubName("none").RequireNotResolve()
		tc("name2", "blumbus_111-1").WithSubName("b").RequireResolveTo(bech32Addr(106))
		tc("name2", "bb").WithSubName("b").RequireResolveTo(bech32Addr(106))
		tc("name2", "blumbus_111-1").WithSubName("c.b").RequireResolveTo(bech32Addr(107))
		tc("name2", "bb").WithSubName("c.b").RequireResolveTo(bech32Addr(107))
		tc("name2", "blumbus_111-1").WithSubName("none").RequireNotResolve()
		tc("name2", "bb").WithSubName("none").RequireNotResolve()
		tc("name2", "juno-1").RequireResolveTo(bech32Addr(108))
		tc("name2", "juno-1").WithSubName("a.b.c").RequireResolveTo(bech32Addr(109))
		tc("name2", "juno-1").WithSubName("none").RequireNotResolve()
		tc("name2", "froopyland_100-1").WithSubName("a").RequireResolveTo(bech32Addr(110))
		tc("name2", "froopyland").WithSubName("a").RequireNotResolve()
		tc("name2", "cosmoshub-4").RequireNotResolve()
		tc("name2", "cosmoshub-4").WithSubName("a").RequireNotResolve()

		tc("name3", s.chainId).WithSubName("s1").RequireResolveTo(bech32Addr(203))
		tc("name3", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(203))
		tc("name3", s.chainId).WithSubName("s2").RequireResolveTo(bech32Addr(204))
		tc("name3", "dym").WithSubName("s2").RequireResolveTo(bech32Addr(204))
		tc("name3", s.chainId).WithSubName("a.s5").RequireResolveTo(bech32Addr(205))
		tc("name3", "dym").WithSubName("a.s5").RequireResolveTo(bech32Addr(205))
		tc("name3", s.chainId).WithSubName("none").RequireNotResolve()
		tc("name3", "dym").WithSubName("none").RequireNotResolve()
		tc("name3", "blumbus_111-1").WithSubName("b").RequireResolveTo(bech32Addr(206))
		tc("name3", "bb").WithSubName("b").RequireResolveTo(bech32Addr(206))
		tc("name3", "blumbus_111-1").WithSubName("c.b").RequireResolveTo(bech32Addr(207))
		tc("name3", "bb").WithSubName("c.b").RequireResolveTo(bech32Addr(207))
		tc("name3", "blumbus_111-1").WithSubName("none").RequireNotResolve()
		tc("name3", "bb").WithSubName("none").RequireNotResolve()
		tc("name3", "juno-1").RequireResolveTo(bech32Addr(208))
		tc("name3", "juno-1").WithSubName("a.b.c").RequireResolveTo(bech32Addr(209))
		tc("name3", "juno-1").WithSubName("none").RequireNotResolve()
		tc("name3", "froopyland_100-1").WithSubName("a").RequireResolveTo(bech32Addr(210))
		tc("name3", "froopyland").WithSubName("a").RequireNotResolve()
		tc("name3", "cosmoshub-4").RequireNotResolve()
		tc("name3", "cosmos").WithSubName("a").RequireResolveTo(bech32Addr(211))

		tc("name4", s.chainId).WithSubName("s1").RequireResolveTo(bech32Addr(302))
		tc("name4", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(302))
		tc("name4", s.chainId).WithSubName("none").RequireNotResolve()
		tc("name4", "dym").WithSubName("none").RequireNotResolve()
		tc("name4", s.chainId).RequireResolveTo("dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn")
		tc("name4", "dym").RequireResolveTo("dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn")
		tc("name4", rollAppNim.RollappId).RequireResolveTo(
			"nim1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8wkcvv",
		)
	})
}

type input struct {
	s *KeeperTestSuite
	//
	name           string
	chainIdOrAlias string
	subName        string
}

func newInputTestcase(name, chainIdOrAlias string, s *KeeperTestSuite) input {
	return input{name: name, chainIdOrAlias: chainIdOrAlias, s: s}
}

func (m input) WithSubName(subName string) input {
	m.subName = subName
	return m
}

func (m input) buildDymNameAddrsCases() []string {
	var dymNameAddrs []string
	func() {
		dymNameAddr := m.name + "." + m.chainIdOrAlias
		if len(m.subName) > 0 {
			dymNameAddr = m.subName + "." + dymNameAddr
		}
		dymNameAddrs = append(dymNameAddrs, dymNameAddr)
	}()
	func() {
		dymNameAddr := m.name + "@" + m.chainIdOrAlias
		if len(m.subName) > 0 {
			dymNameAddr = m.subName + "." + dymNameAddr
		}
		dymNameAddrs = append(dymNameAddrs, dymNameAddr)
	}()
	return dymNameAddrs
}

func (m input) RequireNotResolve() {
	for _, dymNameAddr := range m.buildDymNameAddrsCases() {
		_, err := m.s.dymNsKeeper.ResolveByDymNameAddress(m.s.ctx, dymNameAddr)
		m.s.Require().Error(err)
	}
}

func (m input) RequireResolveTo(wantAddr string) {
	for _, dymNameAddr := range m.buildDymNameAddrsCases() {
		gotAddr, err := m.s.dymNsKeeper.ResolveByDymNameAddress(m.s.ctx, dymNameAddr)
		m.s.Require().NoError(err)
		m.s.Require().Equal(wantAddr, gotAddr)
	}
}

//goland:noinspection SpellCheckingInspection
func (s *KeeperTestSuite) Test_ParseDymNameAddress() {
	tests := []struct {
		name               string
		dymNameAddress     string
		wantErr            bool
		wantErrContains    string
		wantSubName        string
		wantDymName        string
		wantChainIdOrAlias string
	}{
		{
			name:               "pass - valid input, no sub-name, chain-id, @",
			dymNameAddress:     "a@dymension_1100-1",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "pass - valid input, no sub-name, chain-id",
			dymNameAddress:     "a.dymension_1100-1",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "pass - valid input, no sub-name, alias, @",
			dymNameAddress:     "a@dym",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - valid input, no sub-name, alias",
			dymNameAddress:     "a.dym",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - valid input, sub-name, chain-id, @",
			dymNameAddress:     "b.a@dymension_1100-1",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "pass - valid input, sub-name, chain-id",
			dymNameAddress:     "b.a.dymension_1100-1",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "pass - valid input, sub-name, alias, @",
			dymNameAddress:     "b.a@dym",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - valid input, sub-name, alias",
			dymNameAddress:     "b.a.dym",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - valid input, multi-sub-name, chain-id, @",
			dymNameAddress:     "c.b.a@dymension_1100-1",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "pass - valid input, multi-sub-name, chain-id",
			dymNameAddress:     "c.b.a.dymension_1100-1",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "pass - valid input, multi-sub-name, alias, @",
			dymNameAddress:     "c.b.a@dym",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - valid input, multi-sub-name, alias",
			dymNameAddress:     "c.b.a.dym",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:            "fail - invalid '.' after '@', no sub-name",
			dymNameAddress:  "a@dymension_1100-1.dym",
			wantErr:         true,
			wantErrContains: "misplaced '.'",
		},
		{
			name:            "fail - invalid '.' after '@', sub-name",
			dymNameAddress:  "a.b@dymension_1100-1.dym",
			wantErr:         true,
			wantErrContains: "misplaced '.'",
		},
		{
			name:            "fail - invalid '.' after '@', multi-sub-name",
			dymNameAddress:  "a.b.c@dymension_1100-1.dym",
			wantErr:         true,
			wantErrContains: "misplaced '.'",
		},
		{
			name:            "fail - missing chain-id/alias, @",
			dymNameAddress:  "a@",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - missing chain-id/alias",
			dymNameAddress:  "a",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - missing chain-id/alias",
			dymNameAddress:  "a.",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - not accept space, no sub-name",
			dymNameAddress:  "a .dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - not accept space, sub-name",
			dymNameAddress:  "b .a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - not accept space, multi-sub-name",
			dymNameAddress:  "c.b .a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - invalid chain-id/alias, @",
			dymNameAddress:  "a@-dym",
			wantErr:         true,
			wantErrContains: "chain-id/alias is not well-formed",
		},
		{
			name:            "fail - invalid chain-id/alias",
			dymNameAddress:  "a.-dym",
			wantErr:         true,
			wantErrContains: "chain-id/alias is not well-formed",
		},
		{
			name:            "fail - invalid Dym-Name, @",
			dymNameAddress:  "-a@dym",
			wantErr:         true,
			wantErrContains: "Dym-Name is not well-formed",
		},
		{
			name:            "fail - invalid Dym-Name",
			dymNameAddress:  "-a.dym",
			wantErr:         true,
			wantErrContains: "Dym-Name is not well-formed",
		},
		{
			name:            "fail - invalid sub-Dym-Name, @",
			dymNameAddress:  "-b.a@dym",
			wantErr:         true,
			wantErrContains: "Sub-Dym-Name part is not well-formed",
		},
		{
			name:            "fail - invalid sub-Dym-Name",
			dymNameAddress:  "-b.a.dym",
			wantErr:         true,
			wantErrContains: "Sub-Dym-Name part is not well-formed",
		},
		{
			name:            "fail - invalid multi-sub-Dym-Name, @",
			dymNameAddress:  "c-.b.a@dym",
			wantErr:         true,
			wantErrContains: "Sub-Dym-Name part is not well-formed",
		},
		{
			name:            "fail - invalid multi-sub-Dym-Name",
			dymNameAddress:  "c-.b.a.dym",
			wantErr:         true,
			wantErrContains: "Sub-Dym-Name part is not well-formed",
		},
		{
			name:            "fail - blank path",
			dymNameAddress:  "b. .a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - do not accept continuous dot",
			dymNameAddress:  "b..a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - do not accept continuous '@'",
			dymNameAddress:  "a@@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "fail - do not accept continuous '@'",
			dymNameAddress:  "b.a@@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "fail - do not accept multiple '@'",
			dymNameAddress:  "b@a@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "fail - do not accept multiple '@'",
			dymNameAddress:  "@a@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "fail - do not accept multiple '@'",
			dymNameAddress:  "@a.b@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "fail - do not accept multiple '@'",
			dymNameAddress:  "a@b@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "fail - bad name",
			dymNameAddress:  "a.@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - bad name",
			dymNameAddress:  "a.b.@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - bad name",
			dymNameAddress:  "a.b@.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - bad name",
			dymNameAddress:  "a.b.@.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - bad name",
			dymNameAddress:  ".b.a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - bad name",
			dymNameAddress:  ".b.a@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - empty input",
			dymNameAddress:  "",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:               "pass - allow hex address pattern",
			dymNameAddress:     "0x1234567890123456789012345678901234567890@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "0x1234567890123456789012345678901234567890",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - allow 32 bytes hex address pattern",
			dymNameAddress:     "0x1234567890123456789012345678901234567890123456789012345678901234@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "0x1234567890123456789012345678901234567890123456789012345678901234",
			wantChainIdOrAlias: "dym",
		},
		{
			name:            "fail - reject non-20 or 32 bytes hex address pattern, case 19 bytes",
			dymNameAddress:  "0x123456789012345678901234567890123456789@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - reject non-20 or 32 bytes hex address pattern, case 21 bytes",
			dymNameAddress:  "0x12345678901234567890123456789012345678901@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - reject non-20 or 32 bytes hex address pattern, case 31 bytes",
			dymNameAddress:  "0x123456789012345678901234567890123456789012345678901234567890123@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - reject non-20 or 32 bytes hex address pattern, case 33 bytes",
			dymNameAddress:  "0x12345678901234567890123456789012345678901234567890123456789012345@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:               "pass - allow valid bech32 address pattern",
			dymNameAddress:     "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "pass - allow valid bech32 address pattern, Interchain Account",
			dymNameAddress:     "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			wantChainIdOrAlias: "dym",
		},
		{
			name:            "fail - reject invalid bech32 address pattern",
			dymNameAddress:  "dym1zzzzzzzzzz69v7yszg69v7yszg69v7ys8xdv96@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "fail - reject invalid bech32 address pattern, Interchain Account",
			dymNameAddress:  "dym1zzzzzzzzzg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			gotSubName, gotDymName, gotChainIdOrAlias, err := dymnskeeper.ParseDymNameAddress(tt.dymNameAddress)
			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)

				// cross-check ResolveByDymNameAddress

				_, err2 := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, tt.dymNameAddress)
				s.Require().NotNil(err2, "when invalid address passed in, ResolveByDymNameAddress should return false")
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tt.wantSubName, gotSubName)
			s.Require().Equal(tt.wantDymName, gotDymName)
			s.Require().Equal(tt.wantChainIdOrAlias, gotChainIdOrAlias)
		})
	}
}

//goland:noinspection SpellCheckingInspection
func (s *KeeperTestSuite) TestKeeper_ReverseResolveDymNameAddress() {
	const rollAppId1 = "rollapp_1-1"
	const rollApp1Bech32 = "nim"
	const rollAppId2 = "rollapp_2-2"
	const rollApp2Bech32 = "man"
	const rollApp2Alias = "ral"

	ownerAcc := testAddr(1)
	anotherAcc := testAddr(2)
	icaAcc := testICAddr(3)

	tests := []struct {
		name            string
		dymNames        []dymnstypes.DymName
		additionalSetup func(*KeeperTestSuite)
		inputAddress    string
		workingChainId  string
		wantErr         bool
		wantErrContains string
		want            dymnstypes.ReverseResolvedDymNameAddresses
	}{
		{
			name: "pass - can resolve bech32 on host-chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - can resolve bech32 on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN(rollAppId1, "", ownerAcc.bech32C("ra")).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  rollAppId1,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - can resolve case-insensitive bech32 on host-chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(ownerAcc.bech32()),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - can resolve case-insensitive bech32 on Roll-App",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN(rollAppId1, "", ownerAcc.bech32C("ra")).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(ownerAcc.bech32C("ra")),
			workingChainId:  rollAppId1,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - case-sensitive resolve bech32 on non-host-chain/non-Roll-App",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  "blumbus_111-1",
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "bb",
					Name:           "a",
					ChainIdOrAlias: "blumbus_111-1",
				},
			},
		},
		{
			name: "pass - case-sensitive resolve bech32 on non-host-chain/non-Roll-App",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(ownerAcc.bech32()),
			workingChainId:  "blumbus_111-1",
			wantErr:         false,
			want:            nil,
		},
		{
			name: "pass - can resolve ICA bech32 on host-chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("", "ica", icaAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaAcc.bech32(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - can resolve ICA bech32 on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId1, "ica", icaAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaAcc.bech32(),
			workingChainId:  rollAppId1,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - can resolve case-insensitive ICA bech32 on host-chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("", "ica", icaAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(icaAcc.bech32()),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - can resolve case-insensitive ICA bech32 on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId1, "ica", icaAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(icaAcc.bech32()),
			workingChainId:  rollAppId1,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - case-sensitive resolve ICA bech32 on non-host-chain/non-RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("blumbus_111-1", "ica", icaAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaAcc.bech32(),
			workingChainId:  "blumbus_111-1",
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: "blumbus_111-1",
				},
			},
		},
		{
			name: "pass - case-sensitive resolve ICA bech32 on non-host-chain/non-RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("blumbus_111-1", "ica", icaAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(icaAcc.bech32()),
			workingChainId:  "blumbus_111-1",
			wantErr:         false,
			want:            nil,
		},

		{
			name: "pass - case-sensitive resolve other address on non-host-chain/non-RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("another", "", "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5").
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5",
			workingChainId:  "another",
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: "another",
				},
			},
		},
		{
			name: "pass - case-sensitive resolve other address on non-host-chain/non-RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("another", "", "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5").
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase("X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5"),
			workingChainId:  "another",
			wantErr:         false,
			want:            nil,
		},
		{
			name: "pass - only take records matching input chain-id",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  "blumbus_111-1",
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "bb",
					Name:           "a",
					ChainIdOrAlias: "blumbus_111-1",
				},
			},
		},
		{
			name: "pass - if no result, return empty without error",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    anotherAcc.bech32(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want:            nil,
		},
		{
			name: "pass - lookup by hex on host chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.hexStr(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - lookup by hex on host chain, uppercase address",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    strings.ToUpper(ownerAcc.hexStr()),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - lookup by hex on host chain, checksum address",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    common.BytesToAddress(ownerAcc.bytes()).String(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - lookup ICA by hex on host chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("", "ica", icaAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaAcc.hexStr(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - lookup by hex on RollApp with bech32 prefix mapped, find out the matching configuration",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId1, "ra", anotherAcc.bech32C(rollApp1Bech32)).
				buildSlice(),
			inputAddress:   anotherAcc.hexStr(),
			workingChainId: rollAppId1,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					// when bech32 found from mapped by chain-id,
					// we convert the hex address into bech32
					// and perform lookup, so we should find out
					// the existing configuration
					SubName:        "ra",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - lookup by hex on RollApp with bech32 prefix mapped, but matching configuration of corresponding address so we do fallback lookup",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: rollAppId1,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "", // fallback lookup does not have Path => SubName
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - lookup by hex on RollApp with bech32 prefix mapped, find out the matching configuration",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId1, "ra", ownerAcc.bech32C(rollApp1Bech32)).
				buildSlice(),
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: rollAppId1,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					// when bech32 found from mapped by chain-id,
					// we convert the hex address into bech32
					// and perform lookup, so we should find out
					// the existing configuration
					SubName:        "ra",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - lookup by hex on RollApp with bech32 prefix mapped, find out the matching configuration, even tho Chain-ID of RollApp in config is EIP-155 part only",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId1, "ra", ownerAcc.bech32C(rollApp1Bech32)).
				buildSlice(),
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: rollAppId1,
			additionalSetup: func(s *KeeperTestSuite) {
				dymName := s.dymNsKeeper.GetDymName(s.ctx, "a")
				s.Require().Len(dymName.Configs, 3)

				var found bool
				for _, cfg := range dymName.Configs {
					if cfg.ChainId == "1" /*EIP-155*/ && cfg.Path == "ra" {
						found = true
						break
					}
				}
				s.Require().True(found, "expected to find the configuration with Chain-ID is EIP-155 part")
			},
			wantErr: false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					// when bech32 found from mapped by chain-id,
					// we convert the hex address into bech32
					// and perform lookup, so we should find out
					// the existing configuration
					SubName:        "ra",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - skip lookup by hex after first try (direct match) if working-chain-id is Neither host-chain nor RollApp, by bech32",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "", ownerAcc.bech32()).
				buildSlice(),
			inputAddress:   anotherAcc.bech32(),
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - skip lookup by hex if working-chain-id is Neither host-chain nor RollApp, by hex",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "", ownerAcc.bech32()).
				buildSlice(),
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - find result from multiple Dym-Names matched, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - find result from multiple Dym-Names matched, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.hexStr(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - result is sorted",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "b",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - result not contains expired Dym-Name, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, -1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - result not contains expired Dym-Name, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, -1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.hexStr(),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name:            "fail - reject empty input address",
			dymNames:        newDN("a", ownerAcc.bech32()).buildSlice(),
			inputAddress:    "",
			workingChainId:  s.chainId,
			wantErr:         true,
			wantErrContains: "not supported address format",
		},
		{
			name:            "fail - reject bad input address",
			dymNames:        newDN("a", ownerAcc.bech32()).buildSlice(),
			inputAddress:    "0xdym1",
			workingChainId:  s.chainId,
			wantErr:         true,
			wantErrContains: "not supported address format",
		},
		{
			name:            "fail - reject empty working-chain-id",
			dymNames:        newDN("a", ownerAcc.bech32()).buildSlice(),
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  "",
			wantErr:         true,
			wantErrContains: "invalid chain-id format",
		},
		{
			name:            "fail - reject bad working-chain-id",
			dymNames:        newDN("a", ownerAcc.bech32()).buildSlice(),
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  "@",
			wantErr:         true,
			wantErrContains: "invalid chain-id format",
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
				newDN("c", anotherAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, ownerAcc.bech32(), "c")
				s.Require().NoError(err)
				err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, common.HexToAddress(ownerAcc.hexStr()).Bytes(), "c")
				s.Require().NoError(err)
			},
			inputAddress:   ownerAcc.bech32(),
			workingChainId: s.chainId,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
				newDN("c", anotherAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, ownerAcc.bech32(), "c")
				s.Require().NoError(err)
				err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, common.HexToAddress(ownerAcc.hexStr()).Bytes(), "c")
				s.Require().NoError(err)
			},
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: s.chainId,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherAcc.bech32(), "a")
				s.Require().NoError(err)
				err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, common.HexToAddress(anotherAcc.hexStr()).Bytes(), "a")
				s.Require().NoError(err)
			},
			inputAddress:   anotherAcc.bech32(),
			workingChainId: s.chainId,
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(s.now, +1).
					build(),
			},
			additionalSetup: func(s *KeeperTestSuite) {
				err := s.dymNsKeeper.AddReverseMappingConfiguredAddressToDymName(s.ctx, anotherAcc.bech32(), "a")
				s.Require().NoError(err)
				err = s.dymNsKeeper.AddReverseMappingFallbackAddressToDymName(s.ctx, common.HexToAddress(anotherAcc.hexStr()).Bytes(), "a")
				s.Require().NoError(err)
			},
			inputAddress:   anotherAcc.hexStr(),
			workingChainId: s.chainId,
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - matching by hex if bech32 is not found, on host chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32C(rollApp1Bech32),
			workingChainId:  s.chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: s.chainId,
				},
			},
		},
		{
			name: "pass - matching by hex if bech32 is not found, on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32C(rollApp1Bech32),
			workingChainId:  rollAppId1,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: rollAppId1,
				},
			},
		},
		{
			name: "pass - alias is used if available, by bech32, alias from params",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: func(s *KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: s.chainId,
						Aliases: []string{"dym", "dymension"},
					},
				}
				s.Require().NoError(s.dymNsKeeper.SetParams(s.ctx, moduleParams))
			},
			inputAddress:   ownerAcc.bech32(),
			workingChainId: s.chainId,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: "dym", // alias is used instead of chain-id
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: "dym",
				},
			},
		},
		{
			name: "pass - alias is used if available, by bech32, alias from RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				cfgNForRollApp(rollAppId2, "b", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(s *KeeperTestSuite) {
			},
			inputAddress:   ownerAcc.bech32C(rollApp2Bech32),
			workingChainId: rollAppId2,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: rollApp2Alias, // alias is used instead of chain-id
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: rollApp2Alias,
				},
			},
		},
		{
			name: "pass - alias is used if available, by hex, alias from params",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(s *KeeperTestSuite) {
				moduleParams := s.dymNsKeeper.GetParams(s.ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: s.chainId,
						Aliases: []string{"dym", "dymension"},
					},
				}
				s.Require().NoError(s.dymNsKeeper.SetParams(s.ctx, moduleParams))
			},
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: s.chainId,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: "dym", // alias is used instead of chain-id
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: "dym",
				},
			},
		},
		{
			name: "pass - alias is used if available, by hex, alias from RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgNForRollApp(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				cfgNForRollApp(rollAppId2, "b", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(s *KeeperTestSuite) {
			},
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: rollAppId2,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: rollApp2Alias, // alias is used instead of chain-id
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: rollApp2Alias,
				},
			},
		},
		{
			name: "pass - RollApp ID detected when config is EIP-155, alias is used if available",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(s.now, +1).
				cfgNForRollApp(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				cfgNForRollApp(rollAppId2, "b", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(s *KeeperTestSuite) {
				dymName := s.dymNsKeeper.GetDymName(s.ctx, "a")
				s.Require().Len(dymName.Configs, 2)

				const eip155Id = "2"
				for _, cfg := range dymName.Configs {
					s.Require().Equal(eip155Id, cfg.ChainId)
				}
			},
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: rollAppId2,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: rollApp2Alias, // alias is used instead of chain-id
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: rollApp2Alias,
				},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()

			s.persistRollApp(
				*newRollApp(rollAppId1).WithBech32(rollApp1Bech32),
				*newRollApp(rollAppId2).WithBech32(rollApp2Bech32).WithAlias(rollApp2Alias),
			)

			for _, dymName := range tt.dymNames {
				s.setDymNameWithFunctionsAfter(dymName)
			}

			if tt.additionalSetup != nil {
				tt.additionalSetup(s)
			}

			s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, rollAppId1), "bad-setup")

			got, err := s.dymNsKeeper.ReverseResolveDymNameAddress(s.ctx, tt.inputAddress, tt.workingChainId)
			if tt.wantErr {
				s.Require().NotEmpty(tt.wantErrContains, "mis-configured test case")
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tt.wantErrContains)
				return
			}

			s.Require().NoError(err)
			s.Require().Equal(tt.want, got)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_ReplaceChainIdWithAliasIfPossible() {
	moduleParams := s.dymNsKeeper.GetParams(s.ctx)
	moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
		{
			ChainId: s.chainId,
			Aliases: []string{"dym", "dymension"},
		},
		{
			ChainId: "blumbus_111-1",
			Aliases: []string{"bb"},
		},
		{
			ChainId: "froopyland_100-1",
			Aliases: nil,
		},
		{
			ChainId: "another-1",
			Aliases: []string{"another"},
		},
	}
	s.Require().NoError(s.dymNsKeeper.SetParams(s.ctx, moduleParams))

	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_1-1",
		Owner:     testAddr(1).bech32(),
	})
	s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, "rollapp_1-1"))
	s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_1-1", "ra1"))

	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_2-2",
		Owner:     testAddr(2).bech32(),
	})
	s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, "rollapp_2-2"))

	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_3-3",
		Owner:     testAddr(3).bech32(),
	})
	s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, "rollapp_3-3"))

	s.rollAppKeeper.SetRollapp(s.ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_4-4",
		Owner:     testAddr(4).bech32(),
	})
	s.Require().True(s.dymNsKeeper.IsRollAppId(s.ctx, "rollapp_4-4"))
	s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_4-4", "another"))

	s.Run("can replace from params", func() {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: s.chainId,
			},
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "blumbus_111-1",
			},
			{
				SubName:        "",
				Name:           "z",
				ChainIdOrAlias: "blumbus_111-1",
			},
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "froopyland_100-1",
			},
			{
				SubName:        "",
				Name:           "a",
				ChainIdOrAlias: "froopyland_100-1",
			},
		}

		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "dym",
				},
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "bb",
				},
				{
					SubName:        "",
					Name:           "z",
					ChainIdOrAlias: "bb",
				},
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "froopyland_100-1",
				},
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: "froopyland_100-1",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("ful-fill with host-chain-id if empty", func() {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "", // empty
			},
		}
		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "dym",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("mapping correct alias for RollApp by ID", func() {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "rollapp_1-1",
			},
			{
				Name:           "a",
				ChainIdOrAlias: "rollapp_2-2",
			},
			{
				Name:           "b",
				ChainIdOrAlias: "rollapp_3-3",
			},
		}
		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "ra1",
				},
				{
					Name:           "a",
					ChainIdOrAlias: "rollapp_2-2",
				},
				{
					Name:           "b",
					ChainIdOrAlias: "rollapp_3-3",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("mapping correct EIP-155 part to RollApp ID and alias", func() {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "1",
			},
			{
				Name:           "a",
				ChainIdOrAlias: "2",
			},
			{
				Name:           "b",
				ChainIdOrAlias: "3",
			},
		}
		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "ra1",
				},
				{
					Name:           "a",
					ChainIdOrAlias: "rollapp_2-2",
				},
				{
					Name:           "b",
					ChainIdOrAlias: "rollapp_3-3",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("mapping correct alias for RollApp by ID, when RollApp has multiple alias", func() {
		s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_1-1", "ral12"))
		s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_1-1", "ral13"))
		s.Require().NoError(s.dymNsKeeper.SetAliasForRollAppId(s.ctx, "rollapp_1-1", "ral14"))

		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "rollapp_1-1",
			},
		}
		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "ra1",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("mixed replacement from both params and RolApp alias", func() {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "rollapp_1-1",
			},
			{
				Name:           "a",
				ChainIdOrAlias: "rollapp_2-2",
			},
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "",
			},
			{
				SubName:        "a",
				Name:           "c",
				ChainIdOrAlias: s.chainId,
			},
		}
		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "ra1",
				},
				{
					Name:           "a",
					ChainIdOrAlias: "rollapp_2-2",
				},
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "dym",
				},
				{
					SubName:        "a",
					Name:           "c",
					ChainIdOrAlias: "dym",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("do not use Roll-App alias if occupied in Params", func() {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "rollapp_4-4",
			},
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "another-1",
			},
		}
		s.Require().Equal(
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "rollapp_4-4", // keep as is, even tho it has alias
				},
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "another",
				},
			},
			s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, input),
		)
	})

	s.Run("allow passing empty", func() {
		s.Require().Empty(s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, nil))
		s.Require().Empty(s.dymNsKeeper.ReplaceChainIdWithAliasIfPossible(s.ctx, []dymnstypes.ReverseResolvedDymNameAddress{}))
	})
}

func swapCase(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLower(r):
			return unicode.ToUpper(r)
		case unicode.IsUpper(r):
			return unicode.ToLower(r)
		}
		return r
	}, s)
}
