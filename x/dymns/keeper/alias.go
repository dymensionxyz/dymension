package keeper

import (
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
		return
	}

	for rid, data := range mockRollAppsData {
		if data.alias == alias {
			rollAppId = rid
			return
		}
	}

	return
}

// GetAliasByRollAppId returns the alias by the RollApp-Id.
func (k Keeper) GetAliasByRollAppId(ctx sdk.Context, chainId string) (alias string, found bool) {
	// TODO DymNS: support returns multiple aliases
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
		if len(multipleAliases.Aliases) > 0 {
			alias = multipleAliases.Aliases[0]
			return
		}
	}

	if data, ok := mockRollAppsData[chainId]; ok {
		alias = data.alias
		return
	}

	return
}

// SetAliasForRollAppId assigns the usage of an alias to a RollApp.
func (k Keeper) SetAliasForRollAppId(ctx sdk.Context, rollAppId, alias string) error {
	if !k.IsRollAppId(ctx, rollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "not a RollApp chain-id: %s", rollAppId)
	}

	if alias == "" {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias can not be empty")
	}

	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias: %s", alias)
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
func (k Keeper) GetAliasesOfRollAppId(ctx sdk.Context, rollAppId string) []string {
	store := ctx.KVStore(k.storeKey)
	keyR2A := dymnstypes.RollAppIdToAliasesKey(rollAppId)

	var multipleAliases dymnstypes.MultipleAliases
	if bz := store.Get(keyR2A); bz != nil {
		k.cdc.MustUnmarshal(bz, &multipleAliases)
	}

	return multipleAliases.Aliases
}

// RemoveAliasFromRollAppId removes the linking of an existing alias from a RollApp.
func (k Keeper) RemoveAliasFromRollAppId(ctx sdk.Context, rollAppId, alias string) error {
	if !k.IsRollAppId(ctx, rollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "not a RollApp chain-id: %s", rollAppId)
	}

	if alias == "" {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias can not be empty")
	}

	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias: %s", alias)
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

	var newMultipleAliases dymnstypes.MultipleAliases
	for _, a := range multipleAliases.Aliases {
		if a != alias {
			newMultipleAliases.Aliases = append(newMultipleAliases.Aliases, a)
		}
	}
	if len(newMultipleAliases.Aliases) == len(multipleAliases.Aliases) {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias not found: %s", alias)
	}

	if len(newMultipleAliases.Aliases) == 0 {
		store.Delete(keyR2A)
	} else {
		store.Set(keyR2A, k.cdc.MustMarshal(&newMultipleAliases))
	}
	store.Delete(keyA2R)

	return nil
}

// MoveAliasToRollAppId moves the linking of an existing alias from a RollApp to another RollApp.
func (k Keeper) MoveAliasToRollAppId(ctx sdk.Context, srcRollAppId, alias, dstRollAppId string) error {
	if !k.IsRollAppId(ctx, srcRollAppId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "source RollApp does not exists: %s", srcRollAppId)
	}

	if !dymnsutils.IsValidAlias(alias) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid alias: %s", alias)
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

// IsAliasPresentsInParamsAsAliasOrChainId returns true if the alias presents in the params.
// Extra check if it collision with chain-ids there.
func (k Keeper) IsAliasPresentsInParamsAsAliasOrChainId(ctx sdk.Context, alias string) bool {
	params := k.GetParams(ctx)

	for _, aliasesOfChainId := range params.Chains.AliasesOfChainIds {
		if alias == aliasesOfChainId.ChainId {
			return true
		}

		for _, a := range aliasesOfChainId.Aliases {
			if alias == a {
				return true
			}
		}
	}

	return false
}
