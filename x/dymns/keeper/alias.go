package keeper

import (
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GetRollAppIdByAlias returns the RollApp-Id by the alias.
func (k Keeper) GetRollAppIdByAlias(ctx sdk.Context, alias string) (rollAppId string, found bool) {
	defer func() {
		found = rollAppId != ""
	}()

	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.AliasToRollAppIdRvlKey(alias)
	bz := store.Get(key)
	if bz != nil {
		rollAppId = string(bz)
	}

	return
}

// GetAliasByRollAppId returns the first alias (in case RollApp has multiple aliases) by the RollApp-Id.
func (k Keeper) GetAliasByRollAppId(ctx sdk.Context, chainId string) (alias string, found bool) {
	if !k.IsRollAppId(ctx, chainId) {
		return
	}

	defer func() {
		found = alias != ""
	}()

	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.RollAppIdToAliasesKey(chainId)
	bz := store.Get(key)
	if bz != nil {
		var multipleAliases dymnstypes.MultipleAliases
		k.cdc.MustUnmarshal(bz, &multipleAliases)
		alias = multipleAliases.Aliases[0]
	}

	return
}

// SetAliasForRollAppId assigns the usage of an alias to a RollApp.
func (k Keeper) SetAliasForRollAppId(ctx sdk.Context, rollAppId, alias string) error {
	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias: %s", alias)
	}

	if !k.IsRollAppId(ctx, rollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "not a RollApp chain-id: %s", rollAppId)
	}

	store := ctx.KVStore(k.storeKey)
	keyR2A := dymnstypes.RollAppIdToAliasesKey(rollAppId)
	keyA2R := dymnstypes.AliasToRollAppIdRvlKey(alias)

	if bz := store.Get(keyA2R); bz != nil {
		return errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias currently being in used by: %s", string(bz))
	}

	var multipleAliases dymnstypes.MultipleAliases
	if bz := store.Get(keyR2A); bz != nil {
		k.cdc.MustUnmarshal(bz, &multipleAliases)
	}
	multipleAliases.Aliases = append(multipleAliases.Aliases, alias)

	store.Set(keyR2A, k.cdc.MustMarshal(&multipleAliases))
	store.Set(keyA2R, []byte(rollAppId))

	return nil
}

// GetAliasesOfRollAppId returns all aliases linked to a RollApp.
// Notes: the result does not exclude aliases reserved in params.
func (k Keeper) GetAliasesOfRollAppId(ctx sdk.Context, rollAppId string) []string {
	store := ctx.KVStore(k.storeKey)
	keyR2A := dymnstypes.RollAppIdToAliasesKey(rollAppId)

	var multipleAliases dymnstypes.MultipleAliases
	if bz := store.Get(keyR2A); bz != nil {
		k.cdc.MustUnmarshal(bz, &multipleAliases)
	}

	return multipleAliases.Aliases
}

// GetEffectiveAliasesByChainId returns all effective aliases by chain-id.
// Effective means: if an alias is reserved in params, it will be excluded from the result if the chain-id is a RollApp.
func (k Keeper) GetEffectiveAliasesByChainId(ctx sdk.Context, chainId string) []string {
	var result []string
	for _, aliasesOfChainId := range k.GetParams(ctx).Chains.AliasesOfChainIds {
		if aliasesOfChainId.ChainId != chainId {
			continue
		}
		result = aliasesOfChainId.Aliases
		break
	}

	if k.IsRollAppId(ctx, chainId) {
		reservedAliases := k.GetAllAliasAndChainIdInParams(ctx)

		aliases := k.GetAliasesOfRollAppId(ctx, chainId)

		aliases = slices.DeleteFunc(aliases, func(a string) bool {
			_, found := reservedAliases[a]
			return found
		})

		result = append(result, aliases...)
	}

	return result
}

// RemoveAliasFromRollAppId removes the linking of an existing alias from a RollApp.
func (k Keeper) RemoveAliasFromRollAppId(ctx sdk.Context, rollAppId, alias string) error {
	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias: %s", alias)
	}

	if !k.IsRollAppId(ctx, rollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "not a RollApp chain-id: %s", rollAppId)
	}

	store := ctx.KVStore(k.storeKey)
	keyR2A := dymnstypes.RollAppIdToAliasesKey(rollAppId)
	keyA2R := dymnstypes.AliasToRollAppIdRvlKey(alias)

	bzRollAppId := store.Get(keyA2R)
	if bzRollAppId == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found: %s", alias)
	} else if string(bzRollAppId) != rollAppId {
		return errorsmod.Wrapf(gerrc.ErrPermissionDenied, "alias currently being in used by: %s", string(bzRollAppId))
	}

	var multipleAliases dymnstypes.MultipleAliases
	if bz := store.Get(keyR2A); bz != nil {
		k.cdc.MustUnmarshal(bz, &multipleAliases)
	}
	countAliases := len(multipleAliases.Aliases)

	multipleAliases.Aliases = slices.DeleteFunc(multipleAliases.Aliases, func(a string) bool {
		return a == alias
	})
	if len(multipleAliases.Aliases) == countAliases {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found: %s", alias)
	}

	if len(multipleAliases.Aliases) == 0 {
		store.Delete(keyR2A)
	} else {
		store.Set(keyR2A, k.cdc.MustMarshal(&multipleAliases))
	}
	store.Delete(keyA2R)

	return nil
}

// MoveAliasToRollAppId moves the linking of an existing alias from a RollApp to another RollApp.
func (k Keeper) MoveAliasToRollAppId(ctx sdk.Context, srcRollAppId, alias, dstRollAppId string) error {
	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias: %s", alias)
	}

	if !k.IsRollAppId(ctx, srcRollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "source RollApp does not exists: %s", srcRollAppId)
	}

	if !k.IsRollAppId(ctx, dstRollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "destination RollApp does not exists: %s", dstRollAppId)
	}

	inUsedByRollApp, found := k.GetRollAppIdByAlias(ctx, alias)
	if !found {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found: %s", alias)
	}

	if inUsedByRollApp != srcRollAppId {
		return errorsmod.Wrapf(gerrc.ErrPermissionDenied, "source RollApp mis-match: %s", inUsedByRollApp)
	}

	if err := k.RemoveAliasFromRollAppId(ctx, srcRollAppId, alias); err != nil {
		return err
	}

	return k.SetAliasForRollAppId(ctx, dstRollAppId, alias)
}

// GetAllAliasAndChainIdInParams returns all aliases and chain-ids in the params.
// Note: this method returns a map so the iteration is non-deterministic,
// any implementation should not rely on the order of the result.
func (k Keeper) GetAllAliasAndChainIdInParams(ctx sdk.Context) map[string]struct{} {
	params := k.GetParams(ctx)

	result := make(map[string]struct{})
	for _, aliasesOfChainId := range params.Chains.AliasesOfChainIds {
		result[aliasesOfChainId.ChainId] = struct{}{}
		for _, a := range aliasesOfChainId.Aliases {
			result[a] = struct{}{}
		}
	}

	return result
}

// IsAliasPresentsInParamsAsAliasOrChainId returns true if the alias presents in the params.
// Extra check if it collision with chain-ids there.
func (k Keeper) IsAliasPresentsInParamsAsAliasOrChainId(ctx sdk.Context, alias string) bool {
	_, found := k.GetAllAliasAndChainIdInParams(ctx)[alias]
	return found
}

// SetDefaultAlias move the alias into the first place, so it can be used as default alias in resolution.
func (k Keeper) SetDefaultAlias(ctx sdk.Context, rollAppId, alias string) error {
	existingAliases := k.GetAliasesOfRollAppId(ctx, rollAppId)
	existingIndex := -1

	for i, existingAlias := range existingAliases {
		if alias == existingAlias {
			existingIndex = i
			break
		}
	}

	if existingIndex < 0 {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias is not linked to the RollApp: %s", alias)
	}

	if existingIndex == 0 {
		// no need to do anything
		return nil
	}

	existingAliases[0], existingAliases[existingIndex] = existingAliases[existingIndex], existingAliases[0]

	store := ctx.KVStore(k.storeKey)
	keyR2A := dymnstypes.RollAppIdToAliasesKey(rollAppId)
	store.Set(keyR2A, k.cdc.MustMarshal(&dymnstypes.MultipleAliases{Aliases: existingAliases}))

	return nil
}

// GetAllRollAppsWithAliases returns all RollAppIDs which have aliases and their aliases.
func (k Keeper) GetAllRollAppsWithAliases(ctx sdk.Context) (list []dymnstypes.AliasesOfChainId) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixRollAppIdToAliases)
	defer func() {
		_ = iterator.Close()
	}()

	for ; iterator.Valid(); iterator.Next() {
		var multipleAliases dymnstypes.MultipleAliases
		k.cdc.MustUnmarshal(iterator.Value(), &multipleAliases)
		list = append(list, dymnstypes.AliasesOfChainId{
			ChainId: string(iterator.Key()[len(dymnstypes.KeyPrefixRollAppIdToAliases):]),
			Aliases: multipleAliases.Aliases,
		})
	}

	return list
}
