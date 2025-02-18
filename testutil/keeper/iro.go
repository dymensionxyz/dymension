package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
)

func IROKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	app := apptesting.Setup(t)
	return app.IROKeeper, app.NewContext(false)
}
