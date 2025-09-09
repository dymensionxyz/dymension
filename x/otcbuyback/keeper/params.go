package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

func (k Keeper) IsAcceptedDenom(ctx sdk.Context, denom string) bool {
	for _, t := range k.MustGetParams(ctx).AcceptedTokens {
		if t.Token == denom {
			return true
		}
	}
	return false
}

func (k Keeper) MustGetParams(ctx sdk.Context) types.Params {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	return params
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return types.Params{}, err
	}
	return params, nil
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	for _, token := range params.AcceptedTokens {
		denoms, err := k.ammKeeper.GetPoolDenoms(ctx, token.PoolId)
		if err != nil {
			return err
		}

		if len(denoms) != 2 {
			return fmt.Errorf("pool must have two denoms")
		}
		if (denoms[0] != k.baseDenom && denoms[1] != token.Token) ||
			(denoms[1] != k.baseDenom && denoms[0] != token.Token) {
			return fmt.Errorf("pool must have the token denom and the base denom, got %s and %s", denoms[0], denoms[1])
		}
	}

	err := k.params.Set(ctx, params)
	if err != nil {
		return err
	}

	return nil
}
