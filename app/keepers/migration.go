package keepers

import (
	"fmt"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"golang.org/x/exp/maps"
)

func (a *AppKeepers) MigrateModuleAccountPerms(ctx sdk.Context) {
	keys := maps.Keys(maccPerms)
	slices.Sort(keys)
	for _, moduleName := range keys {
		perms := maccPerms[moduleName]

		accI := a.AccountKeeper.GetModuleAccount(ctx, moduleName)
		if accI == nil {
			panic(fmt.Sprintf("module account not been set: %s", moduleName))
		}
		//nolint:all - we want to panic here
		acc := accI.(*authtypes.ModuleAccount)
		acc.Permissions = perms
		a.AccountKeeper.SetModuleAccount(ctx, acc)
	}
}
