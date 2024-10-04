package keeper

import (
	"fmt"
	"slices"
	"strconv"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GetRollAppIdByAlias returns the RollApp ID which is linked to the input alias.
func (k Keeper) GetRollAppIdByAlias(ctx sdk.Context, alias string) (rollAppId string, found bool) {
	defer func() {
		found = rollAppId != ""
	}()

	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.AliasToRollAppEip155IdRvlKey(alias)
	bz := store.Get(key)
	if bz != nil {
		rollAppEip155Id := string(bz)
		eip155, _ := strconv.ParseUint(rollAppEip155Id, 10, 64)
		var foundRollApp bool
		rollAppId, foundRollApp = k.rollappKeeper.GetRollAppIdByEIP155(ctx, eip155)
		if !foundRollApp {
			// this should not happen as validated before
			panic(fmt.Sprintf("rollapp not found by EIP155 of alias '%s': %s", alias, rollAppEip155Id))
		}
	}

	return
}

// GetAliasByRollAppId returns the first alias (in case RollApp has multiple aliases) linked to the RollApp ID.
func (k Keeper) GetAliasByRollAppId(ctx sdk.Context, rollAppId string) (alias string, found bool) {
	if !k.IsRollAppId(ctx, rollAppId) {
		return
	}

	defer func() {
		found = alias != ""
	}()

	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.RollAppIdToAliasesKey(rollAppId)
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
	keyRE155toA := dymnstypes.RollAppIdToAliasesKey(rollAppId)
	keyAtoRE155 := dymnstypes.AliasToRollAppEip155IdRvlKey(alias)

	// ensure the alias is not being used by another RollApp
	if bz := store.Get(keyAtoRE155); bz != nil {
		eip155, _ := strconv.ParseUint(string(bz), 10, 64)
		usedByRollAppId, _ := k.rollappKeeper.GetRollAppIdByEIP155(ctx, eip155)
		return errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias currently being in used by: %s", usedByRollAppId)
	}

	// one RollApp can have multiple aliases, append to the existing list
	var multipleAliases dymnstypes.MultipleAliases
	if bz := store.Get(keyRE155toA); bz != nil {
		k.cdc.MustUnmarshal(bz, &multipleAliases)
	}
	multipleAliases.Aliases = append(multipleAliases.Aliases, alias)

	store.Set(keyRE155toA, k.cdc.MustMarshal(&multipleAliases))
	store.Set(keyAtoRE155, []byte(dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollAppId)))

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
	var effectiveAliases []string

	// check if there is a mapping in params
	for _, aliasesOfChainId := range k.ChainsParams(ctx).AliasesOfChainIds {
		if aliasesOfChainId.ChainId != chainId {
			continue
		}
		effectiveAliases = aliasesOfChainId.Aliases
		break
	}

	if k.IsRollAppId(ctx, chainId) {
		aliasesOfRollApp := k.GetAliasesOfRollAppId(ctx, chainId)

		// If the chain-id is a RollApp, must exclude the aliases which being reserved in params.
		// Please read the `processCompleteSellOrderWithAssetTypeAlias` method (msg_server_complete_sell_order.go) for more information.
		reservedAliases := k.GetAllAliasAndChainIdInParams(ctx)
		aliasesOfRollApp = slices.DeleteFunc(aliasesOfRollApp, func(a string) bool {
			_, found := reservedAliases[a]
			return found
		})

		effectiveAliases = append(effectiveAliases, aliasesOfRollApp...)
	}

	return effectiveAliases
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
	keyRE155toA := dymnstypes.RollAppIdToAliasesKey(rollAppId)
	keyAtoRE155 := dymnstypes.AliasToRollAppEip155IdRvlKey(alias)

	// ensure the alias is being used by the RollApp
	bzRollAppEip155Id := store.Get(keyAtoRE155)
	if bzRollAppEip155Id == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found: %s", alias)
	}

	if string(bzRollAppEip155Id) != dymnsutils.MustGetEIP155ChainIdFromRollAppId(rollAppId) {
		eip155, _ := strconv.ParseUint(string(bzRollAppEip155Id), 10, 64)
		usedByRollAppId, _ := k.rollappKeeper.GetRollAppIdByEIP155(ctx, eip155)
		return errorsmod.Wrapf(gerrc.ErrPermissionDenied, "alias currently being in used by: %s", usedByRollAppId)
	}

	// load the existing aliases of the RollApp
	var multipleAliases dymnstypes.MultipleAliases
	if bz := store.Get(keyRE155toA); bz != nil {
		k.cdc.MustUnmarshal(bz, &multipleAliases)
	}

	// remove the alias from the RollApp alias list
	originalAliasesCount := len(multipleAliases.Aliases)
	multipleAliases.Aliases = slices.DeleteFunc(multipleAliases.Aliases, func(a string) bool {
		return a == alias
	})
	if len(multipleAliases.Aliases) == originalAliasesCount {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found: %s", alias)
	}

	// update store

	// if no alias left, remove the key, otherwise update new list
	if len(multipleAliases.Aliases) == 0 {
		store.Delete(keyRE155toA)
	} else {
		store.Set(keyRE155toA, k.cdc.MustMarshal(&multipleAliases))
	}

	// remove the alias to RollAppId mapping
	store.Delete(keyAtoRE155)

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

	// remove the existing link

	if err := k.RemoveAliasFromRollAppId(ctx, srcRollAppId, alias); err != nil {
		return err
	}

	// set the new link

	return k.SetAliasForRollAppId(ctx, dstRollAppId, alias)
}

// GetAllAliasAndChainIdInParams returns all aliases and chain-ids in the params.
// Note: this method returns a map so the iteration is non-deterministic,
// any implementation should not rely on the order of the result.
func (k Keeper) GetAllAliasAndChainIdInParams(ctx sdk.Context) map[string]struct{} {
	result := make(map[string]struct{})
	for _, aliasesOfChainId := range k.ChainsParams(ctx).AliasesOfChainIds {
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

// SetDefaultAliasForRollApp move the alias into the first place, so it can be used as default alias in resolution.
func (k Keeper) SetDefaultAliasForRollApp(ctx sdk.Context, rollAppId, alias string) error {
	if _, err := rollapptypes.NewChainID(rollAppId); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid RollApp chain-id: %s", rollAppId)
	}

	// load the existing aliases of the RollApp from store
	existingAliases := k.GetAliasesOfRollAppId(ctx, rollAppId)

	// swap the alias to the first place
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

	if existingIndex == 0 { // no need to do anything
		return nil
	}

	existingAliases[0], existingAliases[existingIndex] = existingAliases[existingIndex], existingAliases[0]

	// update the new list into store

	store := ctx.KVStore(k.storeKey)
	keyRE155toA := dymnstypes.RollAppIdToAliasesKey(rollAppId)
	store.Set(keyRE155toA, k.cdc.MustMarshal(&dymnstypes.MultipleAliases{Aliases: existingAliases}))

	return nil
}

// GetAllRollAppsWithAliases returns all RollApp IDs which have aliases and their aliases.
func (k Keeper) GetAllRollAppsWithAliases(ctx sdk.Context) (list []dymnstypes.AliasesOfChainId) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, dymnstypes.KeyPrefixRollAppEip155IdToAliases)
	defer func() {
		_ = iterator.Close()
	}()

	eip155ToRollAppIdCache := make(map[string]string)
	// Describe usage of Go Map: used for caching purpose, no iteration.

	for ; iterator.Valid(); iterator.Next() {
		var multipleAliases dymnstypes.MultipleAliases
		k.cdc.MustUnmarshal(iterator.Value(), &multipleAliases)

		var rollAppId string
		eip155Id := string(iterator.Key()[len(dymnstypes.KeyPrefixRollAppEip155IdToAliases):])
		if cachedRollAppId, found := eip155ToRollAppIdCache[eip155Id]; found {
			rollAppId = cachedRollAppId
		} else {
			eip155, _ := strconv.ParseUint(eip155Id, 10, 64)
			var foundRollApp bool
			rollAppId, foundRollApp = k.rollappKeeper.GetRollAppIdByEIP155(ctx, eip155)
			if !foundRollApp {
				// this should not happen as validated before
				panic(fmt.Sprintf("rollapp not found by EIP155: %s", eip155Id))
			}
			eip155ToRollAppIdCache[eip155Id] = rollAppId // cache the result
		}

		list = append(list, dymnstypes.AliasesOfChainId{
			ChainId: rollAppId,
			Aliases: multipleAliases.Aliases,
		})
	}

	return list
}
