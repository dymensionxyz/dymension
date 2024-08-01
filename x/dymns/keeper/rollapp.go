package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	// TODO DymNS: implement Get RollApp-Id By Alias

	for rollAppId, data := range mockRollAppsData {
		if data.alias == alias {
			return rollAppId, true
		}
	}

	return "", false
}

// GetAliasByRollAppId returns the alias by the RollApp-Id.
func (k Keeper) GetAliasByRollAppId(ctx sdk.Context, chainId string) (alias string, found bool) {
	if data, found := mockRollAppsData[chainId]; found {
		return data.alias, true
	} else {
		return "", false
	}

	/*
		_, exists := k.rollappKeeper.GetRollapp(ctx, chainId)
		if !exists {
			return "", false
		}

		// TODO DymNS: implement Get Alias by RollApp-Id
		return "", false
	*/
}

// GetRollAppBech32Prefix returns the Bech32 prefix of the RollApp by the chain-id.
func (k Keeper) GetRollAppBech32Prefix(ctx sdk.Context, chainId string) (bech32Prefix string, found bool) {
	if data, found := mockRollAppsData[chainId]; found {
		return data.bech32, true
	} else {
		return "", false
	}

	/*
		// TODO DymNS: implement Get RollApp Bech32 Prefix
		return "", false
	*/
}
