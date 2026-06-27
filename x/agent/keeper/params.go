package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.params.Get(ctx)
}

func (k Keeper) AgentRegistrationFee(ctx sdk.Context) (sdk.Coin, error) {
	p, err := k.GetParams(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}
	return p.AgentRegistrationFee, nil
}
