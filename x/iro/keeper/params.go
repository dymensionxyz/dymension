package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.Params{} //FIXME: implement GetParams
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	//FIXME: implement SetParams
}
