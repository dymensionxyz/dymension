package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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

// UpdateDistribution updates the distribution by applying the provided function to the current distribution.
// It retrieves the current distribution from the state, applies the update function to it, saves the updated distribution
// back to the state, and returns the updated distribution. If any error occurs during these steps, it returns an error.
func (k Keeper) UpdateDistribution(ctx sdk.Context, fn func(types.Distribution) types.Distribution) (types.Distribution, error) {
	// Get the current plan from the state
	current, err := k.GetDistribution(ctx)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to get distribution: %w", err)
	}

	// Apply the update
	result := fn(current)

	// Save the updated distribution
	err = k.SaveDistribution(ctx, result)
	if err != nil {
		return types.Distribution{}, fmt.Errorf("failed to save distribution: %w", err)
	}

	// Return the updated distribution
	return result, nil
}

func (k Keeper) SaveDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress, power math.Int) error {
	store := ctx.KVStore(k.storeKey)
	key := types.DelegatorValidatorPowerKey(voterAddr, valAddr)

	value, err := k.cdc.Marshal(&sdk.IntProto{Int: power})
	if err != nil {
		return fmt.Errorf("can't marshal value: %s", err.Error())
	}
	store.Set(key, value)

	return nil
}

func (k Keeper) GetDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress) (math.Int, error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.DelegatorValidatorPowerKey(voterAddr, valAddr))
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

func (k Keeper) HasDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.DelegatorValidatorPowerKey(voterAddr, valAddr))
}

func (k Keeper) IterateDelegatorValidatorPower(
	ctx sdk.Context,
	voterAddr sdk.AccAddress,
	fn func(valAddr sdk.ValAddress, power math.Int) (stop bool, err error),
) error {
	store := ctx.KVStore(k.storeKey)
	iterKey := types.AllDelegatorValidatorPowersKey(voterAddr)
	iterator := store.Iterator(iterKey, storetypes.PrefixEndBytes(iterKey))
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var power sdk.IntProto
		err := k.cdc.Unmarshal(iterator.Value(), &power)
		if err != nil {
			return fmt.Errorf("can't unmarshal value: %s", err.Error())
		}

		validator := iterator.Key()

		stop, err := fn(validator, power.Int)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	return nil
}

func (k Keeper) DeleteDelegatorValidatorPower(ctx sdk.Context, voterAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.DelegatorValidatorPowerKey(voterAddr, valAddr))
}

func (k Keeper) DeleteDelegatorPower(ctx sdk.Context, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	iterKey := types.AllDelegatorValidatorPowersKey(voterAddr)
	iterator := store.Iterator(iterKey, storetypes.PrefixEndBytes(iterKey))
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) SaveVote(ctx sdk.Context, voterAddr sdk.AccAddress, v types.Vote) error {
	store := ctx.KVStore(k.storeKey)
	key := types.VoteKey(voterAddr)

	value, err := k.cdc.Marshal(&v)
	if err != nil {
		return fmt.Errorf("can't marshal value: %s", err.Error())
	}
	store.Set(key, value)

	return nil
}

func (k Keeper) GetVote(ctx sdk.Context, voterAddr sdk.AccAddress) (types.Vote, error) {
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

func (k Keeper) Voted(ctx sdk.Context, voterAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.VoteKey(voterAddr))
}

func (k Keeper) DeleteVote(ctx sdk.Context, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.VoteKey(voterAddr))
}

func (k Keeper) IterateVotes(
	ctx sdk.Context,
	fn func(voter sdk.AccAddress, vote types.Vote) (stop bool, err error),
) error {
	store := ctx.KVStore(k.storeKey)
	voteByte := []byte{types.VoteByte}
	iterator := store.Iterator(voteByte, storetypes.PrefixEndBytes(voteByte))
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		err := k.cdc.Unmarshal(iterator.Value(), &vote)
		if err != nil {
			return fmt.Errorf("can't unmarshal value: %s", err.Error())
		}

		voter := iterator.Key()

		stop, err := fn(voter, vote)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}

	return nil
}

func (k Keeper) SaveInactiveVoter(ctx sdk.Context, voterAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.InactiveVoterKey()
	store.Set(key, voterAddr.Bytes())
}

func (k Keeper) DequeueInactiveVoters(ctx sdk.Context) []sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	voteByte := []byte{types.VoteByte}
	iterator := store.Iterator(voteByte, storetypes.PrefixEndBytes(voteByte))
	defer iterator.Close() // nolint: errcheck

	var voters []sdk.AccAddress
	for ; iterator.Valid(); iterator.Next() {
		voters = append(voters, iterator.Key())
		store.Delete(iterator.Key())
	}

	return voters
}
