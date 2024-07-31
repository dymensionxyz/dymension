package keeper_test

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetSetDeleteDymName(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_NAME,
			Path:  "www",
			Value: owner,
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
				dymnstypes.DymNamesOwnedByAccountRvlKey(sdk.MustAccAddressFromBech32(owner)),
			)
			require.NoError(t, err)
			require.Empty(t, ownedBy, "reverse mapping should be removed")

			dymNames := dk.GenericGetReverseLookupDymNamesRecord(ctx,
				dymnstypes.ConfiguredAddressToDymNamesIncludeRvlKey(owner),
			)
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")

			dymNames = dk.GenericGetReverseLookupDymNamesRecord(ctx,
				dymnstypes.HexAddressToDymNamesIncludeRvlKey(sdk.MustAccAddressFromBech32(owner)),
			)
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")
		})

		t.Run("reverse mapping should be deleted, check by get", func(t *testing.T) {
			ownedBy, err := dk.GetDymNamesOwnedBy(ctx, owner)
			require.NoError(t, err)
			require.Empty(t, ownedBy, "reverse mapping should be removed")

			dymNames, err := dk.GetDymNamesContainsConfiguredAddress(ctx, owner)
			require.NoError(t, err)
			require.Empty(t, dymNames, "reverse mapping should be removed")

			dymNames, err = dk.GetDymNamesContainsHexAddress(ctx, sdk.MustAccAddressFromBech32(owner).Bytes())
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

//goland:noinspection SpellCheckingInspection
func TestKeeper_BeforeAfterDymNameOwnerChanged(t *testing.T) {
	t.Run("BeforeDymNameOwnerChanged can be called on non-existing Dym-Name without error", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		require.NoError(t, dk.BeforeDymNameOwnerChanged(ctx, "non-exists"))
	})

	t.Run("AfterDymNameOwnerChanged should returns error when calling on non-existing Dym-Name", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		err := dk.AfterDymNameOwnerChanged(ctx, "non-exists")
		require.Error(t, err)
		require.Contains(t, err.Error(), dymnstypes.ErrDymNameNotFound.Error())
	})

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   time.Now().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_NAME,
			Path:  "www",
			Value: owner,
		}},
	}

	t.Run("BeforeDymNameOwnerChanged will remove the reverse mapping owned-by", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

		owned, err := dk.GetDymNamesOwnedBy(ctx, owner)
		require.NoError(t, err)
		require.Len(t, owned, 1)

		require.NoError(t, dk.BeforeDymNameOwnerChanged(ctx, dymName.Name))

		owned, err = dk.GetDymNamesOwnedBy(ctx, owner)
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

		owned, err := dk.GetDymNamesOwnedBy(ctx, owner)
		require.NoError(t, err)
		require.Empty(t, owned)

		require.NoError(t, dk.AfterDymNameOwnerChanged(ctx, dymName.Name))

		owned, err = dk.GetDymNamesOwnedBy(ctx, owner)
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

//goland:noinspection SpellCheckingInspection
func TestKeeper_BeforeAfterDymNameConfigChanged(t *testing.T) {
	t.Run("BeforeDymNameConfigChanged can be called on non-existing Dym-Name without error", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		require.NoError(t, dk.BeforeDymNameConfigChanged(ctx, "non-exists"))
	})

	t.Run("AfterDymNameConfigChanged should returns error when calling on non-existing Dym-Name", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)
		err := dk.AfterDymNameConfigChanged(ctx, "non-exists")
		require.Error(t, err)
		require.Contains(t, err.Error(), dymnstypes.ErrDymNameNotFound.Error())
	})

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	ownerHex := sdk.MustAccAddressFromBech32(owner).Bytes()
	const controller = "dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv"
	controllerHex := sdk.MustAccAddressFromBech32(controller).Bytes()
	const ica = "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul"
	icaHex := sdk.MustAccAddressFromBech32(ica).Bytes()

	dymName := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: controller,
		ExpireAt:   time.Now().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{
			{
				Type:  dymnstypes.DymNameConfigType_NAME,
				Path:  "controller",
				Value: controller,
			}, {
				Type:  dymnstypes.DymNameConfigType_NAME,
				Path:  "ica",
				Value: ica,
			},
		},
	}

	requireConfiguredAddressMappedNoDymName := func(bech32Addr string, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, bech32Addr)
		require.NoError(t, err)
		require.Empty(t, names)
	}

	requireConfiguredAddressMappedDymName := func(bech32Addr string, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsConfiguredAddress(ctx, bech32Addr)
		require.NoError(t, err)
		require.Len(t, names, 1)
		require.Equal(t, dymName.Name, names[0].Name)
	}

	requireHexAddressMappedNoDymName := func(addr []byte, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsHexAddress(ctx, addr)
		require.NoError(t, err)
		require.Empty(t, names)
	}

	requireHexAddressMappedDymName := func(addr []byte, ctx sdk.Context, dk dymnskeeper.Keeper) {
		names, err := dk.GetDymNamesContainsHexAddress(ctx, addr)
		require.NoError(t, err)
		require.Len(t, names, 1)
		require.Equal(t, dymName.Name, names[0].Name)
	}

	t.Run("BeforeDymNameConfigChanged will remove the reverse mapping address", func(t *testing.T) {
		dk, _, _, ctx := testkeeper.DymNSKeeper(t)

		// do setup test

		setDymNameWithFunctionsAfter(ctx, dymName, t, dk)

		requireConfiguredAddressMappedDymName(owner, ctx, dk)
		requireConfiguredAddressMappedDymName(controller, ctx, dk)
		requireConfiguredAddressMappedDymName(ica, ctx, dk)
		requireHexAddressMappedDymName(ownerHex, ctx, dk)
		requireHexAddressMappedNoDymName(controllerHex, ctx, dk)
		requireHexAddressMappedNoDymName(icaHex, ctx, dk)

		// do test

		require.NoError(t, dk.BeforeDymNameConfigChanged(ctx, dymName.Name))

		requireConfiguredAddressMappedNoDymName(owner, ctx, dk)
		requireConfiguredAddressMappedNoDymName(controller, ctx, dk)
		requireConfiguredAddressMappedNoDymName(ica, ctx, dk)
		requireHexAddressMappedNoDymName(ownerHex, ctx, dk)
		requireHexAddressMappedNoDymName(controllerHex, ctx, dk)
		requireHexAddressMappedNoDymName(icaHex, ctx, dk)
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

		requireConfiguredAddressMappedNoDymName(owner, ctx, dk)
		requireConfiguredAddressMappedNoDymName(controller, ctx, dk)
		requireConfiguredAddressMappedNoDymName(ica, ctx, dk)
		requireHexAddressMappedNoDymName(ownerHex, ctx, dk)
		requireHexAddressMappedNoDymName(controllerHex, ctx, dk)
		requireHexAddressMappedNoDymName(icaHex, ctx, dk)

		// do test

		require.NoError(t, dk.AfterDymNameConfigChanged(ctx, dymName.Name))

		requireConfiguredAddressMappedDymName(owner, ctx, dk)
		requireConfiguredAddressMappedDymName(controller, ctx, dk)
		requireConfiguredAddressMappedDymName(ica, ctx, dk)
		requireHexAddressMappedDymName(ownerHex, ctx, dk)
		requireHexAddressMappedNoDymName(controllerHex, ctx, dk)
		requireHexAddressMappedNoDymName(icaHex, ctx, dk)
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

	//goland:noinspection SpellCheckingInspection
	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Unix() + 1,
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_NAME,
			Path:  "www",
			Value: owner,
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

func TestKeeper_GetAllNonExpiredDymNames(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	//goland:noinspection SpellCheckingInspection
	dymName1 := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_NAME,
			Path:  "www",
			Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName1))

	//goland:noinspection SpellCheckingInspection
	dymName2 := dymnstypes.DymName{
		Name:       "not-bonded-pool",
		Owner:      "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
		Controller: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
		ExpireAt:   time.Now().UTC().Add(time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_NAME,
			Path:  "www",
			Value: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName2))

	//goland:noinspection SpellCheckingInspection
	dymName3 := dymnstypes.DymName{
		Name:       "streamer",
		Owner:      "dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv",
		Controller: "dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv",
		ExpireAt:   time.Now().UTC().Add(-time.Hour).Unix(),
		Configs: []dymnstypes.DymNameConfig{{
			Type:  dymnstypes.DymNameConfigType_NAME,
			Path:  "www",
			Value: "dym1ysjlrjcankjpmpxxzk27mvzhv25e266r80p5pv",
		}},
	}
	require.NoError(t, dk.SetDymName(ctx, dymName3))

	list := dk.GetAllNonExpiredDymNames(ctx)
	require.Len(t, list, 2)
	require.Contains(t, list, dymName1)
	require.Contains(t, list, dymName2)
	require.NotContains(t, list, dymName3, "should not include expired Dym-Name")
}

//goland:noinspection SpellCheckingInspection
func TestKeeper_GetDymNamesOwnedBy(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName1 := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName1, t, dk)

	dymName2 := dymnstypes.DymName{
		Name:       "a",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Add(-time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName2, t, dk)

	dymName3 := dymnstypes.DymName{
		Name:       "b",
		Owner:      "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
		Controller: "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4",
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName3, t, dk)

	t.Run("returns owned Dym-Names with filtered expiration", func(t *testing.T) {
		ownedBy, err := dk.GetDymNamesOwnedBy(ctx, owner)
		require.NoError(t, err)
		require.Len(t, ownedBy, 1)
		require.Equal(t, owner, ownedBy[0].Owner)
		require.Equal(t, dymName1, ownedBy[0])
	})
}

func TestKeeper_PruneDymName(t *testing.T) {
	now := time.Now().UTC()

	dk, _, _, ctx := testkeeper.DymNSKeeper(t)
	ctx = ctx.WithBlockTime(now)

	require.NoError(t, dk.PruneDymName(ctx, "non-exists"))

	//goland:noinspection SpellCheckingInspection
	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"

	dymName1 := dymnstypes.DymName{
		Name:       "bonded-pool",
		Owner:      owner,
		Controller: owner,
		ExpireAt:   now.Add(time.Hour).Unix(),
	}
	setDymNameWithFunctionsAfter(ctx, dymName1, t, dk)
	require.NotNil(t, dk.GetDymName(ctx, dymName1.Name))

	t.Run("able to prune non-expired Dym-Name", func(t *testing.T) {
		require.NoError(t, dk.PruneDymName(ctx, dymName1.Name))
		require.Nil(t, dk.GetDymName(ctx, dymName1.Name))
	})

	setDymNameWithFunctionsAfter(ctx, dymName1, t, dk)
	require.NotNil(t, dk.GetDymName(ctx, dymName1.Name))
	owned, err := dk.GetDymNamesOwnedBy(ctx, dymName1.Owner)
	require.NoError(t, err)
	require.Len(t, owned, 1)

	// setup historical SO
	expiredSo := dymnstypes.SellOrder{
		Name:      dymName1.Name,
		ExpireAt:  1,
		MinPrice:  dymnsutils.TestCoin(100),
		SellPrice: dymnsutils.TestCoinP(300),
	}
	err = dk.SetSellOrder(ctx, expiredSo)
	require.NoError(t, err)
	err = dk.MoveSellOrderToHistorical(ctx, expiredSo.Name)
	require.NoError(t, err)
	require.Len(t, dk.GetHistoricalSellOrders(ctx, dymName1.Name), 1)
	minExpiry, found := dk.GetMinExpiryHistoricalSellOrder(ctx, dymName1.Name)
	require.True(t, found)
	require.Equal(t, expiredSo.ExpireAt, minExpiry)

	// setup active SO
	so := dymnstypes.SellOrder{
		Name:     dymName1.Name,
		ExpireAt: now.Add(time.Hour).Unix(),
		MinPrice: dymnsutils.TestCoin(100),
	}
	err = dk.SetSellOrder(ctx, so)
	require.NoError(t, err)
	require.NotNil(t, dk.GetSellOrder(ctx, dymName1.Name))

	// prune
	err = dk.PruneDymName(ctx, dymName1.Name)
	require.NoError(t, err)

	require.Nil(t, dk.GetDymName(ctx, dymName1.Name), "Dym-Name should be removed")

	owned, err = dk.GetDymNamesOwnedBy(ctx, dymName1.Owner)
	require.NoError(t, err)
	require.Empty(t, owned, "reserve mapping should be removed")

	require.Nil(t, dk.GetSellOrder(ctx, dymName1.Name), "active SO should be removed")

	require.Empty(t,
		dk.GetHistoricalSellOrders(ctx, dymName1.Name),
		"historical SO should be removed",
	)

	_, found = dk.GetMinExpiryHistoricalSellOrder(ctx, dymName1.Name)
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

	addr1 := "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	addr2 := "dym1gtcunp63a3aqypr250csar4devn8fjpqulq8d4"
	addr3 := "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d"

	generalSetupAlias := func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
			{
				ChainId: "nim_1122-1",
				Aliases: []string{"nim"},
			},
		}
		err := dk.SetParams(ctx, params)
		require.NoError(t, err)
	}

	tests := []struct {
		name              string
		dymName           *dymnstypes.DymName
		preSetup          func(sdk.Context, dymnskeeper.Keeper)
		dymNameAddress    string
		wantError         bool
		wantErrContains   string
		wantOutputAddress string
		postTest          func(sdk.Context, dymnskeeper.Keeper)
	}{
		{
			name: "success, no sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: addr3,
				}},
			},
			dymNameAddress:    "a.dymension_1100-1",
			wantOutputAddress: addr3,
		},
		{
			name: "success, no sub name, chain-id, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: addr3,
				}},
			},
			dymNameAddress:    "a@dymension_1100-1",
			wantOutputAddress: addr3,
		},
		{
			name: "success, sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr3,
				}},
			},
			dymNameAddress:    "b.a.dymension_1100-1",
			wantOutputAddress: addr3,
		},
		{
			name: "success, sub name, chain-id, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr3,
				}},
			},
			dymNameAddress:    "b.a@dymension_1100-1",
			wantOutputAddress: addr3,
		},
		{
			name: "success, multi-sub name, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}},
			},
			dymNameAddress:    "c.b.a.dymension_1100-1",
			wantOutputAddress: addr3,
		},
		{
			name: "success, multi-sub name, chain-id, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}},
			},
			dymNameAddress:    "c.b.a@dymension_1100-1",
			wantOutputAddress: addr3,
		},
		{
			name: "success, no sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a.dym",
			wantOutputAddress: addr3,
		},
		{
			name: "success, no sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a@dym",
			wantOutputAddress: addr3,
		},
		{
			name: "success, sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "b.a.dym",
			wantOutputAddress: addr3,
		},
		{
			name: "success, sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "b.a@dym",
			wantOutputAddress: addr3,
		},
		{
			name: "success, multi-sub name, alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a.dym",
			wantOutputAddress: addr3,
		},
		{
			name: "success, match multiple alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr2,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a.dymension",
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "c.b.a.dym")
				require.NoError(t, err)
				require.Equal(t, addr3, outputAddr)
			},
		},
		{
			name: "success, multi-sub name, alias, @",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a@dym",
			wantOutputAddress: addr3,
		},
		{
			name: "success, multi-sub config, chain-id",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr2,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr1,
				}},
			},
			preSetup:          nil,
			dymNameAddress:    "c.b.a.dymension_1100-1",
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr1, outputAddr)

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
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr2,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr1,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "c.b.a@dym",
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a.dym")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.a@dym")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@dym")
				require.NoError(t, err)
				require.Equal(t, addr1, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "non-exists.a@dym")
				require.Error(t, err)
				require.Contains(t, err.Error(), "no resolution found")
			},
		},
		{
			name: "lookup through multiple sub-domains",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "b",
					Value: addr3,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr3,
				}},
			},
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				dymNameB := dymnstypes.DymName{
					Name:       "b",
					Owner:      addr1,
					Controller: addr2,
					ExpireAt:   now.Unix() + 1,
					Configs: []dymnstypes.DymNameConfig{{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "b",
						Value: addr2,
					}, {
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "",
						Value: addr2,
					}},
				}
				require.NoError(t, dk.SetDymName(ctx, dymNameB))
			},
			dymNameAddress:    "b.a.dymension_1100-1",
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b@dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "b.b.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)
			},
		},
		{
			name: "matching by chain-id, no alias",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "b",
					Value:   addr2,
				}, {
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   addr2,
				}, {
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr3,
				}, {
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "",
					Value:   addr3,
				}},
			},
			dymNameAddress:    "a.blumbus_111-1",
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.blumbus_111-1")
				require.NoError(t, err)
				require.Equal(t, addr3, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "a@bb")
				require.Error(t, err)

				_, err = dk.ResolveByDymNameAddress(ctx, "a@blumbus")
				require.Error(t, err)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

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
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "b",
					Value:   addr2,
				}, {
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   addr2,
				}, {
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr3,
				}, {
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "",
					Value:   addr3,
				}},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a.blumbus_111-1",
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				var outputAddr string
				var err error

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.blumbus_111-1")
				require.NoError(t, err)
				require.Equal(t, addr3, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@bb")
				require.NoError(t, err)
				require.Equal(t, addr3, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@blumbus")
				require.NoError(t, err)
				require.Equal(t, addr3, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dymension_1100-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dym")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a.dymension")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)
			},
		},
		{
			name: "not configured sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "c.b",
					Value: addr3,
				}, {
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr2,
				}},
			},
			dymNameAddress:  "b.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "no resolution found",
		},
		{
			name: "when no Dym-Name does not exists",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr3,
				}},
			},
			dymNameAddress:  "b@dym",
			wantError:       true,
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
		},
		{
			name: "resolve to owner when no Dym-Name config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs:    nil,
			},
			dymNameAddress:    "a.dymension_1100-1",
			wantError:         false,
			wantOutputAddress: addr1,
		},
		{
			name: "resolve to owner when no default (without sub-name) Dym-Name config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Path:  "sub",
						Value: addr3,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_111-1",
						Path:    "",
						Value:   addr2,
					},
				},
			},
			preSetup:          generalSetupAlias,
			dymNameAddress:    "a.dymension_1100-1",
			wantError:         false,
			wantOutputAddress: addr1,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "sub.a.dym")
				require.NoError(t, err)
				require.Equal(t, addr3, outputAddr)

				_, err = dk.ResolveByDymNameAddress(ctx, "non-exists.a.dym")
				require.Error(t, err)
				require.Contains(t, err.Error(), "no resolution found")

				outputAddr, err = dk.ResolveByDymNameAddress(ctx, "a@bb")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)
			},
		},
		{
			name: "do not fallback for sub-name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs:    nil,
			},
			dymNameAddress:  "sub.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: "no resolution found",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "a.dymension_1100-1")
				require.NoError(t, err, "should fallback if not sub-name")
				require.Equal(t, addr1, outputAddr)
			},
		},
		{
			name: "should not resolve for expired Dym-Name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() - 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr3,
				}},
			},
			dymNameAddress:  "a.dymension_1100-1",
			wantError:       true,
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
		},
		{
			name: "should not resolve for expired Dym-Name",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() - 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr3,
				}},
			},
			dymNameAddress:  "a.a.dymension_1100-1",
			wantError:       true,
			wantErrContains: dymnstypes.ErrDymNameNotFound.Error(),
		},
		{
			name: "should not resolve if input addr is invalid",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{{
					Type:  dymnstypes.DymNameConfigType_NAME,
					Path:  "",
					Value: addr3,
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
				Owner:      addr1,
				Controller: addr2,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus_111-1",
						Path:    "",
						Value:   addr2,
					},
					{
						Type:    dymnstypes.DymNameConfigType_NAME,
						ChainId: "blumbus",
						Path:    "",
						Value:   addr3,
					},
				},
			},
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
			wantOutputAddress: addr3,
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "a@blumbus_111-1")
				require.NoError(t, err)
				require.Equal(t, addr2, outputAddr)
			},
		},
		{
			name:              "resolve extra format 0x1234...6789@dym",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
			wantErrContains:   dymnstypes.ErrDymNameNotFound.Error(),
			wantOutputAddress: "",
		},
		{
			name:    "resolve extra format 0x1234...6789@dym, do not resolve if chain-id is not RollApp, even tho alias was defined",
			dymName: nil,
			preSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
			wantErrContains:   dymnstypes.ErrDymNameNotFound.Error(),
			wantOutputAddress: "",
		},
		{
			name:              "resolve extra format 0x1234...6789@dym, Interchain Account",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890123456789012345678901234@dymension_1100-1",
			wantError:         false,
			wantOutputAddress: "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
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
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9.dym")
				require.NoError(t, err)
				require.Equal(t, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96", outputAddr)
			},
		},
		{
			// must resolve to address with nim prefix of corresponding address from default-config
			// TODO DymNS FIXME: resolve to rollapp based address using bech32 prefix.
			// This testcase is failed atm.
			name: "FIXME * fallback resolve follow default config",
			dymName: &dymnstypes.DymName{
				Name:       "a",
				Owner:      addr1,
				Controller: addr1,
				ExpireAt:   now.Unix() + 1,
				Configs: []dymnstypes.DymNameConfig{
					{
						Type:  dymnstypes.DymNameConfigType_NAME,
						Value: addr2,
					},
				},
			},
			preSetup:       nil,
			dymNameAddress: "a@nim",
			wantError:      false,
			wantOutputAddress: func() string {
				_, bz, err := bech32.DecodeAndConvert(addr2)
				require.NoError(t, err)
				addr2OnNim, err := bech32.ConvertAndEncode("nim", bz)
				require.NoError(t, err)
				return addr2OnNim
			}(),
			postTest: nil,
		},
		{
			// must resolve to address with nim prefix
			// TODO DymNS FIXME: resolve to rollapp based address using bech32 prefix.
			// This testcase is failed atm.
			name:              "FIXME * resolve extra format 0x1234...6789@nim (RollApp)",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "0x1234567890123456789012345678901234567890@nim_1122-1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "0x1234567890123456789012345678901234567890.nim")
				require.NoError(t, err)
				require.Equal(t, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
		{
			// must resolve to address with nim prefix
			// TODO DymNS FIXME: resolve to rollapp based address using bech32 prefix.
			// This testcase is failed atm.
			name:              "FIXME * resolve extra format dym1...@nim (RollApp), cross bech32 format",
			dymName:           nil,
			preSetup:          generalSetupAlias,
			dymNameAddress:    "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@nim_1122-1",
			wantError:         false,
			wantOutputAddress: "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9",
			postTest: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				outputAddr, err := dk.ResolveByDymNameAddress(ctx, "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96.nim")
				require.NoError(t, err)
				require.Equal(t, "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9", outputAddr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, _, ctx := setupTest()

			if tt.preSetup != nil {
				tt.preSetup(ctx, dk)
			}

			if tt.dymName != nil {
				require.NoError(t, dk.SetDymName(ctx, *tt.dymName))
			}

			outputAddress, err := dk.ResolveByDymNameAddress(ctx, tt.dymNameAddress)

			defer func() {
				if t.Failed() {
					return
				}

				if tt.postTest != nil {
					tt.postTest(ctx, dk)
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

	t.Run("FIXME * mixed tests", func(t *testing.T) {
		dk, rk, ctx := setupTest()

		addr := func(no uint64) string {
			bz1 := sdk.Uint64ToBigEndian(no)
			bz2 := make([]byte, 20)
			copy(bz2, bz1)
			return sdk.MustBech32ifyAddressBytes(params.AccountAddressPrefix, bz2)
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
			Owner:      addr(1),
			Controller: addr(2),
			ExpireAt:   now.Unix() + 1,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   addr(3),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s2",
					Value:   addr(4),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "a.s5",
					Value:   addr(5),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr(6),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "c.b",
					Value:   addr(7),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "juno-1",
					Path:    "",
					Value:   addr(8),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "juno-1",
					Path:    "a.b.c",
					Value:   addr(9),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "cosmoshub-4",
					Path:    "",
					Value:   addr(10),
				},
			},
		}
		require.NoError(t, dk.SetDymName(ctx, dymName1))

		dymName2 := dymnstypes.DymName{
			Name:       "name2",
			Owner:      addr(100),
			Controller: addr(101),
			ExpireAt:   now.Unix() + 1,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   addr(103),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s2",
					Value:   addr(104),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "a.s5",
					Value:   addr(105),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr(106),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "c.b",
					Value:   addr(107),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "juno-1",
					Path:    "",
					Value:   addr(108),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "juno-1",
					Path:    "a.b.c",
					Value:   addr(109),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "froopyland_100-1",
					Path:    "a",
					Value:   addr(110),
				},
			},
		}
		require.NoError(t, dk.SetDymName(ctx, dymName2))

		dymName3 := dymnstypes.DymName{
			Name:       "name3",
			Owner:      addr(200),
			Controller: addr(201),
			ExpireAt:   now.Unix() + 1,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   addr(203),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s2",
					Value:   addr(204),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "a.s5",
					Value:   addr(205),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "b",
					Value:   addr(206),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "c.b",
					Value:   addr(207),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "juno-1",
					Path:    "",
					Value:   addr(208),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "juno-1",
					Path:    "a.b.c",
					Value:   addr(209),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "froopyland_100-1",
					Path:    "a",
					Value:   addr(210),
				},
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "cosmoshub-4",
					Path:    "a",
					Value:   addr(211),
				},
			},
		}
		require.NoError(t, dk.SetDymName(ctx, dymName3))

		dymName4 := dymnstypes.DymName{
			Name:       "name4",
			Owner:      "dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn",
			Controller: addr(301),
			ExpireAt:   now.Unix() + 1,
			Configs: []dymnstypes.DymNameConfig{
				{
					Type:    dymnstypes.DymNameConfigType_NAME,
					ChainId: "",
					Path:    "s1",
					Value:   addr(302),
				},
			},
		}
		require.NoError(t, dk.SetDymName(ctx, dymName4))

		rollAppNim := rollapptypes.Rollapp{
			RollappId: "nim_1122-1",
			Creator:   addr(1122),
		}
		rk.SetRollapp(ctx, rollAppNim)
		rollAppNim, found := rk.GetRollapp(ctx, rollAppNim.RollappId)
		require.True(t, found)

		tc := func(name, chainIdOrAlias string) input {
			return newInputTestcase(name, chainIdOrAlias, ctx, dk, t)
		}

		tc("name1", chainId).WithSubName("s1").RequireResolveTo(addr(3))
		tc("name1", "dym").WithSubName("s1").RequireResolveTo(addr(3))
		tc("name1", chainId).WithSubName("s2").RequireResolveTo(addr(4))
		tc("name1", "dym").WithSubName("s2").RequireResolveTo(addr(4))
		tc("name1", chainId).WithSubName("a.s5").RequireResolveTo(addr(5))
		tc("name1", "dym").WithSubName("a.s5").RequireResolveTo(addr(5))
		tc("name1", chainId).WithSubName("none").RequireNotResolve()
		tc("name1", "dym").WithSubName("none").RequireNotResolve()
		tc("name1", "blumbus_111-1").WithSubName("b").RequireResolveTo(addr(6))
		tc("name1", "bb").WithSubName("b").RequireResolveTo(addr(6))
		tc("name1", "blumbus_111-1").WithSubName("c.b").RequireResolveTo(addr(7))
		tc("name1", "bb").WithSubName("c.b").RequireResolveTo(addr(7))
		tc("name1", "blumbus_111-1").WithSubName("none").RequireNotResolve()
		tc("name1", "bb").WithSubName("none").RequireNotResolve()
		tc("name1", "juno-1").RequireResolveTo(addr(8))
		tc("name1", "juno-1").WithSubName("a.b.c").RequireResolveTo(addr(9))
		tc("name1", "juno-1").WithSubName("none").RequireNotResolve()
		tc("name1", "cosmoshub-4").RequireResolveTo(addr(10))
		tc("name1", "cosmos").RequireResolveTo(addr(10))

		tc("name2", chainId).WithSubName("s1").RequireResolveTo(addr(103))
		tc("name2", "dym").WithSubName("s1").RequireResolveTo(addr(103))
		tc("name2", chainId).WithSubName("s2").RequireResolveTo(addr(104))
		tc("name2", "dym").WithSubName("s2").RequireResolveTo(addr(104))
		tc("name2", chainId).WithSubName("a.s5").RequireResolveTo(addr(105))
		tc("name2", "dym").WithSubName("a.s5").RequireResolveTo(addr(105))
		tc("name2", chainId).WithSubName("none").RequireNotResolve()
		tc("name2", "dym").WithSubName("none").RequireNotResolve()
		tc("name2", "blumbus_111-1").WithSubName("b").RequireResolveTo(addr(106))
		tc("name2", "bb").WithSubName("b").RequireResolveTo(addr(106))
		tc("name2", "blumbus_111-1").WithSubName("c.b").RequireResolveTo(addr(107))
		tc("name2", "bb").WithSubName("c.b").RequireResolveTo(addr(107))
		tc("name2", "blumbus_111-1").WithSubName("none").RequireNotResolve()
		tc("name2", "bb").WithSubName("none").RequireNotResolve()
		tc("name2", "juno-1").RequireResolveTo(addr(108))
		tc("name2", "juno-1").WithSubName("a.b.c").RequireResolveTo(addr(109))
		tc("name2", "juno-1").WithSubName("none").RequireNotResolve()
		tc("name2", "froopyland_100-1").WithSubName("a").RequireResolveTo(addr(110))
		tc("name2", "froopyland").WithSubName("a").RequireNotResolve()
		tc("name2", "cosmoshub-4").RequireNotResolve()
		tc("name2", "cosmoshub-4").WithSubName("a").RequireNotResolve()

		tc("name3", chainId).WithSubName("s1").RequireResolveTo(addr(203))
		tc("name3", "dym").WithSubName("s1").RequireResolveTo(addr(203))
		tc("name3", chainId).WithSubName("s2").RequireResolveTo(addr(204))
		tc("name3", "dym").WithSubName("s2").RequireResolveTo(addr(204))
		tc("name3", chainId).WithSubName("a.s5").RequireResolveTo(addr(205))
		tc("name3", "dym").WithSubName("a.s5").RequireResolveTo(addr(205))
		tc("name3", chainId).WithSubName("none").RequireNotResolve()
		tc("name3", "dym").WithSubName("none").RequireNotResolve()
		tc("name3", "blumbus_111-1").WithSubName("b").RequireResolveTo(addr(206))
		tc("name3", "bb").WithSubName("b").RequireResolveTo(addr(206))
		tc("name3", "blumbus_111-1").WithSubName("c.b").RequireResolveTo(addr(207))
		tc("name3", "bb").WithSubName("c.b").RequireResolveTo(addr(207))
		tc("name3", "blumbus_111-1").WithSubName("none").RequireNotResolve()
		tc("name3", "bb").WithSubName("none").RequireNotResolve()
		tc("name3", "juno-1").RequireResolveTo(addr(208))
		tc("name3", "juno-1").WithSubName("a.b.c").RequireResolveTo(addr(209))
		tc("name3", "juno-1").WithSubName("none").RequireNotResolve()
		tc("name3", "froopyland_100-1").WithSubName("a").RequireResolveTo(addr(210))
		tc("name3", "froopyland").WithSubName("a").RequireNotResolve()
		tc("name3", "cosmoshub-4").RequireNotResolve()
		tc("name3", "cosmos").WithSubName("a").RequireResolveTo(addr(211))

		tc("name4", chainId).WithSubName("s1").RequireResolveTo(addr(302))
		tc("name4", "dym").WithSubName("s1").RequireResolveTo(addr(302))
		tc("name4", chainId).WithSubName("none").RequireNotResolve()
		tc("name4", "dym").WithSubName("none").RequireNotResolve()
		tc("name4", chainId).RequireResolveTo("dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn")
		tc("name4", "dym").RequireResolveTo("dym1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqp7vezn")
		tc("name4", rollAppNim.RollappId).RequireResolveTo(
			// must resolve to owner with nim prefix
			// TODO DymNS FIXME: resolve to rollapp based address using bech32 prefix.
			// This testcase is failed atm.
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
			name:               "valid, no sub-name, chain-id, @",
			dymNameAddress:     "a@dymension_1100-1",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "valid, no sub-name, chain-id",
			dymNameAddress:     "a.dymension_1100-1",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "valid, no sub-name, alias, @",
			dymNameAddress:     "a@dym",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "valid, no sub-name, alias",
			dymNameAddress:     "a.dym",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "valid, sub-name, chain-id, @",
			dymNameAddress:     "b.a@dymension_1100-1",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "valid, sub-name, chain-id",
			dymNameAddress:     "b.a.dymension_1100-1",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "valid, sub-name, alias, @",
			dymNameAddress:     "b.a@dym",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "valid, sub-name, alias",
			dymNameAddress:     "b.a.dym",
			wantSubName:        "b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "valid, multi-sub-name, chain-id, @",
			dymNameAddress:     "c.b.a@dymension_1100-1",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "valid, multi-sub-name, chain-id",
			dymNameAddress:     "c.b.a.dymension_1100-1",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dymension_1100-1",
		},
		{
			name:               "valid, multi-sub-name, alias, @",
			dymNameAddress:     "c.b.a@dym",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "valid, multi-sub-name, alias",
			dymNameAddress:     "c.b.a.dym",
			wantSubName:        "c.b",
			wantDymName:        "a",
			wantChainIdOrAlias: "dym",
		},
		{
			name:            "invalid '.' after '@', no sub-name",
			dymNameAddress:  "a@dymension_1100-1.dym",
			wantErr:         true,
			wantErrContains: "misplaced '.'",
		},
		{
			name:            "invalid '.' after '@', sub-name",
			dymNameAddress:  "a.b@dymension_1100-1.dym",
			wantErr:         true,
			wantErrContains: "misplaced '.'",
		},
		{
			name:            "invalid '.' after '@', multi-sub-name",
			dymNameAddress:  "a.b.c@dymension_1100-1.dym",
			wantErr:         true,
			wantErrContains: "misplaced '.'",
		},
		{
			name:            "missing chain-id/alias, @",
			dymNameAddress:  "a@",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "missing chain-id/alias",
			dymNameAddress:  "a",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "missing chain-id/alias",
			dymNameAddress:  "a.",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "not accept space, no sub-name",
			dymNameAddress:  "a .dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "not accept space, sub-name",
			dymNameAddress:  "b .a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "not accept space, multi-sub-name",
			dymNameAddress:  "c.b .a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "invalid chain-id/alias, @",
			dymNameAddress:  "a@-dym",
			wantErr:         true,
			wantErrContains: "chain-id/alias is not well-formed",
		},
		{
			name:            "invalid chain-id/alias",
			dymNameAddress:  "a.-dym",
			wantErr:         true,
			wantErrContains: "chain-id/alias is not well-formed",
		},
		{
			name:            "invalid Dym-Name, @",
			dymNameAddress:  "-a@dym",
			wantErr:         true,
			wantErrContains: "Dym-Name is not well-formed",
		},
		{
			name:            "invalid Dym-Name",
			dymNameAddress:  "-a.dym",
			wantErr:         true,
			wantErrContains: "Dym-Name is not well-formed",
		},
		{
			name:            "invalid sub-Dym-Name, @",
			dymNameAddress:  "-b.a@dym",
			wantErr:         true,
			wantErrContains: "sub-Dym-Name is not well-formed",
		},
		{
			name:            "invalid sub-Dym-Name",
			dymNameAddress:  "-b.a.dym",
			wantErr:         true,
			wantErrContains: "sub-Dym-Name is not well-formed",
		},
		{
			name:            "invalid multi-sub-Dym-Name, @",
			dymNameAddress:  "c-.b.a@dym",
			wantErr:         true,
			wantErrContains: "sub-Dym-Name is not well-formed",
		},
		{
			name:            "invalid multi-sub-Dym-Name",
			dymNameAddress:  "c-.b.a.dym",
			wantErr:         true,
			wantErrContains: "sub-Dym-Name is not well-formed",
		},
		{
			name:            "blank path",
			dymNameAddress:  "b. .a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "do not accept continuous dot",
			dymNameAddress:  "b..a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "do not accept continuous '@'",
			dymNameAddress:  "a@@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "do not accept continuous '@'",
			dymNameAddress:  "b.a@@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "do not accept multiple '@'",
			dymNameAddress:  "b@a@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "do not accept multiple '@'",
			dymNameAddress:  "@a@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "do not accept multiple '@'",
			dymNameAddress:  "@a.b@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "do not accept multiple '@'",
			dymNameAddress:  "a@b@dym",
			wantErr:         true,
			wantErrContains: "multiple '@' found",
		},
		{
			name:            "bad name",
			dymNameAddress:  "a.@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "bad name",
			dymNameAddress:  "a.b.@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "bad name",
			dymNameAddress:  "a.b@.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "bad name",
			dymNameAddress:  "a.b.@.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "bad name",
			dymNameAddress:  ".b.a.dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "bad name",
			dymNameAddress:  ".b.a@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "empty",
			dymNameAddress:  "",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:               "allow hex address pattern",
			dymNameAddress:     "0x1234567890123456789012345678901234567890@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "0x1234567890123456789012345678901234567890",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "allow 32 bytes hex address pattern",
			dymNameAddress:     "0x1234567890123456789012345678901234567890123456789012345678901234@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "0x1234567890123456789012345678901234567890123456789012345678901234",
			wantChainIdOrAlias: "dym",
		},
		{
			name:            "reject non-20 or 32 bytes hex address pattern, case 19 bytes",
			dymNameAddress:  "0x123456789012345678901234567890123456789@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "reject non-20 or 32 bytes hex address pattern, case 21 bytes",
			dymNameAddress:  "0x12345678901234567890123456789012345678901@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "reject non-20 or 32 bytes hex address pattern, case 31 bytes",
			dymNameAddress:  "0x123456789012345678901234567890123456789012345678901234567890123@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "reject non-20 or 32 bytes hex address pattern, case 33 bytes",
			dymNameAddress:  "0x12345678901234567890123456789012345678901234567890123456789012345@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:               "allow valid bech32 address pattern",
			dymNameAddress:     "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96",
			wantChainIdOrAlias: "dym",
		},
		{
			name:               "allow valid bech32 address pattern, Interchain Account",
			dymNameAddress:     "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul@dym",
			wantErr:            false,
			wantSubName:        "",
			wantDymName:        "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul",
			wantChainIdOrAlias: "dym",
		},
		{
			name:            "reject invalid bech32 address pattern",
			dymNameAddress:  "dym1zzzzzzzzzz69v7yszg69v7yszg69v7ys8xdv96@dym",
			wantErr:         true,
			wantErrContains: dymnstypes.ErrBadDymNameAddress.Error(),
		},
		{
			name:            "reject invalid bech32 address pattern, Interchain Account",
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

	const owner = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	ownerHex := common.BytesToAddress(sdk.MustAccAddressFromBech32(owner)).Hex()
	ownerWithRollApp1Bech32, err := bech32.ConvertAndEncode(rollApp1Bech32, sdk.MustAccAddressFromBech32(owner))
	require.NoError(t, err)
	const anotherAcc = "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d"
	anotherAccHex := common.BytesToAddress(sdk.MustAccAddressFromBech32(anotherAcc)).Hex()
	const ica = "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul"
	icaHex := common.BytesToHash(sdk.MustAccAddressFromBech32(ica)).Hex()
	const rollAppAccBech32 = "nim1az0ze53xf8c3safe4x38yy8e6rnmlpxa225vf0"
	_, bzRollAppAcc, err := bech32.DecodeAndConvert(rollAppAccBech32)
	require.NoError(t, err)
	rollAppAccHex := common.BytesToAddress(bzRollAppAcc).Hex()

	require.False(t, strings.EqualFold(ownerHex, rollAppAccHex))

	setupTest := func() (dymnskeeper.Keeper, rollappkeeper.Keeper, sdk.Context) {
		dk, _, rk, ctx := testkeeper.DymNSKeeper(t)
		ctx = ctx.WithBlockTime(now).WithChainID(chainId)

		rk.SetRollapp(ctx, rollapptypes.Rollapp{
			RollappId: rollAppId1,
			Creator:   owner,
		})

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
			name: "pass - can resolve",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    owner,
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
			name: "pass - can resolve ICA",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("", "ica", ica).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ica,
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
			name: "pass - only take records matching input chain-id",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    owner,
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
			dymNames: newDN("a", owner).
				exp(now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    anotherAcc,
			workingChainId:  chainId,
			wantErr:         false,
			want:            nil,
		},
		{
			name: "pass - lookup by hex on host chain",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerHex,
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
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("", "ica", ica).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    icaHex,
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
			name: "pass - lookup by hex on coin-type-60 chain-id without bech32 prefix mapped, should do fallback lookup",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				moduleParams := dk.GetParams(ctx)
				moduleParams.Chains.CoinType60ChainIds = []string{"blumbus_111-1"}
				require.NoError(t, dk.SetParams(ctx, moduleParams))
			},
			inputAddress:   ownerHex,
			workingChainId: "blumbus_111-1",
			wantErr:        false,
			want: dymnstypes.ReverseResolvedDymNameAddresses{
				{
					SubName:        "", // fallback lookup from default name config so does not have Path => SubName
					Name:           "a",
					ChainIdOrAlias: "blumbus_111-1",
				},
			},
		},
		{
			// TODO DymNS FIXME: this test-case is failed because current x/rollapp implementation
			//  does not support get bech32 of RollApp
			name: "FIXME * pass - lookup by hex on RollApp (coin-type-60 chain-id) with bech32 prefix mapped, find out the matching configuration",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				cfgN(rollAppId1, "ra", rollAppAccBech32).
				buildSlice(),
			inputAddress:   rollAppAccHex,
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
			name: "pass - lookup by hex on RollApp (coin-type-60 chain-id) with bech32 prefix mapped, but matching configuration of corresponding address so we do fallback lookup",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				buildSlice(),
			inputAddress:   ownerHex,
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
			// TODO DymNS FIXME: this test-case is failed because current x/rollapp implementation
			//  does not support get bech32 of RollApp
			name: "FIXME * pass - lookup by hex on RollApp (coin-type-60 chain-id) with bech32 prefix mapped, find out the matching configuration",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
				cfgN(rollAppId1, "ra", ownerWithRollApp1Bech32).
				buildSlice(),
			inputAddress:   ownerHex,
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
			name: "pass - skip lookup by hex after first try (direct match) if working-chain-id is NOT recorgnized as a coin-type-60 chain, by bech32",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "", owner).
				buildSlice(),
			inputAddress:   anotherAcc,
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - skip lookup by hex if working-chain-id is NOT recorgnized as a coin-type-60 chain, by hex",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "", owner).
				buildSlice(),
			inputAddress:   ownerHex,
			workingChainId: "cosmoshub-4",
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - find result from multiple Dym-Names matched, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", owner).
					exp(now, +1).
					cfgN("", "b", owner).
					build(),
				newDN("b", owner).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    owner,
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
				newDN("a", owner).
					exp(now, +1).
					cfgN("", "b", owner).
					build(),
				newDN("b", owner).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerHex,
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
				newDN("a", owner).
					exp(now, +1).
					cfgN("", "b", owner).
					build(),
				newDN("b", owner).
					exp(now, +1).
					cfgN("", "b", owner).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    owner,
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
				newDN("a", owner).
					exp(now, -1).
					cfgN("", "b", owner).
					build(),
				newDN("b", owner).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    owner,
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
				newDN("a", owner).
					exp(now, -1).
					cfgN("", "b", owner).
					build(),
				newDN("b", owner).
					exp(now, +1).
					build(),
			},
			additionalSetup: nil,
			inputAddress:    ownerHex,
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
			dymNames:        newDN("a", owner).buildSlice(),
			inputAddress:    "",
			workingChainId:  chainId,
			wantErr:         true,
			wantErrContains: "not supported address format",
		},
		{
			name:            "fail - reject bad input address",
			dymNames:        newDN("a", owner).buildSlice(),
			inputAddress:    "0xdym1",
			workingChainId:  chainId,
			wantErr:         true,
			wantErrContains: "not supported address format",
		},
		{
			name:            "fail - reject empty working-chain-id",
			dymNames:        newDN("a", owner).buildSlice(),
			inputAddress:    owner,
			workingChainId:  "",
			wantErr:         true,
			wantErrContains: "invalid chain-id format",
		},
		{
			name:            "fail - reject bad working-chain-id",
			dymNames:        newDN("a", owner).buildSlice(),
			inputAddress:    owner,
			workingChainId:  "@",
			wantErr:         true,
			wantErrContains: "invalid chain-id format",
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by bech32",
			dymNames: []dymnstypes.DymName{
				newDN("a", owner).
					exp(now, +1).
					build(),
				newDN("b", owner).
					exp(now, +1).
					build(),
				newDN("c", anotherAcc).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, owner, "c")
				require.NoError(t, err)
				err = dk.AddReverseMappingHexAddressToDymName(ctx, common.HexToAddress(ownerHex).Bytes(), "c")
				require.NoError(t, err)
			},
			inputAddress:   owner,
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
				newDN("a", owner).
					exp(now, +1).
					build(),
				newDN("b", owner).
					exp(now, +1).
					build(),
				newDN("c", anotherAcc).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, owner, "c")
				require.NoError(t, err)
				err = dk.AddReverseMappingHexAddressToDymName(ctx, common.HexToAddress(ownerHex).Bytes(), "c")
				require.NoError(t, err)
			},
			inputAddress:   ownerHex,
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
				newDN("a", owner).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherAcc, "a")
				require.NoError(t, err)
				err = dk.AddReverseMappingHexAddressToDymName(ctx, common.HexToAddress(anotherAccHex).Bytes(), "a")
				require.NoError(t, err)
			},
			inputAddress:   anotherAcc,
			workingChainId: chainId,
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - should not include the Dym-Name that mistakenly linked to Dym-Name that does not correct config relates to the account, by hex",
			dymNames: []dymnstypes.DymName{
				newDN("a", owner).
					exp(now, +1).
					build(),
			},
			additionalSetup: func(ctx sdk.Context, dk dymnskeeper.Keeper) {
				err := dk.AddReverseMappingConfiguredAddressToDymName(ctx, anotherAcc, "a")
				require.NoError(t, err)
				err = dk.AddReverseMappingHexAddressToDymName(ctx, common.HexToAddress(anotherAccHex).Bytes(), "a")
				require.NoError(t, err)
			},
			inputAddress:   anotherAccHex,
			workingChainId: chainId,
			wantErr:        false,
			want:           nil,
		},
		{
			name: "pass - matching by hex if bech32 is not found, on host chain",
			dymNames: newDN("a", owner).
				exp(now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerWithRollApp1Bech32,
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
			name: "pass - matching by hex if bech32 is not found, on coin-type-60 chain",
			dymNames: newDN("a", owner).
				exp(now, +1).
				buildSlice(),
			additionalSetup: nil,
			inputAddress:    ownerWithRollApp1Bech32,
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
			name: "pass - alias is used if available, by bech32",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
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
			inputAddress:   owner,
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
			name: "pass - alias is used if available, by hex",
			dymNames: newDN("a", owner).
				exp(now, +1).
				cfgN("", "b", owner).
				cfgN("blumbus_111-1", "bb", owner).
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
			inputAddress:   ownerHex,
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

//goland:noinspection SpellCheckingInspection
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
			ChainId: "rollapp_2-2",
			Aliases: nil,
		},
		{
			ChainId: "rollapp_3-3",
			Aliases: nil,
		},
	}
	require.NoError(t, dk.SetParams(ctx, moduleParams))

	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_1-1",
		Creator:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_1-1"))
	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_2-2",
		Creator:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_2-2"))
	rk.SetRollapp(ctx, rollapptypes.Rollapp{
		RollappId: "rollapp_3-3",
		Creator:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
	})
	require.True(t, dk.IsRollAppId(ctx, "rollapp_3-3"))

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

	t.Run("FIXME * mapping correct alias for RollApp by ID", func(t *testing.T) {
		input := []dymnstypes.ReverseResolvedDymNameAddress{
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
		}
		require.Equal(t,
			[]dymnstypes.ReverseResolvedDymNameAddress{
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
			},
			dk.ReplaceChainIdWithAliasIfPossible(ctx, input),
		)
	})

	t.Run("allow passing empty", func(t *testing.T) {
		require.Empty(t, dk.ReplaceChainIdWithAliasIfPossible(ctx, nil))
		require.Empty(t, dk.ReplaceChainIdWithAliasIfPossible(ctx, []dymnstypes.ReverseResolvedDymNameAddress{}))
	})
}
