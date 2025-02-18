package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
)

func EIBCKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	app := apptesting.Setup(t)
	return &app.EIBCKeeper, app.NewContext(false)
}
