package keeper_test

import (
	"strings"
	"testing"
	"time"
	"unicode"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetSetDeleteDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

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

	setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

	t.Run("event should be fired", func(t *testing.T) {
		events := ctx.EventManager().Events()
		require.NotEmpty(t, events)

		for _, event := range events {
			if event.Type == dymnstypes.EventTypeSetDymName {
				return
			}
		}

		t.Errorf("event %s not found", dymnstypes.EventTypeSetDymName)
	})

	t.Run("Dym-Name should be equals to original", func(t *testing.T) {
		require.Equal(t, dymName, *dk.GetDymName(ctx, dymName.Name))
	})

	t.Run("delete", func(t *testing.T) {
		err := dk.DeleteDymName(ctx, dymName.Name)
		require.NoError(t, err)
		require.Nil(t, dk.GetDymName(ctx, dymName.Name))

		t.Run("reverse mapping should be deleted, check by key", func(t *testing.T) {
			ownedBy := dk.GenericGetReverseLookupDymNamesRecord(ctx,
				dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(ownerA)),
			)
			require.NoError(t, err)
			require.Empty(t, ownedBy, "reverse mapping should be removed")

			dymNames := dk.GenericGetReverseLookupDymNamesRecord(ctx,
				dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(ownerA),
			)
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")

			dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx,
				dymnstypes.FallbackAddressToDymNamesIncludeRvlKey(dymnstypes.FallbackAddress(sdk.MustAccAddressFromBech32(ownerA))),
			)
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")
		})

		t.Run("reverse mapping should be deleted, check by get", func(t *testing.T) {
			ownedBy, err := dk.GetDymNamesOwnedBy(ctx, ownerA)
			require.NoError(t, err)
			require.Empty(t, ownedBy, "reverse mapping should be removed")

			dymNames, err := dk.GetDymNamesContainsConfiguredAddress(ctx, ownerA)
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")

			dymNames, err = dk.GetDymNamesContainsFallbackAddress(ctx, sdk.MustAccAddressFromBech32(ownerA).Bytes())
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")
		})
	})

	t.Run("can not set invalid Dym-Name", func(t *testing.T) {
		require.Error(t, dk.SetDymName(ctx, dymnstypes.DymName{}))
	})

	t.Run("get returns nil if non-exists", func(t *testing.T) {
		require.Nil(t, dk.GetDymName(ctx, "non-exists"))
	})

	t.Run("delete a non-exists Dym-Name should be ok", func(t *testing.T) {
		err := dk.DeleteDymName(ctx, "non-exists")
		require.NoError(t, err)
	})
}

func TestKeeper_BeforeAfterDymNameOwnerChanged(t *testing.T) {
	t.Run("BeforeDymNameOwnerChanged can be called on non-existing Dym-Name without error", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		require.NoError(t, dk.BeforeDymNameOwnerChanged(ctx, "non-exists"))
	})

	t.Run("AfterDymNameOwnerChanged should returns error when calling on non-existing Dym-Name", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		err := dk.AfterDymNameOwnerChanged(ctx, "non-exists")
		require.Error(t, err)
		require.Contains(t, err.Error(), "Dym-Name: non-exists: not found")
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

	t.Run("BeforeDymNameOwnerChanged will remove the reverse mapping owned-by", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

		owned, err := dk.GetDymNamesOwnedBy(ctx, ownerA)
		require.NoError(t, err)
		require.Len(t, owned, 1)

		require.NoError(t, dk.BeforeDymNameOwnerChanged(ctx, dymName.Name))

		owned, err = dk.GetDymNamesOwnedBy(ctx, ownerA)
		require.NoError(t, err)
		require.Empty(t, owned)
	})

	t.Run("after run BeforeDymNameOwnerChanged, Dym-Name must be kept", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

		require.NoError(t, dk.BeforeDymNameOwnerChanged(ctx, dymName.Name))

		require.Equal(t, dymName, *dk.GetDymName(ctx, dymName.Name))
	})

	t.Run("AfterDymNameOwnerChanged will add the reverse mapping owned-by", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.NoError(t, dk.SetDymName(ctx, dymName))

		owned, err := dk.GetDymNamesOwnedBy(ctx, ownerA)
		require.NoError(t, err)
		require.Empty(t, owned)

		require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, dymName.Name))

		owned, err = dk.GetDymNamesOwnedBy(ctx, ownerA)
		require.NoError(t, err)
		require.Len(t, owned, 1)
	})

	t.Run("after run AfterDymNameOwnerChanged, Dym-Name must be kept", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.NoError(t, dk.SetDymName(ctx, dymName))

		require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, dymName.Name))

		require.Equal(t, dymName, *dk.GetDymName(ctx, dymName.Name))
	})
}

func TestKeeper_BeforeAfterDymNameConfigChanged(t *testing.T) {
	t.Run("BeforeDymNameConfigChanged can be called on non-existing Dym-Name without error", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		require.NoError(t, dk.BeforeDymNameConfigChanged(ctx, "non-exists"))
	})

	t.Run("AfterDymNameConfigChanged should returns error when calling on non-existing Dym-Name", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		err := dk.AfterDymNameConfigChanged(ctx, "non-exists")
		require.Error(t, err)
		require.Contains(t, err.Error(), "Dym-Name: non-exists: not found")
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

	requireConfiguredAddressMappedNoDymName := func(cfgAddr string, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, cfgAddr)
		require.NoError(t, err)
		require.Empty(t, names)
	}

	requireConfiguredAddressMappedDymName := func(cfgAddr string, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, cfgAddr)
		require.NoError(t, err)
		require.Len(t, names, 1)
		require.Equal(t, dymName.Name, names[0].Name)
	}

	requireFallbackAddressMappedNoDymName := func(addr dymnstypes.FallbackAddress, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsFallbackAddress(ctx, addr)
		require.NoError(t, err)
		require.Empty(t, names)
	}

	requireFallbackAddressMappedDymName := func(addr dymnstypes.FallbackAddress, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsFallbackAddress(ctx, addr)
		require.NoError(t, err)
		require.Len(t, names, 1)
		require.Equal(t, dymName.Name, names[0].Name)
	}

	t.Run("BeforeDymNameConfigChanged will remove the reverse mapping address", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		// do setup test

		setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

		requireConfiguredAddressMappedDymName(ownerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedDymName(controllerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedDymName(icaAcc.bech32(), ctx, dk)
		requireFallbackAddressMappedDymName(ownerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(controllerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(icaAcc.bytes(), ctx, dk)

		// do test

		require.NoError(t, dk.BeforeDymNameConfigChanged(ctx, dymName.Name))

		requireConfiguredAddressMappedNoDymName(ownerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedNoDymName(controllerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedNoDymName(icaAcc.bech32(), ctx, dk)
		requireFallbackAddressMappedNoDymName(ownerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(controllerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(icaAcc.bytes(), ctx, dk)
	})

	t.Run("after run BeforeDymNameConfigChanged, Dym-Name must be kept", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

		require.NoError(t, dk.BeforeDymNameConfigChanged(ctx, dymName.Name))

		require.Equal(t, dymName, *dk.GetDymName(ctx, dymName.Name))
	})

	t.Run("AfterDymNameConfigChanged will add the reverse mapping address", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		// do setup test

		require.NoError(t, dk.SetDymName(ctx, dymName))

		requireConfiguredAddressMappedNoDymName(ownerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedNoDymName(controllerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedNoDymName(icaAcc.bech32(), ctx, dk)
		requireFallbackAddressMappedNoDymName(ownerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(controllerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(icaAcc.bytes(), ctx, dk)

		// do test

		require.NoError(t, dk.AfterDymNameConfigChanged(ctx, dymName.Name))

		requireConfiguredAddressMappedDymName(ownerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedDymName(controllerAcc.bech32(), ctx, dk)
		requireConfiguredAddressMappedDymName(icaAcc.bech32(), ctx, dk)
		requireFallbackAddressMappedDymName(ownerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(controllerAcc.bytes(), ctx, dk)
		requireFallbackAddressMappedNoDymName(icaAcc.bytes(), ctx, dk)
	})

	t.Run("after run AfterDymNameConfigChanged, Dym-Name must be kept", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		require.NoError(t, dk.SetDymName(ctx, dymName))

		require.NoError(t, dk.AfterDymNameConfigChanged(ctx, dymName.Name))

		require.Equal(t, dymName, *dk.GetDymName(ctx, dymName.Name))
	})
}

func TestKeeper_GetDymNameWithExpirationCheck(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	t.Run("returns nil if not exists", func(t *testing.T) {
		require.Nil(t, dk.GetDymNameWithExpirationCheck(ctx, "non-exists"))
	})

	ownerA := testAddr(1).bech32()

	dymName := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: ownerA,
		}},
	}

	err := dk.SetDymName(ctx, dymName)
	require.NoError(t, err)

	t.Run("returns if not expired", func(t *testing.T) {
		require.NotNil(t, dk.GetDymNameWithExpirationCheck(ctx, dymName.Name))
	})

	t.Run("returns nil if expired", func(t *testing.T) {
		dymName.ExpireAt = ctx.BlockTime().Unix() - 1000
		require.NoError(t, dk.SetDymName(ctx, dymName))
		require.Nil(t, dk.GetDymNameWithExpirationCheck(
			ctx.WithBlockTime(time.Unix(dymName.ExpireAt+1, 0)), dymName.Name,
		))
	})
}

func TestKeeper_GetAllDymNamesAndNonExpiredDymNames(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()
	owner3a := testAddr(3).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   now.Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: owner1a,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))

	dymName2 := dymnstypes.DymName{
		Name:       "b",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   now.Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: owner2a,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	dymName3 := dymnstypes.DymName{
		Name:       "c",
		Owner:      owner3a,
		Controller: owner3a,
		ExpireAt:   now.Add(-time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Path:  "www",
			Value: owner3a,
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName3))

	listNonExpired := dk.GetAllNonExpiredDymNames(ctx)
	require.Len(t, listNonExpired, 2)
	require.Contains(t, listNonExpired, dymName1)
	require.Contains(t, listNonExpired, dymName2)
	require.NotContains(t, listNonExpired, dymName3, "should not include expired Dym-Name")

	listAll := dk.GetAllDymNames(ctx)
	require.Len(t, listAll, 3)
	require.Contains(t, listAll, dymName1)
	require.Contains(t, listAll, dymName2)
	require.Contains(t, listAll, dymName3, "should include expired Dym-Name")
}

func TestKeeper_GetDymNamesOwnedBy(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	owner1a := testAddr(1).bech32()
	owner2a := testAddr(2).bech32()

	dymName11 := dymnstypes.DymName{
		Name:       "n11",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName11, t, dk)

	dymName12 := dymnstypes.DymName{
		Name:       "n12",
		Owner:      owner1a,
		Controller: owner1a,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName12, t, dk)

	dymName21 := dymnstypes.DymName{
		Name:       "n21",
		Owner:      owner2a,
		Controller: owner2a,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName21, t, dk)

	t.Run("returns owned Dym-Names", func(t *testing.T) {
		ownedBy, err := dk.GetDymNamesOwnedBy(ctx, owner1a)
		require.NoError(t, err)
		requireDymNameList(ownedBy, []string{dymName11.Name, dymName12.Name}, t)
	})

	t.Run("returns owned Dym-Names with filtered expiration", func(t *testing.T) {
		dymName12.ExpireAt = now.Add(-time.Hour).Unix()
		setDymNameWithFunctionsAfter(ctx, dymName12, t, dk)

		ownedBy, err := dk.GetDymNamesOwnedBy(ctx, owner1a)
		require.NoError(t, err)
		requireDymNameList(ownedBy, []string{dymName11.Name}, t)
	})
}

func TestKeeper_PruneDymName(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	t.Run("prune non-exists Dym-Name should be ok", func(t *testing.T) {
		require.NoError(t, dk.PruneDymName(ctx, "non-exists"))
	})

	ownerA := testAddr(1).bech32()

	dymName1 := dymnstypes.DymName{
		Name:       "a",
		Owner:      ownerA,
		Controller: ownerA,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}

	t.Run("able to prune non-expired Dym-Name", func(t *testing.T) {
		setDymNameWithFunctionsAfter(ctx, dymName1, t, dk)
		require.NotNil(t, dk.GetDymName(ctx, dymName1.Name))

		require.NoError(t, dk.PruneDymName(ctx, dymName1.Name))
		require.Nil(t, dk.GetDymName(ctx, dymName1.Name))
	})

	// re-setup record
	setDymNameWithFunctionsAfter(ctx, dymName1, t, dk)
	require.NotNil(t, dk.GetDymName(ctx, dymName1.Name))
	owned, err := dk.GetDymNamesOwnedBy(ctx, dymName1.Owner)
	require.NoError(t, err)
	require.Len(t, owned, 1)

	// setup historical SO
	expiredSo := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, expiredSo)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, expiredSo.AssetId, expiredSo.AssetType)
	require.NoError(t, err)
	require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName), 1)
	minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, dymName1.Name, dymnstypes.TypeName)
	require.True(t, found)
	require.Equal(t, expiredSo.ExpireAt, minExpiry)

	// setup active SO
	so := dymnstypes.SellOrder{
		AssetId:   dymName1.Name,
		AssetType: dymnstypes.TypeName,
		ExpireAt:  now.Add(time.Hour).Unix(),
		MinPrice:  dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so)
	require.NoError(t, err)
	require.NotNil(t, dk.GetSellOrder(ctx, dymName1.Name, dymnstypes.TypeName))

	// prune
	err = dk.PruneDymName(ctx, dymName1.Name)
	require.NoError(t, err)

	require.Nil(t, dk.GetDymName(ctx, dymName1.Name), "Dym-Name should be removed")

	owned, err = dk.GetDymNamesOwnedBy(ctx, dymName1.Owner)
	require.NoError(t, err)
	require.Empty(t, owned, "reserve mapping should be removed")

	require.Nil(t, dk.GetSellOrder(ctx, dymName1.Name, dymnstypes.TypeName), "active SO should be removed")

	require.Empty(t,
		dk.GetHistoricalSellOrders(ctx, dymName1.Name, dymnstypes.TypeName),
		"historical SO should be removed",
	)

	_, found = dk.GetMinExpiryHistoricalSellOrder(ctx, dymName1.Name, dymnstypes.TypeName)
	require.False(t, found)
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_ResolveByDymNameAddress(t *testing.T) {
	now := time.Now().UTC()

	const chainId = "dymension_1100-1"

	setupTest := func() (dymnskeeper.Keeper, rollappkeeper.Keeper, sdk.Context) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		return dk, rk, ctx
	}

	addr1a := testAddr(1).bech32()

	addr2Acc := testAddr(2)
	addr2a := addr2Acc.bech32()

	addr3a := testAddr(3).bech32()

	generalSetupAlias := func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
		params := dk.GetParams(ctx)
		params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: chainId,
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
		err := dk.SetParams(ctx, params)
		require.NoError(t, err)
	}

	tests := []struct {
		name              string
		dymName           *dymnstypes.DymName
		preSetup          func(sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
		dymNameAddress    string
		wantError         bool
		wantErrContains   string
		wantOutputAddress string
		postTest          func(sdk.Context, dymnskeeper.Keeper, rollappkeeper.Keeper)
	}{
		{
			name: "success, no sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
			name: "success, no sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "c.b.a.dym")
				require.NoError(t, err)
				require.Equal(t, addr3a, outputAddr)
			},
		},
		{
			name: "success, multi-sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr1a, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "a@dym")
				require.Error(t, err)
				require.Contains(t, err.Error(), "no resolution found")

				_, err = dk.ResolveByDymNameAddress(ctx, "non-exists.a@dymension_1100-1")
				require.Error(t, err)
				require.Contains(t, err.Error(), "no resolution found")
			},
		},
		{
			name: "success, multi-sub config, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a.dym")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dym")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@dym")
				require.NoError(t, err)
				require.Equal(t, addr1a, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "non-exists.a@dym")
				require.Error(t, err)
				require.Contains(t, err.Error(), "no resolution found")
			},
		},
		{
			name: "success, alias of RollApp",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "nim_1122-1",
					Value:   addr2Acc.bech32C("nim"),
				}, {
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "nim_1122-1",
					Path:    "b",
					Value:   addr2Acc.bech32C("nim"),
				}},
			},
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "nim_1122-1", "nim", "nim")
			},
			dymNameAddress:    "a@nim",
			wantOutputAddress: addr2Acc.bech32C("nim"),
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				// should be able to resolve if multiple aliases attached to the same RollApp

				aliases := []string{"nim1", "nim2", "nim3"}
				for _, alias := range aliases {
					require.NoError(t, dk.SetAliasForRollAppId(ctx, "nim_1122-1", alias))
				}

				for _, alias := range aliases {
					outputAddr, err := dk.ResolveByDymNameAddress(ctx, "a@"+alias)
					require.NoError(t, err)
					require.Equal(t, addr2Acc.bech32C("nim"), outputAddr)

					outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@"+alias)
					require.NoError(t, err)
					require.Equal(t, addr2Acc.bech32C("nim"), outputAddr)
				}
			},
		},
		{
			name: "lookup through multiple sub-domains",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				dymNameB := dymnstypes.DymName{
					Name:       "b",
					Owner:      addr1a,
					Controller: addr2a,
					ExpireAt:   now.Unix() + 1,
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
				require.NoError(t, dk.SetDymName(ctx, dymNameB))
			},
			dymNameAddress:    "b.a.dymension_1100-1",
			wantOutputAddress: addr3a,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.b.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)
			},
		},
		{
			name: "matching by chain-id, no alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.blumbus_111-1")
				require.NoError(t, err)
				require.Equal(t, addr3a, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "a@bb")
				require.Error(t, err)

				_, err = dk.ResolveByDymNameAddress(ctx, "a@blumbus")
				require.Error(t, err)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "a.dym")
				require.Error(t, err)

				_, err = dk.ResolveByDymNameAddress(ctx, "a.dymension")
				require.Error(t, err)
			},
		},
		{
			name: "matching by chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.blumbus_111-1")
				require.NoError(t, err)
				require.Equal(t, addr3a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@bb")
				require.NoError(t, err)
				require.Equal(t, addr3a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@blumbus")
				require.NoError(t, err)
				require.Equal(t, addr3a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dym")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dymension")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)
			},
		},
		{
			name: "not configured sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				list, err := dk.ReverseResolveDymNameAddress(ctx, "Ae2tdPwUPEZFSi1cTyL1ZL6bgixhc2vSy5heg6Zg9uP7PpumkAJ82Qprt8b", "another")
				require.NoError(t, err)
				require.Len(t, list, 1)
				require.Equal(t, "sub2.a@another", list[0].String())

				list, err = dk.ReverseResolveDymNameAddress(ctx, "X-avax1tzdcgj4ehsvhhgpl7zylwpw0gl2rxcg4r5afk5", "another")
				require.NoError(t, err)
				require.Len(t, list, 1)
				require.Equal(t, "sub1.a@another", list[0].String())
			},
		},
		{
			name: "resolve to owner when no default (without sub-name) Dym-Name config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "sub.a.dym")
				require.NoError(t, err)
				require.Equal(t, addr3a, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "non-exists.a.dym")
				require.Error(t, err)
				require.Contains(t, err.Error(), "no resolution found")

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@bb")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)
			},
		},
		{
			name: "do not fallback for sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() + 1,
				Configs:    nil,
			},
			dymNameAddress:  "sub.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "no resolution found",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "a.dymension_1100-1")
				require.NoError(t, err, "should fallback if not sub-name")
				require.Equal(t, addr1a, outputAddr)
			},
		},
		{
			name: "should not resolve for expired Dym-Name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr2a,
				ExpireAt:   now.Unix() - 1,
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
				ExpireAt:   now.Unix() - 1,
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
				ExpireAt:   now.Unix() + 1,
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
				ExpireAt:   now.Unix() + 1,
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
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				params := dk.GetParams(ctx)
				params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "blumbus_111-1",
						Aliases: []string{"blumbus"},
					},
				}
				err := dk.SetParams(ctx, params)
				require.NoError(t, err)
			},
			dymNameAddress:    "a.blumbus",
			wantError:         false,
			wantOutputAddress: addr3a,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "a@blumbus_111-1")
				require.NoError(t, err)
				require.Equal(t, addr2a, outputAddr)
			},
		},
		{
			name:              "resolve extra format 0x1234...6789@dym",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "0x1234567890123456789012345678901234567890.dym")
				require.NoError(t, err)
				require.Equal(t, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96", outputAddr)
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
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				params := dk.GetParams(ctx)
				params.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: "blumbus_111-1",
						Aliases: []string{"blumbus"},
					},
				}
				err := dk.SetParams(ctx, params)
				require.NoError(t, err)
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "0x1234567890123456789012345678901234567890123456789012345678901234.dym")
				require.NoError(t, err)
				require.Equal(t, "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul", outputAddr)
			},
		},
		{
			name:              "resolve extra format nim1...@dym, cross bech32 format",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9.dym")
				require.NoError(t, err)
				require.Equal(t, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96", outputAddr)
			},
		},
		{
			name: "fallback resolve follow default config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1a,
				Controller: addr1a,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_DCT_NAME,
						Value: addr2Acc.bech32(),
					},
				},
			},
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "nim_1122-1", "nim", "nim")
			},
			dymNameAddress:    "a@nim",
			wantError:         false,
			wantOutputAddress: addr2Acc.bech32C("nim"),
			postTest:          nil,
		},
		{
			name:    "resolve extra format 0x1234...6789@nim (RollApp)",
			dymName: nil,
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "nim_1122-1", "nim", "nim")
			},
			dymNameAddress:    "0x1234567890123456789012345678901234567890@nim_1122-1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "0x1234567890123456789012345678901234567890.nim")
				require.NoError(t, err)
				require.Equal(t, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "resolve extra format 0x1234...6789@nim1 (RollApp), alternative alias among multiple aliases for RollApp",
			dymName: nil,
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "nim_1122-1", "nim", "nim")
				require.NoError(t, dk.SetAliasForRollAppId(ctx, "nim_1122-1", "nim1"))
				require.NoError(t, dk.SetAliasForRollAppId(ctx, "nim_1122-1", "nim2"))
			},
			dymNameAddress:    "0x1234567890123456789012345678901234567890@nim1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "0x1234567890123456789012345678901234567890.nim")
				require.NoError(t, err)
				require.Equal(t, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "resolve extra format dym1...@nim (RollApp), cross bech32 format",
			dymName: nil,
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "nim_1122-1", "nim", "nim")
			},
			dymNameAddress:    "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@nim_1122-1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96.nim")
				require.NoError(t, err)
				require.Equal(t, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "resolve extra format dym1...@nim1 (RollApp), cross bech32 format, alternative alias among multiple aliases for RollApp",
			dymName: nil,
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "nim_1122-1", "nim", "nim")
				require.NoError(t, dk.SetAliasForRollAppId(ctx, "nim_1122-1", "nim1"))
				require.NoError(t, dk.SetAliasForRollAppId(ctx, "nim_1122-1", "nim2"))
			},
			dymNameAddress:    "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@nim1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper, _ rollappkeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96.nim")
				require.NoError(t, err)
				require.Equal(t, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			name:    "try resolve extra format dym1...@rollapp, cross bech32 format, but RollApp does not have bech32 configured",
			dymName: nil,
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper, rk rollappkeeper.Keeper) {
				registerRollApp(t, ctx, rk, dk, "rollapp_1-1", "" /*no bech32*/, "")
			},
			dymNameAddress:  "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@rollapp_1-1",
			wantError:       true,
			wantErrContains: "not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, rk, ctx := setupTest()

			if tt.preSetup != nil {
				tt.preSetup(ctx, dk, rk)
			}

			if tt.dymName != nil {
				setDymNameWithFunctionsAfter(ctx, *tt.dymName, t, dk)
			}

			outputAddress, err := dk.ResolveByDymNameAddress(ctx, tt.dymNameAddress)

			defer func() {
				if t.Failed() {
					return
				}

				if tt.postTest != nil {
					tt.postTest(ctx, dk, rk)
				}
			}()

			if tt.wantError {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOutputAddress, outputAddress)
		})
	}

	t.Run("mixed tests", func(t *testing.T) {
		dk, rk, ctx := setupTest()

		bech32Addr := func(no uint64) string {
			return testAddr(no).bech32()
		}

		// setup alias
		moduleParams := dk.GetParams(ctx)
		moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
			{
				ChainId: chainId,
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
		require.NoError(t, dk.SetParams(ctx, moduleParams))

		// setup Dym-Names
		dymName1 := dymnstypes.DymName{
			Name:       "name1",
			Owner:      bech32Addr(1),
			Controller: bech32Addr(2),
			ExpireAt:   now.Unix() + 1,
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
		require.NoError(t, dk.SetDymName(ctx, dymName1))

		dymName2 := dymnstypes.DymName{
			Name:       "name2",
			Owner:      bech32Addr(100),
			Controller: bech32Addr(101),
			ExpireAt:   now.Unix() + 1,
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
		require.NoError(t, dk.SetDymName(ctx, dymName2))

		dymName3 := dymnstypes.DymName{
			Name:       "name3",
			Owner:      bech32Addr(200),
			Controller: bech32Addr(201),
			ExpireAt:   now.Unix() + 1,
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
		require.NoError(t, dk.SetDymName(ctx, dymName3))

		dymName4 := dymnstypes.DymName{
			Name:       "name4",
			Owner:      "dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn",
			Controller: bech32Addr(301),
			ExpireAt:   now.Unix() + 1,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_DCT_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   bech32Addr(302),
				},
			},
		}
		require.NoError(t, dk.SetDymName(ctx, dymName4))

		rollAppNim := rollapptypes.Rollapp{
			RollappId: "nim_1122-1",
			Owner:     bech32Addr(1122),
		}
		registerRollApp(t, ctx, rk, dk, rollAppNim.RollappId, "nim", "nim")
		rollAppNim, found := rk.GetRollapp(ctx, rollAppNim.RollappId)
		require.True(t, found)

		tc := func(name, chainIdOrAlias string) input {
			return newInputTestcase(name, chainIdOrAlias, ctx, dk, t)
		}

		tc("name1", chainId).WithSubName("s1").RequireResolveTo(bech32Addr(3))
		tc("name1", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(3))
		tc("name1", chainId).WithSubName("s2").RequireResolveTo(bech32Addr(4))
		tc("name1", "dym").WithSubName("s2").RequireResolveTo(bech32Addr(4))
		tc("name1", chainId).WithSubName("a.s5").RequireResolveTo(bech32Addr(5))
		tc("name1", "dym").WithSubName("a.s5").RequireResolveTo(bech32Addr(5))
		tc("name1", chainId).WithSubName("none").RequireNotResolve()
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

		tc("name2", chainId).WithSubName("s1").RequireResolveTo(bech32Addr(103))
		tc("name2", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(103))
		tc("name2", chainId).WithSubName("s2").RequireResolveTo(bech32Addr(104))
		tc("name2", "dym").WithSubName("s2").RequireResolveTo(bech32Addr(104))
		tc("name2", chainId).WithSubName("a.s5").RequireResolveTo(bech32Addr(105))
		tc("name2", "dym").WithSubName("a.s5").RequireResolveTo(bech32Addr(105))
		tc("name2", chainId).WithSubName("none").RequireNotResolve()
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

		tc("name3", chainId).WithSubName("s1").RequireResolveTo(bech32Addr(203))
		tc("name3", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(203))
		tc("name3", chainId).WithSubName("s2").RequireResolveTo(bech32Addr(204))
		tc("name3", "dym").WithSubName("s2").RequireResolveTo(bech32Addr(204))
		tc("name3", chainId).WithSubName("a.s5").RequireResolveTo(bech32Addr(205))
		tc("name3", "dym").WithSubName("a.s5").RequireResolveTo(bech32Addr(205))
		tc("name3", chainId).WithSubName("none").RequireNotResolve()
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

		tc("name4", chainId).WithSubName("s1").RequireResolveTo(bech32Addr(302))
		tc("name4", "dym").WithSubName("s1").RequireResolveTo(bech32Addr(302))
		tc("name4", chainId).WithSubName("none").RequireNotResolve()
		tc("name4", "dym").WithSubName("none").RequireNotResolve()
		tc("name4", chainId).RequireResolveTo("dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn")
		tc("name4", "dym").RequireResolveTo("dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn")
		tc("name4", rollAppNim.RollappId).RequireResolveTo(
			"nim1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq8wkcvv",
		)
	})
}

type input struct {
	t   *testing.T
	ctx sdk.Context
	dk  dymnskeeper.Keeper
	//
	name           string
	chainIdOrAlias string
	subName        string
}

func newInputTestcase(name, chainIdOrAlias string, ctx sdk.Context, dk dymnskeeper.Keeper, t *testing.T) input {
	return input{name: name, chainIdOrAlias: chainIdOrAlias, ctx: ctx, dk: dk, t: t}
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
		_, err := m.dk.ResolveByDymNameAddress(m.ctx, dymNameAddr)
		require.Error(m.t, err)
	}
}

func (m input) RequireResolveTo(wantAddr string) {
	for _, dymNameAddr := range m.buildDymNameAddrsCases() {
		gotAddr, err := m.dk.ResolveByDymNameAddress(m.ctx, dymNameAddr)
		require.NoError(m.t, err)
		require.Equal(m.t, wantAddr, gotAddr)
	}
}

//goland:noinspection SpellCheckingInspection
func Test_ParseDymNameAddress(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			gotSubName, gotDymName, gotChainIdOrAlias, err := dymnskeeper.ParseDymNameAddress(tt.dymNameAddress)
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)

				// cross-check ResolveByDymNameAddress
				dk, _, _, ctx := testkeeper.DymNSKeeper(t)
				_, err2 := dk.ResolveByDymNameAddress(ctx, tt.dymNameAddress)
				require.NotNil(t, err2, "when invalid address passed in, ResolveByDymNameAddress should return false")
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantSubName, gotSubName)
			require.Equal(t, tt.wantDymName, gotDymName)
			require.Equal(t, tt.wantChainIdOrAlias, gotChainIdOrAlias)
		})
	}
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_ReverseResolveDymNameAddress(t *testing.T) {
	now := time.Now().UTC()
	const chainId = "dymension_1100-1"
	const rollAppId1 = "rollapp_1-1"
	const rollApp1Bech32 = "nim"
	const rollAppId2 = "rollapp_2-2"
	const rollApp2Bech32 = "man"
	const rollApp2Alias = "ral"

	ownerAcc := testAddr(1)
	anotherAcc := testAddr(2)
	icaAcc := testICAddr(3)

	setupTest := func() (dymnskeeper.Keeper, rollappkeeper.Keeper, sdk.Context) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		registerRollApp(t, ctx, rk, dk, rollAppId1, rollApp1Bech32, "")
		registerRollApp(t, ctx, rk, dk, rollAppId2, rollApp2Bech32, rollApp2Alias)

		return dk, rk, ctx
	}

	tests := []struct {
		name            string
		dymNames        []dymnstypes.DymName
		additionalSetup func(ctx sdk.Context, dk dymnskeeper.Keeper)
		inputAddress    string
		workingChainId  string
		wantErr         bool
		wantErrContains string
		want            dymnstypes.ReverseResolvedDymNameAddresses
	}{
		{
			name: "pass - can resolve bech32 on host-chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - can resolve bech32 on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(ownerAcc.bech32()),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - can resolve case-insensitive bech32 on Roll-App",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
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
				exp(now, +1).
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
				exp(now, +1).
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("", "ica", icaAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaAcc.bech32(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - can resolve ICA bech32 on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN(rollAppId1, "ica", icaAcc.bech32()).
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("", "ica", icaAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    swapCase(icaAcc.bech32()),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - can resolve case-insensitive ICA bech32 on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN(rollAppId1, "ica", icaAcc.bech32()).
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
				exp(now, +1).
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
				exp(now, +1).
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
				exp(now, +1).
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
				exp(now, +1).
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
				exp(now, +1).
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
				exp(now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    anotherAcc.bech32(),
			workingChainId:  chainId,
			wantErr:         false,
			want:            nil,
		},
		{
			name: "pass - lookup by hex on host chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.hexStr(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - lookup by hex on host chain, uppercase address",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    strings.ToUpper(ownerAcc.hexStr()),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - lookup by hex on host chain, checksum address",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    common.BytesToAddress(ownerAcc.bytes()).String(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - lookup ICA by hex on host chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("", "ica", icaAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaAcc.hexStr(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "ica",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - lookup by hex on RollApp with bech32 prefix mapped, find out the matching configuration",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgN(rollAppId1, "ra", anotherAcc.bech32C(rollApp1Bech32)).
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
				exp(now, +1).
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgN(rollAppId1, "ra", ownerAcc.bech32C(rollApp1Bech32)).
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
			name: "pass - skip lookup by hex after first try (direct match) if working-chain-id is Neither host-chain nor RollApp, by bech32",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
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
				exp(now, +1).
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
					exp(now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - find result from multiple Dym-Names matched, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.hexStr(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - result is sorted",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "b",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - result not contains expired Dym-Name, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, -1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - result not contains expired Dym-Name, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, -1).
					cfgN("", "b", ownerAcc.bech32()).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerAcc.hexStr(),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name:            "fail - reject empty input address",
			dymNames:        newDN("a", ownerAcc.bech32()).buildSlice(),
			inputAddress:    "",
			workingChainId:  chainId,
			wantErr:         true,
			wantErrContains: "not supported address format",
		},
		{
			name:            "fail - reject bad input address",
			dymNames:        newDN("a", ownerAcc.bech32()).buildSlice(),
			inputAddress:    "0xdym1",
			workingChainId:  chainId,
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
					exp(now, +1).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					build(),
				newDN("c", anotherAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, ownerAcc.bech32(), "c")
				require.NoError(t, err)
				err = dk.AddReverseMappingFallbackAddressToDymName(ctx, common.HexToAddress(ownerAcc.hexStr()).Bytes(), "c")
				require.NoError(t, err)
			},
			inputAddress:   ownerAcc.bech32(),
			workingChainId: chainId,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, +1).
					build(),
				newDN("b", ownerAcc.bech32()).
					exp(now, +1).
					build(),
				newDN("c", anotherAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, ownerAcc.bech32(), "c")
				require.NoError(t, err)
				err = dk.AddReverseMappingFallbackAddressToDymName(ctx, common.HexToAddress(ownerAcc.hexStr()).Bytes(), "c")
				require.NoError(t, err)
			},
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: chainId,
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
				{
					SubName:        "",
					Name:           "b",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherAcc.bech32(), "a")
				require.NoError(t, err)
				err = dk.AddReverseMappingFallbackAddressToDymName(ctx, common.HexToAddress(anotherAcc.hexStr()).Bytes(), "a")
				require.NoError(t, err)
			},
			inputAddress:   anotherAcc.bech32(),
			workingChainId: chainId,
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", ownerAcc.bech32()).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherAcc.bech32(), "a")
				require.NoError(t, err)
				err = dk.AddReverseMappingFallbackAddressToDymName(ctx, common.HexToAddress(anotherAcc.hexStr()).Bytes(), "a")
				require.NoError(t, err)
			},
			inputAddress:   anotherAcc.hexStr(),
			workingChainId: chainId,
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - matching by hex if bech32 is not found, on host chain",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerAcc.bech32C(rollApp1Bech32),
			workingChainId:  chainId,
			wantErr:         false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "",
					Name:           "a",
					ChainIdOrAlias: chainId,
				},
			},
		},
		{
			name: "pass - matching by hex if bech32 is not found, on RollApp",
			dymNames: newDN("a", ownerAcc.bech32()).
				exp(now, +1).
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				buildSlice(),
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: chainId,
						Aliases: []string{"dym", "dymension"},
					},
				}
				require.NoError(t, dk.SetParams(ctx, moduleParams))
			},
			inputAddress:   ownerAcc.bech32(),
			workingChainId: chainId,
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgN(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				cfgN(rollAppId2, "b", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgN(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
					{
						ChainId: chainId,
						Aliases: []string{"dym", "dymension"},
					},
				}
				require.NoError(t, dk.SetParams(ctx, moduleParams))
			},
			inputAddress:   ownerAcc.hexStr(),
			workingChainId: chainId,
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
				exp(now, +1).
				cfgN("", "b", ownerAcc.bech32()).
				cfgN("blumbus_111-1", "bb", ownerAcc.bech32()).
				cfgN(rollAppId2, "", ownerAcc.bech32C(rollApp2Bech32)).
				cfgN(rollAppId2, "b", ownerAcc.bech32C(rollApp2Bech32)).
				buildSlice(),
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
		t.Run(tt.name, func(t *testing.T) {
			dk, _, ctx := setupTest()

			for _, dymName := range tt.dymNames {
				setDymNameWithFunctionsAfter(ctx, dymName, t, dk)
			}

			if tt.additionalSetup != nil {
				tt.additionalSetup(ctx, dk)
			}

			require.True(t, dk.IsRollAppId(ctx, rollAppId1), "bad-setup")

			got, err := dk.ReverseResolveDymNameAddress(ctx, tt.inputAddress, tt.workingChainId)
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestKeeper_ReplaceChainIdWithAliasIfPossible(t *testing.T) {
	dk, _, rk, ctx := testkeeper.DymNSKeeper(t)

	const chainId = "dymension_1100-1"
	ctx = ctx.WithChainID(chainId)

	moduleParams := dk.GetParams(ctx)
	moduleParams.Chains.AliasesOfChainIds = []dymnstypes.AliasesOfChainId{
		{
			ChainId: chainId,
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
	require.NoError(t, dk.SetParams(ctx, moduleParams))

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_1-1",
		Owner:     testAddr(1).bech32(),
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_1-1"))
	require.NoError(t, dk.SetAliasForRollAppId(ctx, "rollapp_1-1", "ra1"))

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_2-2",
		Owner:     testAddr(2).bech32(),
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_2-2"))

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_3-3",
		Owner:     testAddr(3).bech32(),
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_3-3"))

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_4-4",
		Owner:     testAddr(4).bech32(),
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_4-4"))
	require.NoError(t, dk.SetAliasForRollAppId(ctx, "rollapp_4-4", "another"))

	t.Run("can replace from params", func(t *testing.T) {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: chainId,
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

		require.Equal(t,
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
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("ful-fill with host-chain-id if empty", func(t *testing.T) {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "", // empty
			},
		}
		require.Equal(t,
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "dym",
				},
			},
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("mapping correct alias for RollApp by ID", func(t *testing.T) {
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
		require.Equal(t,
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
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("mapping correct alias for RollApp by ID, when RollApp has multiple alias", func(t *testing.T) {
		require.NoError(t, dk.SetAliasForRollAppId(ctx, "rollapp_1-1", "ral12"))
		require.NoError(t, dk.SetAliasForRollAppId(ctx, "rollapp_1-1", "ral13"))
		require.NoError(t, dk.SetAliasForRollAppId(ctx, "rollapp_1-1", "ral14"))

		input := []dymnstypes.ReverseResolvedDymNameAddress{
			{
				SubName:        "a",
				Name:           "b",
				ChainIdOrAlias: "rollapp_1-1",
			},
		}
		require.Equal(t,
			[]dymnstypes.ReverseResolvedDymNameAddress{
				{
					SubName:        "a",
					Name:           "b",
					ChainIdOrAlias: "ra1",
				},
			},
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("mixed replacement from both params and RolApp alias", func(t *testing.T) {
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
				ChainIdOrAlias: chainId,
			},
		}
		require.Equal(t,
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
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("do not use Roll-App alias if occupied in Params", func(t *testing.T) {
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
		require.Equal(t,
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
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("allow passing empty", func(t *testing.T) {
		require.Empty(t, dk.ReplaceChainIdWithAliasIfPossible(ctx, nil))
		require.Empty(t, dk.ReplaceChainIdWithAliasIfPossible(ctx, []dymnstypes.ReverseResolvedDymNameAddress{}))
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
