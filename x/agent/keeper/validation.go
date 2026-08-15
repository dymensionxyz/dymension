package keeper

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func (k Keeper) GetValidationRequest(ctx sdk.Context, hash []byte) (types.ValidationRequest, bool) {
	v, err := k.validationRequests.Get(ctx, hash)
	return v, err == nil
}

func (k Keeper) SetValidationRequest(ctx sdk.Context, v types.ValidationRequest) error {
	return k.validationRequests.Set(ctx, v.RequestHash, v)
}

func (k Keeper) GetValidationResponse(ctx sdk.Context, hash []byte, seq uint64) (types.ValidationResponse, bool) {
	v, err := k.validationResponses.Get(ctx, collections.Join(hash, seq))
	return v, err == nil
}
