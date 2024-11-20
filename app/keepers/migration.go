package keepers

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (a *AppKeepers) MigrateModuleAccountPerms(ctx sdk.Context) {
	for moduleName, perms := range maccPerms {
		accI := a.AccountKeeper.GetModuleAccount(ctx, moduleName)
		if accI == nil {
			panic(fmt.Sprintf("module account not been set: %s", moduleName))
		}
		acc := accI.(*authtypes.ModuleAccount)
		acc.Permissions = perms
		a.AccountKeeper.SetModuleAccount(ctx, acc)
	}
}
