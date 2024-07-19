package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// SetParam sets a specific sponsorship module's parameter with the provided parameter.
func (k Keeper) SetParam(ctx sdk.Context, key []byte, value interface{}) {
	k.paramSpace.Set(ctx, key, value)
}

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SaveDistribution(ctx sdk.Context, d types.Distribution) error {
	store := ctx.KVStore(k.storeKey)
	key := types.DistributionKey()

	value, err := k.cdc.Marshal(&d)
	if err != nil {
		return fmt.Errorf("can't marshal value: %s", err.Error())
	}
	store.Set(key, value)

	return nil
}

func (k Keeper) GetDistribution(ctx sdk.Context) (types.Distribution, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.DistributionKey())
	if b == nil {
		return types.Distribution{}, sdkerrors.ErrNotFound
	}

	var v types.Distribution
	err := k.cdc.Unmarshal(b, &v)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("can't unmarshal value: %s", err.Error())
	}

	return v, nil
}

func (k Keeper) SaveVotingPower(ctx sdk.Context, voterAddr, validatorAddr string, power math.Int) error {
	store := ctx.KVStore(k.storeKey)
	key := types.VotingPowerKey(voterAddr, validatorAddr)

	value, err := k.cdc.Marshal(&sdk.IntProto{Int: power})
	if err != nil {
		return fmt.Errorf("can't marshal value: %s", err.Error())
	}
	store.Set(key, value)

	return nil
}

func (k Keeper) GetVotingPower(ctx sdk.Context, voterAddr, validatorAddr string) (math.Int, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.VoteKey(voterAddr))
	if b == nil {
		return math.ZeroInt(), sdkerrors.ErrNotFound
	}

	var v sdk.IntProto
	err := k.cdc.Unmarshal(b, &v)
	if err != nil {
		return math.ZeroInt(), fmt.Errorf("can't unmarshal value: %s", err.Error())
	}

	return v.Int, nil
}

func (k Keeper) SaveVote(ctx sdk.Context, voterAddr string, v types.Vote) error {
	store := ctx.KVStore(k.storeKey)
	key := types.VoteKey(voterAddr)

	value, err := k.cdc.Marshal(&v)
	if err != nil {
		return fmt.Errorf("can't marshal value: %s", err.Error())
	}
	store.Set(key, value)

	return nil
}

func (k Keeper) GetVote(ctx sdk.Context, voterAddr string) (types.Vote, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.VoteKey(voterAddr))
	if b == nil {
		return types.Vote{}, sdkerrors.ErrNotFound
	}

	var v types.Vote
	err := k.cdc.Unmarshal(b, &v)
	if err != nil {
		return types.Vote{}, fmt.Errorf("can't unmarshal value: %s", err.Error())
	}

	return v, nil
}
