package keeper

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// TODO DymNS: delete this, only apply for playground

func (k Keeper) BeginBlockMigrationForPlayground(ctx sdk.Context) {
	var migrate bool
	if ctx.ChainID() == "dymension_2018-1" { // Playground
		migrate = ctx.BlockHeight()%100 == 0
	} else if ctx.ChainID() == "dymension_100-1" { // localnet
		migrate = true
	}

	if migrate {
		k.MigrateStoreForPlayground(ctx)
	}
}

func (k Keeper) MigrateStoreForPlayground(ctx sdk.Context) (anyMigrated bool) {
	m1 := k.migrateDymNameForPlayground(ctx)
	m2 := k.migrateLinkingRollAppToAliasForPlayground(ctx)

	anyMigrated = m1 || m2
	if anyMigrated {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"dymns_migrated_for_playground",
				sdk.NewAttribute("migrated dym-name", fmt.Sprintf("%t", m1)),
				sdk.NewAttribute("migrated roll-app <=> alias linking", fmt.Sprintf("%t", m2)),
			),
		)
	}
	return
}

func (k Keeper) migrateDymNameForPlayground(ctx sdk.Context) (anyMigrated bool) {
	dymNames := k.GetAllDymNames(ctx)
	for _, dymName := range dymNames {
		var anyUpdated bool
		for i, config := range dymName.Configs {
			if config.ChainId == "" {
				continue
			}

			if dymnsutils.IsValidEIP155ChainId(config.ChainId) {
				// migrated
				continue
			}

			rollAppId, err := rollapptypes.NewChainID(config.ChainId)
			if err != nil {
				// not RollApp
				continue
			}

			if !k.IsRollAppId(ctx, config.ChainId) {
				// not target
				continue
			}

			config.ChainId = fmt.Sprintf("%d", rollAppId.GetEIP155ID())
			dymName.Configs[i] = config

			anyUpdated = true
		}

		if anyUpdated {
			err := k.SetDymName(ctx, dymName)
			if err != nil {
				panic(err)
			}
			anyMigrated = true

			// no need to call hooks
		}
	}

	return
}

func (k Keeper) migrateLinkingRollAppToAliasForPlayground(ctx sdk.Context) (anyMigrated bool) {
	store := ctx.KVStore(k.storeKey)

	type newLinking struct {
		rollAppId       string
		multipleAliases dymnstypes.MultipleAliases
	}

	newLinkingRecords, keysToDelete := func() ([]newLinking, [][]byte) {
		// do not update on the fly
		newLinkingRecords := make([]newLinking, 0)
		keysToDelete := make([][]byte, 0)

		iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixRollAppEip155IdToAliases)
		defer func() {
			_ = iterator.Close()
		}()

		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			rollAppId := string(key[len(dymnstypes.KeyPrefixRollAppEip155IdToAliases):])
			if dymnsutils.IsValidEIP155ChainId(rollAppId) {
				// migrated
				continue
			}

			keysToDelete = append(keysToDelete, key)

			var multipleAliases dymnstypes.MultipleAliases
			k.cdc.MustUnmarshal(iterator.Value(), &multipleAliases)

			if len(multipleAliases.Aliases) == 0 {
				// no need to migrate
				continue
			}

			newLinkingRecords = append(newLinkingRecords, newLinking{
				rollAppId:       rollAppId,
				multipleAliases: multipleAliases,
			})
		}

		return newLinkingRecords, keysToDelete
	}()

	anyMigrated = len(newLinkingRecords) > 0 || len(keysToDelete) > 0

	for _, key := range keysToDelete {
		store.Delete(key)
	}

	for _, linking := range newLinkingRecords {
		key := dymnstypes.RollAppIdToAliasesKey(linking.rollAppId)

		bzExistingData := store.Get(key)
		if len(bzExistingData) > 0 { // merge existing
			var existingMultipleAliases dymnstypes.MultipleAliases
			k.cdc.MustUnmarshal(bzExistingData, &existingMultipleAliases)

			if len(existingMultipleAliases.Aliases) > 0 {
				unique := make(map[string]struct{})
				for _, alias := range existingMultipleAliases.Aliases {
					unique[alias] = struct{}{}
				}
				for _, alias := range linking.multipleAliases.Aliases {
					unique[alias] = struct{}{}
				}

				uniqueAliases := make([]string, 0, len(unique))
				for alias := range unique {
					uniqueAliases = append(uniqueAliases, alias)
				}

				// important: sort to guarantee of consensus state
				slices.Sort(uniqueAliases)

				linking.multipleAliases.Aliases = uniqueAliases
			}
		}

		store.Set(key, k.cdc.MustMarshal(&linking.multipleAliases))

		bzEip155 := []byte(dymnsutils.MustGetEIP155ChainIdFromRollAppId(linking.rollAppId))
		for _, alias := range linking.multipleAliases.Aliases {
			keyAtoRE155 := dymnstypes.AliasToRollAppEip155IdRvlKey(alias)
			store.Set(keyAtoRE155, bzEip155)
		}
	}

	return
}
