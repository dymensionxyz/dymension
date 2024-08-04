package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// MigrateChainIds called by GOV handler, migrate chain-ids in module params
// as well as Dym-Names configurations (of non-expired) those contain chain-ids.
func (k Keeper) MigrateChainIds(ctx sdk.Context, replacement []dymnstypes.MigrateChainId) error {
	previousChainIdsToNewChainId := make(map[string]string)
	// Describe usage of Go Map: only used for mapping previous chain id to new chain id
	// and should not be used for other purposes as well as iteration.

	for _, r := range replacement {
		previousChainIdsToNewChainId[r.PreviousChainId] = r.NewChainId
	}

	if err := k.migrateChainIdsInParams(ctx, previousChainIdsToNewChainId); err != nil {
		return err
	}

	if err := k.migrateChainIdsInDymNames(ctx, previousChainIdsToNewChainId); err != nil {
		return err
	}

	return nil
}

// migrateChainIdsInParams migrates chain-ids in module params.
func (k Keeper) migrateChainIdsInParams(ctx sdk.Context, previousChainIdsToNewChainId map[string]string) error {
	params := k.GetParams(ctx)

	if len(params.Chains.AliasesOfChainIds) > 0 {
		existingAliasesOfChainIds := make(map[string]dymnstypes.AliasesOfChainId)
		// Describe usage of Go Map: only used for mapping chain id to the alias configuration  of chain id
		// and should not be used for other purposes as well as iteration.

		for _, record := range params.Chains.AliasesOfChainIds {
			existingAliasesOfChainIds[record.ChainId] = record
		}

		newAliasesByChainId := make([]dymnstypes.AliasesOfChainId, 0)
		for _, record := range params.Chains.AliasesOfChainIds {
			chainId := record.ChainId
			aliases := record.Aliases

			if newChainId, isPreviousChainId := previousChainIdsToNewChainId[chainId]; isPreviousChainId {
				if _, foundDeclared := existingAliasesOfChainIds[newChainId]; foundDeclared {
					// we don't override, we keep the aliases of the new chain id

					// ignore and remove the aliases of the previous chain id
				} else {
					newAliasesByChainId = append(newAliasesByChainId, dymnstypes.AliasesOfChainId{
						ChainId: newChainId,
						Aliases: aliases,
					})
				}
			} else {
				newAliasesByChainId = append(newAliasesByChainId, dymnstypes.AliasesOfChainId{
					ChainId: chainId,
					Aliases: aliases,
				})
			}
		}
		params.Chains.AliasesOfChainIds = newAliasesByChainId
	}

	if err := k.SetParams(ctx, params); err != nil {
		k.Logger(ctx).Error(
			"failed to update params",
			"error", err,
			"migration-state", "aborted",
		)
		return err
	}

	return nil
}

// migrateChainIdsInDymNames migrates chain-ids in non-expired Dym-Names configurations.
func (k Keeper) migrateChainIdsInDymNames(ctx sdk.Context, previousChainIdsToNewChainId map[string]string) error {
	// We only migrate for Dym-Names that not expired to reduce IO needed.

	nonExpiredDymNames := k.GetAllNonExpiredDymNames(ctx)
	if len(nonExpiredDymNames) < 1 {
		return nil
	}

	for _, dymName := range nonExpiredDymNames {
		newConfigs := make([]dymnstypes.DymNameConfig, len(dymName.Configs))
		var anyConfigUpdated bool
		for i, config := range dymName.Configs {
			if config.ChainId != "" {
				if newChainId, isPreviousChainId := previousChainIdsToNewChainId[config.ChainId]; isPreviousChainId {
					config.ChainId = newChainId
					anyConfigUpdated = true
				}
			}

			newConfigs[i] = config
		}

		if !anyConfigUpdated {
			// Skip migration for this Dym-Name if nothing updated to reduce IO.
			continue
		}

		dymName.Configs = newConfigs

		if err := dymName.Validate(); err != nil {
			k.Logger(ctx).Error(
				"failed to migrate chain ids for Dym-Name",
				"dymName", dymName.Name,
				"step", "Validate",
				"error", err,
				"migration-state", "continue",
			)
			// Skip migration for this Dym-Name.
			// We don't want to break the migration process for other Dym-Names.
			// The replacement should be done later by the owner.
			continue
		}

		// From here, any step can procedures dirty state, so we need to abort the migration

		// We do not call BeforeDymNameConfigChanged and AfterDymNameConfigChanged
		// here because we only change the chain id, which does not affect any data
		// that need to be updated in those methods, so we can skip them to reduce IO.
		// Reverse-resolve records are re-computed in runtime anyway.

		if err := k.SetDymName(ctx, dymName); err != nil {
			k.Logger(ctx).Error(
				"failed to migrate chain ids for Dym-Name",
				"dymName", dymName.Name,
				"step", "SetDymName",
				"error", err,
				"migration-state", "aborted",
			)
			return err
		}

		k.Logger(ctx).Info("migrated chain ids for Dym-Name", "dymName", dymName.Name)
	}

	return nil
}

// UpdateAliases called by GOV handler, update aliases of chain-ids in module params.
func (k Keeper) UpdateAliases(ctx sdk.Context, add, remove []dymnstypes.UpdateAlias) error {
	params := k.GetParams(ctx)

	chainIdToAliasConfig := make(map[string]map[string]bool)
	// Describe usage of Go Map: used to map from chain id to alias configuration.
	// This map is used to quickly find the alias configuration of a chain id.
	// Data should be sorted before persist.

	for _, record := range params.Chains.AliasesOfChainIds {
		aliasesPerChainId := make(map[string]bool)
		for _, alias := range record.Aliases {
			aliasesPerChainId[alias] = true
		}
		chainIdToAliasConfig[record.ChainId] = aliasesPerChainId
	}

	if len(add) > 0 {
		for _, record := range add {
			chainId := record.ChainId
			alias := record.Alias

			existingAliases, foundExistingChainId := chainIdToAliasConfig[chainId]
			if !foundExistingChainId {
				existingAliases = make(map[string]bool)
			}

			_, foundAlias := existingAliases[alias]
			if foundAlias {
				err := errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias: %s for %s", alias, chainId)
				k.Logger(ctx).Error(
					"failed to add alias for chain-id",
					"chain-id", chainId,
					"alias", alias,
					"step", "add",
					"error", err,
					"update-state", "aborted",
				)
				return err
			}

			existingAliases[alias] = true
			chainIdToAliasConfig[chainId] = existingAliases
		}
	}

	if len(remove) > 0 {
		for _, record := range remove {
			chainId := record.ChainId
			alias := record.Alias

			aliasesPerChainId, foundExistingChainId := chainIdToAliasConfig[chainId]
			if !foundExistingChainId {
				err := errorsmod.Wrapf(gerrc.ErrNotFound, "chain id not found to remove: %s", chainId)
				k.Logger(ctx).Error(
					"failed to remove alias for chain-id",
					"chain-id", chainId,
					"alias", alias,
					"step", "remove",
					"error", err,
					"update-state", "aborted",
				)
				return err
			}

			_, foundAlias := aliasesPerChainId[alias]
			if !foundAlias {
				err := errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found to remove: %s", alias)
				k.Logger(ctx).Error(
					"failed to remove alias for chain-id",
					"chain-id", chainId,
					"alias", alias,
					"step", "remove",
					"error", err,
					"update-state", "aborted",
				)
				return err
			}

			delete(aliasesPerChainId, alias)
			if len(aliasesPerChainId) == 0 {
				delete(chainIdToAliasConfig, chainId)
			}
		}
	}

	// build new params
	// Note: data must be sorted before persist

	sortedChainIds := dymnsutils.GetSortedStringKeys(chainIdToAliasConfig)

	var newAliasesOfChainIds []dymnstypes.AliasesOfChainId
	for _, chainId := range sortedChainIds {
		newAliasesOfChainIds = append(newAliasesOfChainIds, dymnstypes.AliasesOfChainId{
			ChainId: chainId,
			Aliases: dymnsutils.GetSortedStringKeys(chainIdToAliasConfig[chainId]),
		})
	}
	params.Chains.AliasesOfChainIds = newAliasesOfChainIds

	if err := k.SetParams(ctx, params); err != nil {
		k.Logger(ctx).Error(
			"failed to update params",
			"error", err,
			"migration-state", "aborted",
		)
		return err
	}

	return nil
}
