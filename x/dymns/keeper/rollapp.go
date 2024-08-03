package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// TODO DymNS: remove this mock
type mockRollAppData struct {
	alias  string
	bech32 string
}

var mockRollAppsData = map[string]mockRollAppData{
	"nim_1122-1": {
		alias:  "nim",
		bech32: "nim",
	},
	"mande_18071918-1": {
		alias:  "mande",
		bech32: "mande",
	},
}

// end of mock

// IsRollAppId checks if the chain-id is a RollApp-Id.
func (k Keeper) IsRollAppId(ctx sdk.Context, chainId string) bool {
	_, found := k.rollappKeeper.GetRollapp(ctx, chainId)

	if !found {
		_, found = mockRollAppsData[chainId]
	}

	return found
}

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
	if !k.IsRollAppId(ctx, chainId) {
		return
	}

	defer func() {
		found = alias != ""
	}()

	store := ctx.KVStore(k.storeKey)
	key := dymnstypes.RollAppIdToAliasKey(chainId)
	bz := store.Get(key)
	if bz != nil {
		alias = string(bz)
		return
	}

	if data, ok := mockRollAppsData[chainId]; ok {
		alias = data.alias
		return
	}

	return
}

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
	keyR2A := dymnstypes.RollAppIdToAliasKey(rollAppId)
	keyA2R := dymnstypes.AliasToRollAppIdRvlKey(alias)

	store.Set(keyR2A, []byte(alias))
	store.Set(keyA2R, []byte(rollAppId))

	return nil
}

// GetRollAppBech32Prefix returns the Bech32 prefix of the RollApp by the chain-id.
func (k Keeper) GetRollAppBech32Prefix(ctx sdk.Context, chainId string) (bech32Prefix string, found bool) {
	rollApp, found := k.rollappKeeper.GetRollapp(ctx, chainId)
	if found && len(rollApp.Bech32Prefix) > 0 {
		return rollApp.Bech32Prefix, true
	}

	if data, found := mockRollAppsData[chainId]; found {
		return data.bech32, len(data.bech32) > 0
	}

	return "", false
}
