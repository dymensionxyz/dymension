package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func RollappKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	app, _ := apptesting.SetupTestingApp()
	return app.RollappKeeper, app.NewContext(false)
}
